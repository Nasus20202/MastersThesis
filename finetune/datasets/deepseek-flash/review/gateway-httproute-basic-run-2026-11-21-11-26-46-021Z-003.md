---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-httproute-basic/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-httproute-basic · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

The Service app in the default namespace is healthy. Expose it with the Gateway API by creating a Gateway named app-gateway (gatewayClassName traefik) with an HTTP listener named http on port 8000, and an HTTPRoute named app-route that routes host app.example.com and path /app (PathPrefix) to Service app on port 80. Requests are served through the traefik Service on port 80.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > Gateway, Resource model > HTTPRoute

## Write commands

1. `kubectl apply -f /tmp/gw.yaml`
2. `kubectl auth can-i create pods -n default`
3. `kubectl auth can-i create pods/portforward -n default`
4. `kubectl auth can-i create pods/exec -n default`
5. `kubectl run curl-test --image=curlimages/curl:8.10.1 --restart=Never -n default --command -- sh -c 'echo "== /app =="`
6. `kubectl run curl-test2 --image=curlimages/curl:8.10.1 --restart=Never -n default --command -- sh -c 'echo -n "wrong-host /app: "`
7. `kubectl delete pod curl-test curl-test2 -n default --wait=false`

## Final report

Done. Here's what I observed and verified.

## What I found

- `Service/app` (default ns, ClusterIP, port 80 → targetPort 80) backed by a healthy `Deployment/app` (1/1 ready, nginx pod `Running`); no Gateway/HTTPRoute existed yet (`gatewayclass/traefik` was already `Accepted=True`).
- The Traefik deployment runs its `web` entrypoint on `:8000`, and `Service/traefik` in ns `traefik` exposes **port 80 → targetPort `web` (8000)**, with endpoints `10.244.0.3:8000`. That's why the listener port must be `8000` (Traefik maps Gateway listener → matching entrypoint) while clients reach it via the service's port 80.

## Change made (smallest justified, new objects only)

Created in `default` (nothing pre-existing modified):

- `Gateway app-gateway`, `gatewayClassName: traefik`, listener `http`, protocol `HTTP`, port `8000`.
- `HTTPRoute app-route`, parentRef `app-gateway`, hostname `app.example.com`, single rule: `PathPrefix /app` → backendRef `app:80`.

## Verification (direct observations)

- Gateway status: `Accepted=True/Accepted`, `Programmed=True/Programmed`; listener `http` → `Accepted=True`, `ResolvedRefs=True`, `Programmed=True`, `attachedRoutes: 1`.
- HTTPRoute status: parent `app-gateway` → `Accepted=True/Accepted`, `ResolvedRefs=True/ResolvedRefs`.
- Live traffic **through `Service/traefik` on port 80** (from an in-cluster curl pod, now deleted):
  - `GET http://traefik.traefik.svc.cluster.local/app` with `Host: app.example.com` → **200**, body is the app's nginx page (`Welcome to nginx!`).
  - `GET /nope` with `Host: app.example.com` → **404** (path not matched, as expected).
  - `GET /app` with `Host: other.example.com` → **404**, and with no Host header → **404** (hostname restriction enforced).
- Constraints intact: `deployment.apps/app` still `1/1 READY, 1/1 UP-TO-DATE, 1/1 AVAILABLE`; `service/app` unchanged (port 80); Traefik running 1/1; no existing configuration altered.
