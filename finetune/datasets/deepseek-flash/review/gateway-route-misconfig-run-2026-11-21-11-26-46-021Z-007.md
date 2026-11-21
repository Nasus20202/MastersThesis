---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-route-misconfig/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-route-misconfig · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A Gateway and an HTTPRoute are deployed, but HTTP requests through the Gateway do not reach the application. Restore end-to-end routing through host app.example.com and path /app.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > HTTPRoute

## Write commands

1. `kubectl run probe --image=curlimages/curl:8.10.1 --restart=Never --command -- sleep 3600 2>`
2. `kubectl patch httproute -n default app-route --type=json -p '[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]' 2>`
3. `kubectl delete pod -n default probe --wait=true 2>`

## Final report

## Summary

**Diagnosis (evidence):**

- Gateway `default/app-gateway` (gatewayClassName `traefik`) was `Accepted=True`, `Programmed=True`, but its `http` listener reported **`attachedRoutes: 0`**.
- HTTPRoute `default/app-route` had `spec.parentRefs[0].name: app-gw`, a Gateway that does not exist — the actual Gateway is named `app-gateway`. Its stale status (observedGeneration 1 vs. metadata generation 2) still showed a parent `app-gateway`, which was misleading.
- Requests through the Traefik service with `Host: app.example.com` and path `/app` returned **404** (no router bound), while direct calls to `service/app` returned **200**, and the app's nginx ConfigMap serves `/app` via `try_files /index.html`.

**Change made:** patched only the HTTPRoute's parentRef name (`app-gw` → `app-gateway`); hostname `app.example.com`, path match `PathPrefix: /app`, and backendRef `app:80` were left untouched.

**Verification (fresh observations):**

- Gateway listener now reports `attachedRoutes=1`; Gateway conditions `Accepted=True/Accepted`, `Programmed=True/Programmed`.
- HTTPRoute parent `app-gateway` conditions: `Accepted=True/Accepted (observedGeneration=3)`, `ResolvedRefs=True/ResolvedRefs (observedGeneration=3)` — matching its generation.
- Through the Gateway (`http://10.96.84.213` / `http://traefik.traefik.svc.cluster.local`, Host `app.example.com`):
  - `GET /app` → **200** (repeated 3/3 times)
  - `GET /app/sub` → **200**
  - Wrong host `other.example.com` `/app` → 404 (host scoping intact)
- The app pod's nginx access log records the requests arriving from the Traefik pod IP (`10.244.0.3 - - "GET /app HTTP/1.1" 200`), proving traffic traverses the Gateway to the application.
- Temporary `probe` pod removed; only the original `app` pod remains Running/Ready.
