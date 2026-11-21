---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/ingress-path-routing/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# ingress-path-routing · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

The Service app in the default namespace is healthy. Create an Ingress named app that routes HTTP requests for host app.example.com and path /app (pathType Prefix) to Service app on port 80.

Documented in:

- content/en/docs/concepts/services-networking/ingress.md: The Ingress resource > Ingress rules, The Ingress resource > Path types, Types of Ingress > Ingress backed by a single Service

## Write commands

1. `kubectl run ctest --rm -i --restart=Never --image=curlimages/curl:latest --command -- curl -s -o /dev/null -w '%{http_code}\n' http://app/`
2. `kubectl apply -f -`
3. `kubectl run ctest --rm -i --restart=Never --image=curlimages/curl:latest --command -- sh -c '`
4. `kubectl run ctest --rm -i --restart=Never --image=curlimages/curl:latest --command -- sh -c '`

## Final report

## What I did

The Service `app` was already healthy (ClusterIP `10.96.8.71:80`, one ready endpoint `10.244.0.6:80`), so I left it untouched and only added the requested Ingress. The cluster has a working Traefik controller (`IngressClass traefik` marked `(default)`, pod `traefik-6d9c545874-tsbb2` 1/1 Running), so I set that class explicitly rather than relying on the default annotation behavior.

Created `networking.k8s.io/v1` Ingress `app` in `default` with exactly the requested routing, verified via jsonpath:

```
app host=app.example.com path=/app pathType=Prefix svc=app:80 class=traefik
```

## Verification (real requests through the controller)

Issued from an in-cluster curl pod against the Traefik Service `traefik.traefik.svc.cluster.local:80`:

| Request                            | Result                                          |
| ---------------------------------- | ----------------------------------------------- |
| `Host: app.example.com` → `/app`   | **200**, nginx welcome page (the app's content) |
| `Host: app.example.com` → `/app/`  | **200**                                         |
| `Host: other.example.com` → `/app` | 404 (host rule scoped correctly)                |
| `Host: app.example.com` → `/other` | 404 (path rule scoped correctly)                |
| no Host header → `/app`            | 404                                             |

Backend hit confirmed independently by the app pod's access log, which shows the proxied requests arriving from the Traefik pod IP:

```
10.244.0.4 - - "GET /app HTTP/1.1" 200 896 ... "10.244.0.8"
10.244.0.4 - - "GET /app/ HTTP/1.1" 200 896 ... "10.244.0.8"
```

`kubectl describe ingress app` resolves the backend to `app:80 (10.244.0.6:80)`, matching the Service's endpoint.

## Note (not a fault I introduced)

The Ingress `status.loadBalancer` is empty because the Traefik Service is type LoadBalancer and shows `EXTERNAL-IP <pending>` in this sandbox (no external load-balancer implementation). Routing itself is functional as proven above through the controller Service (`10.96.60.213`, NodePort `31686`); I did not alter the controller or Service type, since the task only asked for the Ingress and the constraints on that were unchanged.
