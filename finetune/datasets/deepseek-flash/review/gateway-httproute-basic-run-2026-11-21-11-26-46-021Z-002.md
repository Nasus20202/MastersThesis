---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-httproute-basic/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-httproute-basic · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

The Service app in the default namespace is healthy. Expose it with the Gateway API by creating a Gateway named app-gateway (gatewayClassName traefik) with an HTTP listener named http on port 8000, and an HTTPRoute named app-route that routes host app.example.com and path /app (PathPrefix) to Service app on port 80. Requests are served through the traefik Service on port 80.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > Gateway, Resource model > HTTPRoute

## Write commands

1. `kubectl apply -f -`
2. `kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'echo "== via traefik svc :80, host header, /app =="`
3. `kubectl run curl-test2 --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'echo "== direct app svc =="`
4. `kubectl run curl-test3 --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'for p in /app /app/ /app/sub /other /`
5. `kubectl run curl-test4 --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'echo "wrong-host /app -> $(curl -s -o /dev/null -w "%{http_code}" -H "Host: nope.example.com" http://traefik.traefik.svc.cluster.local/app)"' 2>`

## Final report

## What I did

Both resources were absent (`kubectl get gateway,httproute -A` → none). I created them in `default` with only the requested fields:

**Gateway `app-gateway`** — `gatewayClassName: traefik`, single listener `http` (protocol HTTP, port 8000). Port 8000 matches the Traefik `web` entrypoint, which the `traefik` Service exposes as port 80.

**HTTPRoute `app-route`** — `parentRefs: [app-gateway]`, `hostnames: [app.example.com]`, one rule matching `path: {type: PathPrefix, value: /app}` with `backendRefs: [{name: app, port: 80}]`.

## Verification (fresh observations)

Controller acceptance:

- Gateway status: `Accepted=True/Accepted`, `Programmed=True/Programmed`; listener `http` shows `attachedRoutes: 1`, listener conditions `Accepted=True`, `ResolvedRefs=True`, `Programmed=True`.
- HTTPRoute status parent `traefik.io/gateway-controller`: `Accepted=True/Accepted`, `ResolvedRefs=True/ResolvedRefs`.
- `kubectl get gateway app-gateway` → `PROGRAMMED True`.

Real traffic through the traefik Service on **port 80** (`http://traefik.traefik.svc.cluster.local:80`), from an in-cluster pod:

| Request                             | Result                                          |
| ----------------------------------- | ----------------------------------------------- |
| `Host: app.example.com` `GET /app`  | **200** (nginx welcome page from Service `app`) |
| `GET /app/`                         | 200                                             |
| `GET /app/sub`                      | 200 (PathPrefix honored)                        |
| `GET /other`                        | 404 (path not matched)                          |
| `GET /`                             | 404 (path not matched)                          |
| `Host: nope.example.com` `GET /app` | 404 (host not matched)                          |

The 200 on `/app` proves the route reaches the healthy `app` Service, and the 404s for the wrong host/path prove the hostname and PathPrefix matching are actually enforced rather than a catch-all. No changes were made to the `app` Service, the Traefik Deployment/Service, or the GatewayClass; the temporary curl pods were deleted automatically (`kubectl get pods -n default` shows only `app-bb695fb5-nq5rr 1/1 Running`).

Note: the `traefik` LoadBalancer Service still has a pending external IP (no cloud LB in this sandbox), so the Gateway has no published `.status.addresses`; traffic was verified via the Service's port 80 ClusterIP path as specified.
