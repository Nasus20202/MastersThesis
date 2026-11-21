---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/ingress-path-routing/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# ingress-path-routing · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

The Service app in the default namespace is healthy. Create an Ingress named app that routes HTTP requests for host app.example.com and path /app (pathType Prefix) to Service app on port 80.

Documented in:

- content/en/docs/concepts/services-networking/ingress.md: The Ingress resource > Ingress rules, The Ingress resource > Path types, Types of Ingress > Ingress backed by a single Service

## Write commands

1. `kubectl apply -f /tmp/ingress.yaml`
2. `kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'curl -s -o /dev/null -w "%{http_code}\n" --max-time 5 http://app.default.svc.cluster.local/' 2>`
3. `kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c '`
4. `kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c '`

## Final report

**Done.** The Ingress was created and verified with live requests through the ingress controller.

**What exists now**

- Ingress `app` in `default`, class `traefik` (the cluster's default IngressClass, controller `traefik.io/ingress-controller`), single rule:
  - host `app.example.com`, path `/app`, pathType `Prefix` → Service `app` port `80`.
- Backend Service is healthy: `app` ClusterIP `10.96.51.117`, endpoint `10.244.0.6:80`, backing pod `app-bb695fb5-fm4jp` 1/1 Running.

**Verification (fresh requests sent through `traefik` Service 10.96.200.191:80, using the host header)**

| Request                                              | Result                                                     |
| ---------------------------------------------------- | ---------------------------------------------------------- |
| `app.example.com/app`                                | **200** — body is the app's nginx "Welcome to nginx!" page |
| `app.example.com/app/sub` (confirms Prefix matching) | **200**                                                    |
| `app.example.com/` (no matching rule)                | 404                                                        |
| `other.example.com/app` (host doesn't match)         | 404                                                        |

`kubectl describe ingress app` also shows the rule resolved to a live backend: `/app → app:80 (10.244.0.6:80)`.

Note: `status.loadBalancer` is empty because the `traefik` LoadBalancer Service has no external IP assigned in this kind cluster (`EXTERNAL-IP <pending>`), which is environmental; routing through the controller's service was confirmed working, so no unrelated configuration was changed.
