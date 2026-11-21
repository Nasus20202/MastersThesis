---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/ingress-path-routing/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# ingress-path-routing · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

The Service app in the default namespace is healthy. Create an Ingress named app that routes HTTP requests for host app.example.com and path /app (pathType Prefix) to Service app on port 80.

Documented in:

- content/en/docs/concepts/services-networking/ingress.md: The Ingress resource > Ingress rules, The Ingress resource > Path types, Types of Ingress > Ingress backed by a single Service

## Write commands

1. `kubectl apply -f -`
2. `kubectl run curl-test --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'curl -s -o /dev/null -w "svc-app direct: %{http_code}\n" http://app.default.svc.cluster.local/app`

## Final report

Done. The Ingress `app` was created and verified end-to-end.

**What I observed before the change**

- Service `app` in `default`: ClusterIP `10.96.171.210`, port `80` (name `http`) → targetPort `80`, selector `app=app`, endpoints `10.244.0.6:80` (pod `app-bb695fb5-tkbmc` 1/1 Running).
- No Ingress existed in `default`.
- Cluster has one IngressClass: `traefik`, marked `ingressclass.kubernetes.io/is-default-class: "true"`, controller `traefik.io/ingress-controller`; Traefik pod running in namespace `traefik`.

**Change applied** (single new resource, nothing else modified):

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: app
  namespace: default
spec:
  ingressClassName: traefik
  rules:
    - host: app.example.com
      http:
        paths:
          - path: /app
            pathType: Prefix
            backend:
              service:
                name: app
                port:
                  number: 80
```

**Verification (fresh observations through the controller, not just object status)**

- `kubectl describe ingress app`: `Ingress Class: traefik`, Rule `app.example.com` `/app  -> app:80 (10.244.0.6:80)` — the backend resolved to the real pod endpoint.
- Traffic tests through Traefik (port-forward to `svc/traefik` :80 → container 8000):
  - `Host: app.example.com` + `/app` → **200**, body = nginx welcome page (the `app` service response).
  - `Host: app.example.com` + `/app/x` → **200** (confirms `pathType: Prefix`, not Exact).
  - `Host: app.example.com` + `/other` → **404** (path rule is scoped to `/app`).
  - `Host: wrong.example.com` + `/app` → **404** (host rule is scoped to `app.example.com`).
- In-cluster check from a pod: `http://app.default.svc.cluster.local/app` → **200**, and `Host: app.example.com` → `http://traefik.traefik.svc.cluster.local/app` → **200**.
- Final state: `kubectl get ingress app` → `CLASS traefik`, `HOSTS app.example.com`, `PORTS 80`; Service `app` unchanged and still has endpoint `10.244.0.6` with 1/1 ready pod.

Constraints respected: only the requested Ingress was created in `default`; the existing Service, its ClusterIP/ports, the Traefik install, and the IngressClass configuration were left untouched. Note Traefik's Service has `EXTERNAL-IP <pending>` (kind has no cloud LB), which is why verification used the cluster/NodePort reachable paths rather than an external address.
