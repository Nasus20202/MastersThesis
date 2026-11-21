---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/headless-service-dns/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 1

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl create service headless app -n default --selector app=app --port 80 --dry-run=client -o yaml`
2. `kubectl apply -f /tmp/app-service.yaml`

## Final report

**Verification complete. The headless Service has been created with all required specifications:**

| Requirement        | Observed Value    | Status |
| ------------------ | ----------------- | ------ |
| Name: app          | `app`             | ✓      |
| Namespace: default | `default`         | ✓      |
| clusterIP: None    | `None` (headless) | ✓      |
| selector: app=app  | `"app": "app"`    | ✓      |
| port: 80           | `port: 80`        | ✓      |

**Endpoints confirm both StatefulSet pods are registered:** `10.244.0.4:80,10.244.0.6:80`

The headless Service now provides stable per-Pod DNS names for the StatefulSet's pods via FQDNs like `app-0.app.default.svc.cluster.local` and `app-1.app.default.svc.cluster.local`.
