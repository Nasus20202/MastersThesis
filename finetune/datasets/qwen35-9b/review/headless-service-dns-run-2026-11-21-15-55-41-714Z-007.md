---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/headless-service-dns/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl create service headless app -n default --cluster-ip=None --selector=app=app --port=80 2>`
2. `kubectl apply -f - <<'EOF'`

## Final report

Both StatefulSet replicas (app-0 and app-1) are running. The headless Service has been successfully created with the correct configuration.

**Verification Summary:**

| Check                | Status | Details                    |
| -------------------- | ------ | -------------------------- |
| Service Name         | ✓      | `app` in default namespace |
| clusterIP            | ✓      | `None` (headless)          |
| Selector             | ✓      | `app=app`                  |
| Port                 | ✓      | `80`                       |
| StatefulSet replicas | ✓      | 2/2 ready                  |
| Pods running         | ✓      | app-0, app-1 both Running  |

The headless Service will expose stable per-Pod DNS names: `app-0.app.default.svc` and `app-1.app.default.svc`, allowing pods within the StatefulSet to discover and communicate with each other using these service records.
