---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-14-18-13-727Z/gateway-httproute-basic/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-httproute-basic · deepseek-flash · run-2026-11-21-14-18-13-727Z · attempt 1

## Task

The Service app in the default namespace is healthy. Expose it with the Gateway API by creating a Gateway named app-gateway (gatewayClassName traefik) with an HTTP listener named http on port 8000, and an HTTPRoute named app-route that routes host app.example.com and path /app (PathPrefix) to Service app on port 80. Requests are served through the traefik Service on port 80.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > Gateway, Resource model > HTTPRoute

## Write commands

1. `write_file {"content": "apiVersion: gateway.networking.k8s.io/v1\nkind: Gateway\nmetadata:\n name: app-gateway\n namespace: default\nspec:\n gatewayClassName: traefik\n listeners:\n - name: http\n protocol: HTTP\n port: 8000\n allowedRoutes:\n namespaces:\n from: Same\n---\napiVersion: gateway.networking.k8s.io/v1\nkind: HTTPRoute\nmetadata:\n name: app-route\n namespace: default\nspec:\n parentRefs:\n - name: app-gateway\n sectionName: http\n hostnames:\n - app.example.com\n rules:\n - matches:\n - path:\n type: PathPrefix\n value: /app\n backendRefs:\n - name: app\n port: 80\n", "path": "gateway.yaml"}`
2. `kubectl apply -f gateway.yaml`
3. `kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c '`
4. `kubectl run curl-test2 --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'curl -sS -o /dev/null -w "wrong-host:%{http_code}\n" -H "Host: other.example.com" http://traefik.traefik.svc.cluster.local/app' 2>`
5. `kubectl run curl-test3 --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'echo "-- /app plain --"`

## Final report

Done. The Gateway API exposure is in place and verified end-to-end.

## What I observed first

- `Service app` in `default` existed and healthy: ClusterIP `10.96.42.116`, port `80` → targetPort `80`, selecting pod `app-bb695fb5-sczqb` (`nginx:1.31.5`, 1/1 Running).
- `GatewayClass traefik` already existed with `ACCEPTED=True`, controller `traefik.io/gateway-controller`; Gateway API v1 CRDs present.
- The traefik install already had the Gateway provider enabled and `web` entrypoint on `8000`, exposed by `Service traefik.traefik` on port `80` — so no change to traefik was needed; the listener must use port 8000 to line up with service port 80.

## What I created (only new objects; nothing else modified)

- `Gateway app-gateway` in `default`: `gatewayClassName: traefik`, one HTTP listener `http` on port `8000` (`allowedRoutes` same-namespace).
- `HTTPRoute app-route` in `default`: parentRef `app-gateway` (sectionName `http`), hostname `app.example.com`, match `PathPrefix /app`, backendRef `Service app` port `80`.

## Verified outcome (fresh observations)

- Gateway status: `Accepted=True (Gateway successfully scheduled)`, `Programmed=True (Gateway successfully programmed)`; listener `http` `Accepted/ResolvedRefs/Programmed=True`, `attachedRoutes: 1`.
- HTTPRoute status: parent `app-gateway/http` → `Accepted=True`, `ResolvedRefs=True` from `traefik.io/gateway-controller`.
- Live request through the traefik Service on port 80 (`http://traefik.traefik.svc.cluster.local`):
  - `Host: app.example.com` + `/app` → **HTTP 200**, `Server: nginx/1.31.5`, nginx welcome body (so it reached the `app` backend, not a generic response).
  - `Host: app.example.com` + `/app/deep/path` → **HTTP 200** (confirms PathPrefix, not exact match).
  - `Host: other.example.com` + `/app` → **HTTP 404** (confirms host scoping is enforced, i.e. the route is actually what served the 200s).
- Test pods cleaned up; only `app-bb695fb5-sczqb` remains in `default`.
