---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-httproute-basic/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-httproute-basic · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

The Service app in the default namespace is healthy. Expose it with the Gateway API by creating a Gateway named app-gateway (gatewayClassName traefik) with an HTTP listener named http on port 8000, and an HTTPRoute named app-route that routes host app.example.com and path /app (PathPrefix) to Service app on port 80. Requests are served through the traefik Service on port 80.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > Gateway, Resource model > HTTPRoute

## Write commands

1. `kubectl apply -f -`

## Final report

Done. The Service was exposed via the Gateway API with the exact requested configuration, and end-to-end traffic was verified.

## What I created (namespace `default`)

**Gateway `app-gateway`** — class `traefik`, listener `http` (HTTP, port `8000`). Port 8000 matches Traefik's `--entryPoints.web.address=:8000/tcp`, which is also the `targetPort: web` of the `traefik` Service's port 80.

**HTTPRoute `app-route`** — parentRef `app-gateway`/sectionName `http`, hostname `app.example.com`, single rule matching path `PathPrefix /app` with backendRef Service `app` port `80`.

No existing resources (Service `app`, Traefik Helm release/Service/Deployment, GatewayClass) were modified.

## Verification

Controller status (controller `traefik.io/gateway-controller`):

- Gateway: `Accepted=True` (Gateway successfully scheduled), `Programmed=True` (Gateway successfully programmed); listener `http`: `attachedRoutes: 1`, `Accepted/ResolvedRefs/Programmed` all `True`.
- HTTPRoute: parent `app-gateway`/`http` → `Accepted=True`, `ResolvedRefs=True`.

Live request through the **traefik Service on port 80** (port-forward to `svc/traefik` 80 → 8000, i.e. the Service's `web` port), with real HTTP status codes:

| Host              | Path     | Code                                              |
| ----------------- | -------- | ------------------------------------------------- |
| app.example.com   | /app     | **200** (nginx 1.31.5 welcome page body returned) |
| app.example.com   | /app/sub | **200**                                           |
| app.example.com   | /        | 404 (path not matched)                            |
| other.example.com | /app     | 404 (host not matched)                            |

The 200s confirm routing reaches Service `app` (endpoints `10.244.0.6:80`, pod `app-bb695fb5-g5r66` `1/1 Running`), and the 404s confirm the host and `PathPrefix /app` match conditions are actually being enforced rather than a catch-all.

Note: the Gateway shows no `ADDRESS` because Traefik reports its status address from `svc/traefik`, which is `type: LoadBalancer` with `EXTERNAL-IP <pending>` in this sandbox (no load-balancer provider). This does not affect routing, as proven by the successful requests above.
