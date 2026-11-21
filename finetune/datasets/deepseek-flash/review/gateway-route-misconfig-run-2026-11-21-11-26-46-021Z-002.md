---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-route-misconfig/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-route-misconfig · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A Gateway and an HTTPRoute are deployed, but HTTP requests through the Gateway do not reach the application. Restore end-to-end routing through host app.example.com and path /app.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > HTTPRoute

## Write commands

1. `kubectl patch httproute app-route -n default --type=json -p '[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]'`

## Final report

Done — routing is restored.

## Diagnosis

| Evidence                                                                                    | Finding                                                                                                                  |
| ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `Gateway app-gateway` status: `listeners[0] (http, port 8000) attachedRoutes: 0`            | The Gateway accepted and programmed fine, but no route was bound to it.                                                  |
| `HTTPRoute app-route.spec.parentRefs[0].name = app-gw` (generation 2)                       | The route pointed at a non-existent Gateway; the real Gateway is `app-gateway`.                                          |
| `HTTPRoute.status.parents[0]` showed `parentRef.name: app-gateway`, `observedGeneration: 1` | Stale status from the pre-edit generation (last-applied-config had the correct name), which is why it _looked_ Accepted. |
| Request to Traefik entrypoint (`:8000`) with `Host: app.example.com`, path `/app`           | `404 page not found` — Traefik had no router, so the request never reached the app Service/Pods.                         |

The Gateway listener port 8000 correctly maps to Traefik's `web` entrypoint (`--entryPoints.web.address=:8000`), the Service `app` had a ready endpoint (`10.244.0.6:80`), and the Deployment was `1/1` ready — so the only broken link was route→Gateway attachment.

## Change (single, minimal)

Patched only the owning resource's parent reference, leaving hostnames, path match, and backendRef untouched:

```
kubectl patch httproute app-route -n default --type=json \
  -p '[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]'
```

## Verification (fresh observations)

- Gateway: `attachedRoutes: 1` on listener `http` (was 0); conditions `Accepted=True`, `Programmed=True`.
- HTTPRoute: `parentRef.name: app-gateway`, `Accepted=True(Accepted)`, `ResolvedRefs=True`.
- End-to-end through Traefik **Service** (`traefik:80` → targetPort `web`/8000), host `app.example.com`, path `/app`:
  - `HTTP/1.1 200 OK`, `Content-Length: 896`, `Server: nginx/1.31.5`, body `<title>Welcome to nginx!</title>` — i.e. served by the app's own pod image, not Traefik's 404. Repeat request also `200`.
  - Negative checks confirm the routing is still scoped as configured: wrong host `nope.example.com` → `404`, wrong path `/other` → `404`.
- Deployment `app` unchanged and healthy: `desired=1 ready=1 available=1 updated=1`; pod `1/1 Running`.
- Traefik logs show no route/backend errors. Temporary `port-forward` helpers were stopped; no transient object was used as the fix.
