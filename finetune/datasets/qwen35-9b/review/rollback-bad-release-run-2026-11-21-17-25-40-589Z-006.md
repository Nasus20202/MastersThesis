---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-17-25-40-589Z/rollback-bad-release/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · Qwen3.5-9B-Q4_K_M · run-2026-11-21-17-25-40-589Z · attempt 6

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl rollout undo deployment/app -n default`

## Final report

The application rollout issue has been resolved. Here's what I found and fixed:

## Root Cause

The deployment was attempting a rolling update from a working configuration (using `nginx:1.31.5`) to a broken configuration (using `busybox:1.37.0` with command `exit 1`). The new pods were failing immediately, causing the rollout to fail because they couldn't pass health checks/readiness.

## Fix Applied

Executed `kubectl rollout undo deployment/app -n default` to roll back to revision 1, which restored the working nginx-based configuration.

## Verification Results

**Pod Status:** All 3 replicas are running and healthy:

```
app-575b7496bc-6mv99   1/1   Running
app-575b7496bc-qznfx   1/1   Running
app-575b7496bc-x8wqd   1/1   Running
```

**Deployment Status:**

- Desired replicas: 3
- Ready replicas: 3/3
- Available replicas: 3
- Unavailable replicas: 0

The application is now in a healthy state with all 3 replicas available and running the correct image (nginx:1.31.5).
