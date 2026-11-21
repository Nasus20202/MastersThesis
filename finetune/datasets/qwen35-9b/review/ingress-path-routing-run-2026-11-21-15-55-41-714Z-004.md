---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/ingress-path-routing/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# ingress-path-routing · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 4

## Task

The Service app in the default namespace is healthy. Create an Ingress named app that routes HTTP requests for host app.example.com and path /app (pathType Prefix) to Service app on port 80.

Documented in:

- content/en/docs/concepts/services-networking/ingress.md: The Ingress resource > Ingress rules, The Ingress resource > Path types, Types of Ingress > Ingress backed by a single Service

## Write commands

1. `kubectl create ingress app --namespace default --host app.example.com --path /app --path-type Prefix --rule /app:Service/app:80 -o yaml`
2. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: Ingress\nmetadata:\n name: app\n namespace: default\nspec:\n rules:\n - host: app.example.com\n http:\n paths:\n - path: /app\n pathType: Prefix\n backend:\n service:\n name: app\n port:\n number: 80\n", "path": "ingress-app.yaml"}`
3. `kubectl apply -f ingress-app.yaml`

## Final report

The Ingress `app` has been successfully created in the default namespace with the following configuration:

**Created Resource:**

- **Name:** app
- **Namespace:** default
- **Host:** app.example.com
- **Path:** /app (pathType: Prefix)
- **Backend:** Service `app` on port 80
- **Class:** traefik (cluster-provided Ingress controller)

**Verification:**

```
kubectl get ingress -n default
NAME   CLASS     HOSTS             ADDRESS   PORTS   AGE
app    traefik   app.example.com             80      14s
```

The Ingress routes HTTP requests for `app.example.com` with path `/app` to the Service `app` on port 80 as requested. The underlying Service was already healthy and operational.
