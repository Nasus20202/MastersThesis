---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/ingress-path-routing/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# ingress-path-routing · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

The Service app in the default namespace is healthy. Create an Ingress named app that routes HTTP requests for host app.example.com and path /app (pathType Prefix) to Service app on port 80.

Documented in:

- content/en/docs/concepts/services-networking/ingress.md: The Ingress resource > Ingress rules, The Ingress resource > Path types, Types of Ingress > Ingress backed by a single Service

## Write commands

1. `kubectl apply -f -`

## Final report

Done. The Ingress was created against the cluster's default IngressClass (Traefik) and verified to actually route traffic.

**What I found first (evidence, not assumptions)**

- `Service app` in `default` exists: ClusterIP `10.96.118.155`, port `80/tcp` → targetPort 80, selector `app: app`, with a ready endpoint `10.244.0.6:80` (EndpointSlice `app-xm8q6`), and pod `app-bb695fb5-nm7t7` `1/1 Running` (nginx).
- No Ingresses existed (`kubectl get ingress -A` → none).
- IngressClass `traefik` is marked `(default)`, controller `traefik.io/ingress-controller`, pod `traefik-6d9c545874-zlbdt` Running. So Traefik, not another controller, would serve the object — hence `ingressClassName: traefik` was set explicitly rather than left implicit.

**Change made (single object created)**

```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata: {name: app, namespace: default}
spec:
  ingressClassName: traefik
  rules:
  - host: app.example.com
    http:
      paths:
      - path: /app
        pathType: Prefix
        backend: {service: {name: app, port: {number: 80}}}
```

Nothing else was touched — the Service, its selector/ports, and the controller were left as-is.

**Verification (direct observations)**

- Final spec: `NAME=app CLASS=traefik HOST=app.example.com PATH=/app TYPE=Prefix SVC=app PORT=80`.
- Request through the ingress controller (port-forward to `svc/traefik`, then HTTP with the required Host header):
  - `GET /app` with `Host: app.example.com` → **HTTP/1.1 200 OK**, `Server: nginx/1.31.5`, body = nginx welcome page (896 bytes) — i.e. it reached the `app` backend.
  - Negative controls prove the rule is host- and path-scoped, not a catch-all: `Host: other.example.com` → **404**, and `Host: app.example.com` path `/other` → **404**.
- Backend-side proof: the app pod's access log recorded `10.244.0.5 - - "GET /app HTTP/1.1" 200` — source `10.244.0.5` is the Traefik pod IP, so the 200 came through the ingress controller into Service `app`, not from a direct call.
- Service unchanged and still healthy: `ClusterIP 10.96.118.155`, endpoint `10.244.0.6:80`.

One environmental note, not a defect in the object: `status.loadBalancer` on the Ingress is empty because Traefik's own Service is `type: LoadBalancer` with `EXTERNAL-IP <pending>` (kind has no cloud LB provisioner), so no address is published to copy into Ingress status. Routing works regardless, as demonstrated via the controller's ClusterIP/NodePort (`80:30376`, `443:32164`).
