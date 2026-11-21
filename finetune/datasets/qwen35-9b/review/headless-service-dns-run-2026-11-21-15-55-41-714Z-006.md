---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/headless-service-dns/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 6

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl apply -n default -f - <<'EOF'`
2. `kubectl run -n default -it --rm debug --image=curlimages/curl --restart=Never -- curl -s -o /dev/null -w "%{http_code}" app-0.app.default.svc 2>`
3. `kubectl run -n default -it --rm debug --image=curlimages/curl --restart=Never -- sh -c 'for i in app-0.app.default.svc app-1.app.default.svc`

## Final report

Both pods are successfully accessible through the headless Service:

| Pod   | DNS Name              | Status       | HTTP Code |
| ----- | --------------------- | ------------ | --------- |
| app-0 | app-0.app.default.svc | ✓ Accessible | 200       |
| app-1 | app-1.app.default.svc | ✓ Accessible | 200       |

**Summary of changes:**

- Created headless Service `app` in namespace `default` with:
  - `clusterIP: None` (headless)
  - `selector: app=app`
  - `port: 80`

**Verified outcome:**

- DNS resolution works for both StatefulSet pods via their stable per-Pod DNS names
- HTTP requests to both `app-0.app.default.svc` and `app-1.app.default.svc` return HTTP 200, confirming the pods are accessible through the headless Service
