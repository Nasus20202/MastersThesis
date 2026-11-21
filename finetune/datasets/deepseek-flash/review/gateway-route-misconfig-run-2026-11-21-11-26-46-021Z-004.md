---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-route-misconfig/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-route-misconfig · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A Gateway and an HTTPRoute are deployed, but HTTP requests through the Gateway do not reach the application. Restore end-to-end routing through host app.example.com and path /app.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > HTTPRoute

## Write commands

1. `kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never --command -- curl -sS -o /dev/null -w "%{http_code}\n" -H "Host: app.example.com" http://traefik.traefik.svc.cluster.local/app 2>`
2. `kubectl patch httproute app-route -n default --type=json -p '[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]'`
3. `kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never --command -- sh -c 'curl -sS -o /dev/null -w "svc-app-80:%{http_code}\n" http://app.default.svc.cluster.local/app' 2>`
4. `kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never --command -- sh -c 'curl -sS -o /dev/null -w "gw:%{http_code}\n" -H "Host: app.example.com" http://traefik.traefik.svc.cluster.local:80/app`
5. `kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never --command -- sh -c 'curl -sS -H "Host: app.example.com" http://traefik.traefik.svc.cluster.local/app`

## Final report

Routing is restored and verified. Here's what I found and did:

## Diagnosis

**Observed symptom:** requests to `Host: app.example.com`, path `/app` through the Traefik Gateway returned **404** (from Traefik's default handler), i.e. no router matched.

**Evidence:**

- `Gateway app-gateway` (class `traefik`, listener `http` on port 8000) was `Accepted=True`, `Programmed=True`, but `status.listeners[0].attachedRoutes` was **0** — the route was never attached.
- `HTTPRoute app-route` had `spec.parentRefs[0].name: app-gw`, while the actual Gateway is named **`app-gateway`**. The parentRef is a name-based reference, so it pointed at a non-existent Gateway. (The stale `Accepted`/`ResolvedRefs` conditions in the route status claimed `parentRef.name: app-gateway`, which masked the misconfiguration.)
- Backend was healthy and not the problem: Service `app` (port 80) → Endpoints `10.244.0.6:80`; a direct in-cluster request to `http://app.default.svc.cluster.local/app` returned **200**; Deployment `app` had `replicas=1 / readyReplicas=1 / availableReplicas=1`; the `app-nginx` ConfigMap already serves `/app`.
- The Gateway listener port 8000 correctly maps to Traefik's `web` entrypoint (`--entryPoints.web.address=:8000/tcp`, Service `traefik:80 → targetPort web`), so no gateway change was needed.

## Change (single, minimal)

Patched only the broken reference in the existing HTTPRoute:

```
kubectl patch httproute app-route -n default --type=json \
  -p '[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]'
```

No other fields, objects, or workloads were touched — Gateway listener port 8000, route hostname `app.example.com`, path `PathPrefix /app`, and backend `app:80` were all left as-is, and the app Pod was not restarted.

## Verification (fresh observations)

- `Gateway app-gateway` listener `http`: `attachedRoutes=1`, `Programmed=True` (was 0).
- `HTTPRoute app-route`: `Accepted=True`, `ResolvedRefs=True`, `metadata.generation=3` == `observedGeneration=3`.
- End-to-end request through the Gateway service:
  - `curl -H "Host: app.example.com" http://traefik.traefik.svc.cluster.local/app` → **200**, body = nginx welcome page (application content, not Traefik's fallback).
  - `curl -H "Host: other.example.com" .../app` → **404** (expected: only `app.example.com` is matched).
  - `curl -H "Host: app.example.com" .../` → **404** (expected: route only exposes `/app`).

Both the requested outcome (host `app.example.com` + path `/app` reaching the app end-to-end) and the original constraints (same Gateway, listener, and backend configuration) hold.
