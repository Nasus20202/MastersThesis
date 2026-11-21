---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-route-misconfig/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-route-misconfig · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

A Gateway and an HTTPRoute are deployed, but HTTP requests through the Gateway do not reach the application. Restore end-to-end routing through host app.example.com and path /app.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > HTTPRoute

## Write commands

1. `kubectl patch httproute app-route -n default --type=json -p '[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]' 2>`

## Final report

Routing is restored. One misconfiguration was found and fixed.

## Root cause

The HTTPRoute's `parentRef` pointed at a Gateway name that doesn't exist:

```
kubectl describe httproute app-route   # before
Parent Refs: Name: app-gw        <-- wrong
```

Actual Gateway in `default` is `app-gateway`. Because the reference was dangling, the route never bound to the listener — `Gateway status.listeners[0].attachedRoutes` was **0** and every request through Traefik returned **404** (the route's `status.parents` was also stale: `observedGeneration: 1` while the object was at `generation: 2`). The Gateway, GatewayClass (`Accepted=True/Handled`), Service `app` (`endpoints 10.244.0.6:80`), and Deployment `app` were all healthy, and nginx already serves `/app`, so no other change was needed.

## Change made (smallest justified fix)

Patched only the parentRef name on the persistent owner of the route config:

```
kubectl patch httproute app-route -n default --type=json \
  -p '[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]'
```

Nothing else was touched: listener port 8000 (Traefik `web` entrypoint), hostname `app.example.com`, path `PathPrefix /app`, backend `app:80`, and the Service/Deployment were left as-is.

## Verification (fresh observations after the fix)

- Binding: `httproute app-route` status → `parentRef.name=app-gateway`, `Accepted=True/Accepted`, `ResolvedRefs=True/ResolvedRefs` at `observedGeneration: 3`; `gateway app-gateway` → `attachedRoutes: 1` (was 0).
- Traffic through the Gateway Service (`traefik.traefik.svc:80`, Host `app.example.com`), requested from a pod inside the cluster:
  - `GET /app` → **200** ×3, body `<title>Welcome to nginx!</title>`, `Server: nginx/1.31.5`
  - `GET /app/sub` (PathPrefix) → **200**
  - `GET /nope` → **404** (correct: outside the route match)
  - `Host: other.example.com` → **404** (hostname filtering intact)
- Before the fix the same request returned `HTTP/1.1 404 Not Found` / `404 page not found`, so the reported symptom is gone.
- Controllers match requested capacity: `app` desired=1 updated=1 ready=1 available=1; `traefik` desired=1 updated=1 ready=1 available=1. Gateway conditions: `Accepted=True`, `Programmed=True`.

Note: NodePort/LoadBalancer access (`30507`) couldn't be probed from this shell — it has no route to the node network (`Network is unreachable`) and the `traefik` Service is still `LoadBalancer <pending>` with no assigned address — so verification was done in-cluster through the Service, which exercises the full Service → Traefik → Gateway → HTTPRoute → `app` path.
