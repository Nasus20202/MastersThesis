---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/statefulset-headless-service/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-headless-service · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

A StatefulSet's Pods are running, but other workloads cannot reach them through their stable per-Pod DNS names. Restore stable network identity for the StatefulSet's Pods.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Pod Identity > Stable Network ID

## Write commands

1. `kubectl apply -n default -f - <<'EOF'`

## Final report

The issue has been diagnosed and fixed. Here's the summary:

## Root Cause

The StatefulSet `app` in namespace `default` requires a headless Service named `app` for stable per-Pod DNS names (like `app-0.app`, `app-1.app`). This service was missing from the cluster, causing DNS resolution failures for all workloads trying to reach the StatefulSet pods through their stable DNS identifiers.

## Fix Applied

Created a headless Service that matches the StatefulSet's expectations:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: app
  namespace: default
spec:
  selector:
    app: app
  ports:
    - port: 80
      targetPort: http
      protocol: TCP
  clusterIP: None
```

## Verification

DNS resolution for stable pod names now works correctly:

| DNS Name                              | Resolved IP  | Verified |
| ------------------------------------- | ------------ | -------- |
| `app-0.app.default.svc.cluster.local` | `10.244.0.4` | ✓        |
| `app-1.app.default.svc.cluster.local` | `10.244.0.6` | ✓        |

Network connectivity confirmed:

- `curl http://app-1.app.default.svc.cluster.local:80/` from app-0 pod → Returns nginx welcome page ✓
- `curl http://app-0.app.default.svc.cluster.local:80/` from app-1 pod → Returns nginx welcome page ✓

The headless Service now properly provides endpoint IPs (10.244.0.4:80, 10.244.0.6:80) for stable DNS name resolution, restoring network identity for the StatefulSet's Pods.
