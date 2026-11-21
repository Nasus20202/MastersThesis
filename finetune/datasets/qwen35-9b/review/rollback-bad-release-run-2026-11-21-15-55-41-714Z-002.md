---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/rollback-bad-release/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl rollout undo deployment/app -n default --to-revision=1`

## Final report

## Summary

**Root Cause Identified:**
The deployment was attempting a rolling update that applied a broken configuration:

- **Bad image/command**: `busybox:1.37.0` with command `sh -c exit 1`
- This caused new pods to immediately fail (exit code 1) with Error status

**Fix Applied:**
Executed rollback to restore the previous working revision:

```bash
kubectl rollout undo deployment/app -n default --to-revision=1
```

This reverted the deployment from revision 2 (broken busybox config) back to revision 1 (working nginx:1.31.5 configuration).

**Verification of Fix:**

| Metric      | Before | After   |
| ----------- | ------ | ------- |
| READY       | 2/3    | **3/3** |
| UP-TO-DATE  | 2      | **3**   |
| AVAILABLE   | 2      | **3**   |
| UNAVAILABLE | 2      | **0**   |

All 3 pods are now Running with status `1/1`:

- `app-575b7496bc-992st` - Running (healthy)
- `app-575b7496bc-ffn62` - Running (healthy)
- `app-575b7496bc-csgqk` - Running (healthy, newly created after rollback)

The broken replica set (`app-76ffd9fd7f`) has been cleaned up and the deployment is now operating with the correct nginx image configuration. The application rollout is complete and healthy.
