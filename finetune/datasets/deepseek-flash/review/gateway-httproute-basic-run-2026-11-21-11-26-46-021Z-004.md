---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/gateway-httproute-basic/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-httproute-basic · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

The Service app in the default namespace is healthy. Expose it with the Gateway API by creating a Gateway named app-gateway (gatewayClassName traefik) with an HTTP listener named http on port 8000, and an HTTPRoute named app-route that routes host app.example.com and path /app (PathPrefix) to Service app on port 80. Requests are served through the traefik Service on port 80.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > Gateway, Resource model > HTTPRoute

## Write commands

1. `kubectl apply -f /tmp/gw.yaml -f /tmp/route.yaml`
2. `kubectl run curl-test -n default --image=curlimages/curl:8.10.1 --restart=Never --command -- sleep 300 2>`
3. `kubectl delete pod curl-test -n default --wait=true --timeout=60s`

## Final report

Done. The app is exposed through the Traefik Gateway API, verified with live requests.

## What I discovered

- Gateway API CRDs present; GatewayClass `traefik` already `Accepted=True` (controller `traefik.io/gateway-controller`).
- Traefik deployment (v3.7.13) has `--providers.kubernetesgateway` enabled, entrypoint `web` on `:8000`, and `statusaddress.service` = `traefik/traefik`.
- The `traefik` Service exposes port **80 → targetPort `web` (8000)**, so the requested listener port 8000 maps onto Traefik's web entrypoint (a mismatched port would not bind).
- Backend `Service/app` port 80 → nginx pod on port 80; `app` Deployment 1/1 ready.

## Change applied (default namespace)

- **Gateway `app-gateway`** — `gatewayClassName: traefik`, one listener `name: http`, `protocol: HTTP`, `port: 8000`.
- **HTTPRoute `app-route`** — `parentRefs: app-gateway`, `hostnames: [app.example.com]`, rule match `PathPrefix /app`, `backendRefs: app:80`.

No other resources, Services, or Traefik settings were modified.

## Verified outcomes (fresh observations)

Controller acceptance:

- Gateway: `Accepted=True (Accepted)` and `Programmed=True (Programmed)`, listener `http` `attachedRoutes: 1`, `ResolvedRefs=True`.
- HTTPRoute: parent `app-gateway` `Accepted=True`, `ResolvedRefs=True`.

End-to-end requests through the `traefik` Service on port 80 (from an in-cluster curl pod):

- `Host: app.example.com`, path `/app` → **200**, nginx page body returned.
- `Host: app.example.com`, path `/app/sub/page` → **200** (confirms PathPrefix semantics).
- `Host: app.example.com`, path `/other` → **404** (route does not over-match).
- `Host: wrong.example.com`, path `/app` → **404** (host restriction enforced).
- Baseline direct `http://app.default.svc:80/app` → **200**, confirming the 200 through the gateway is the intended backend.

The transient `curl-test` pod used for probing was deleted; only `app-gateway` and `app-route` remain as the persistent change, and the `app` Deployment still reports 1/1 ready, 1 updated, 1 available.
