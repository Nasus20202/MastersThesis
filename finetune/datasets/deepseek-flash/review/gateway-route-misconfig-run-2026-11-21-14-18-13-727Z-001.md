---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-14-18-13-727Z/gateway-route-misconfig/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# gateway-route-misconfig · deepseek-flash · run-2026-11-21-14-18-13-727Z · attempt 1

## Task

A Gateway and an HTTPRoute are deployed, but HTTP requests through the Gateway do not reach the application. Restore end-to-end routing through host app.example.com and path /app.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > HTTPRoute

## Write commands

1. `kubectl patch httproute app-route -n default --type='json' -p='[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]' 2>`
2. `kubectl run tmp-curl --rm -i --restart=Never --image=curlimages/curl:latest --command -- curl -sS -m 5 -o /dev/null -w '%{http_code}\n' http://app.default.svc.cluster.local/app 2>`
3. `kubectl run gw-test --rm -i --restart=Never --image=curlimages/curl:latest --command -- sh -c 'echo "--- via traefik svc, correct host+path ---"`
4. `kubectl run gw-test2 --rm -i --restart=Never --image=curlimages/curl:latest --command -- sh -c 'echo "--- via NodePort 172.20.0.6:32567 ---"`

## Final report

## Root cause

The HTTPRoute's `spec.parentRefs[0].name` was **`app-gw`**, but the actual Gateway is named **`app-gateway`**. The route therefore never attached to the listener: the Gateway reported `attachedRoutes: 0`, so Traefik had no router for `app.example.com`.

## Fix

Patched the existing persistent HTTPRoute (no deletion/restart):

```
kubectl patch httproute app-route -n default --type=json \
  -p='[{"op":"replace","path":"/spec/parentRefs/0/name","value":"app-gateway"}]'
```

## Verification (fresh observations)

- Gateway listener: **`attachedRoutes: 1`** (was 0); `Accepted=True`, `Programmed=True`.
- HTTPRoute conditions: `Accepted=True/Accepted` and `ResolvedRefs=True/ResolvedRefs`, `observedGeneration=3` (matches spec gen 3), parentRef now `app-gateway`.
- Backend: Service `app` has endpoint `10.244.0.6:80`; Deployment `app` reports `replicas=1 ready=1 updated=1 available=1` (all match).
- End-to-end through the Gateway Service (`traefik.traefik.svc:80`, Host `app.example.com`, path `/app`): **HTTP 200**, body = nginx welcome page.
- End-to-end through the Gateway's external NodePort (`172.20.0.6:32567`, same Host/path): **HTTP 200**, `<title>Welcome to nginx!</title>`.
- Negative controls confirm route matching is in effect: wrong host `other.example.com` → **404**, wrong path `/other` → **404**.

No other settings (listener port 8000, backend service/port 80, path prefix `/app`, hostname) were altered.
