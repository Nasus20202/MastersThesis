---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/ingress-path-routing/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# ingress-path-routing · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The Service app in the default namespace is healthy. Create an Ingress named app that routes HTTP requests for host app.example.com and path /app (pathType Prefix) to Service app on port 80.

Documented in:

- content/en/docs/concepts/services-networking/ingress.md: The Ingress resource > Ingress rules, The Ingress resource > Path types, Types of Ingress > Ingress backed by a single Service

## Write commands

1. `kubectl apply -f -`
2. `kubectl run curl --rm -i --restart=Never --image=curlimages/curl:latest --command -- curl -s -m 10 -o /dev/null -w '%{http_code}\n' http://app.default.svc.cluster.local/ 2>`
3. `kubectl run nettest --image=curlimages/curl:latest --restart=Never --command -- sleep 600 >/dev/null 2>`
4. `kubectl delete pod nettest --wait=false >/dev/null 2>`

## Final report

Done. The Ingress was created and verified end-to-end.

**What I did**

- Inspected the cluster: Service `app` in `default` was healthy (`ClusterIP 10.96.224.211`, port 80 → endpoints `10.244.0.6:80`), and the cluster's Ingress controller is Traefik (`IngressClass traefik`, marked default).
- Created a single new `networking.k8s.io/v1` Ingress named `app` in `default` (the only change made; the Service and all other resources were left untouched):
  - `ingressClassName: traefik`
  - host `app.example.com`, path `/app`, `pathType: Prefix`
  - backend `service app` port `80`

**Verification (fresh requests through the Ingress, not just object status)**

| Request                                                                         | Result                                                                              |
| ------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `GET /app` with `Host: app.example.com` via Traefik ClusterIP `10.96.48.223:80` | HTTP **200**, nginx welcome page returned from the `app` pod                        |
| `GET /app` with `Host: app.example.com` via NodePort `172.20.0.9:32622`         | HTTP **200**, same nginx response                                                   |
| `GET /app` with `Host: other.example.com` (control)                             | HTTP **404** — no route matches, confirming the rule is scoped to `app.example.com` |

`kubectl describe ingress app` resolves the rule to backend `app:80 (10.244.0.6:80)`, and the Ingress spec contains exactly the requested host/path/pathType/backend. The temporary test pod was deleted.
