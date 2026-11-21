---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/ingress-path-routing/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# ingress-path-routing · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 1

## Task

The Service app in the default namespace is healthy. Create an Ingress named app that routes HTTP requests for host app.example.com and path /app (pathType Prefix) to Service app on port 80.

Documented in:

- content/en/docs/concepts/services-networking/ingress.md: The Ingress resource > Ingress rules, The Ingress resource > Path types, Types of Ingress > Ingress backed by a single Service

## Write commands

1. `kubectl apply -f - << 'EOF'`

## Final report

The Ingress `app` has been successfully created in the default namespace with all required specifications:

- **Name**: `app`
- **Namespace**: `default`
- **Host**: `app.example.com`
- **Path**: `/app` with `pathType: Prefix`
- **Backend Service**: `app` on port `80`

The Ingress has been accepted by the cluster and is ready to route traffic.
