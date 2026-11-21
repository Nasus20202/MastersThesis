---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/rollback-bad-release/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl rollout undo deployment/app -n default 2>`

## Final report

## Summary

**Issue Identified:** The deployment `app` in namespace `default` had failed its rolling update to revision 2, which contained a container command of `exit 1`, causing all new pods to fail immediately with exit code 1. Only 2 out of 3 replicas were available.

**Fix Applied:** Used `kubectl rollout undo deployment/app` to revert to revision 1, which was the original healthy configuration using nginx image without the failing command.

**Verification - Before Fix:**

- Deployment: READY=2/3, AVAILABLE=2
- Failing pods in replica set `app-76ffd9fd7f`: CrashLoopBackOff and Error states

**Verification - After Fix:**

- Deployment: READY=3/3, UP-TO-DATE=3, AVAILABLE=3
- All 3 pods Running in replica set `app-575b7496bc` (original revision)
- Replica set `app-76ffd9fd7f`: 0 desired replicas (cleaned up by rollback)

The application has been restored to a healthy state with all requested replicas running and available.
