---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/rollback-bad-release/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl rollout undo deployment/app -n default 2>/dev/null`

## Final report

## Summary

**Issue Identified**: The application deployment rolled out to revision 2 with an incorrect container configuration that causes containers to exit with failure code 1. This prevented the rolling update from completing, leaving only 2/3 replicas available.

**Root Cause Evidence**:

- New pods (app-76ffd9fd7f) were using `busybox:1.37.0` with command `sh -c exit 1`, causing immediate container termination
- Old pods (app-575b7496bc) were using `nginx:1.31.5` and running successfully

**Fix Applied**: Executed `kubectl rollout undo deployment/app -n default` to revert the deployment to revision 1 with the correct working configuration.

**Verification of Restored Health**:

- **Deployment Status**: 3/3 READY, 3/3 UP-TO-DATE, 3/3 AVAILABLE (0 unavailable)
- **Pod Status**: All 3 pods Running with 1/1 READY status
- Rolling update now shows: `Replicas: 3 desired | 3 updated | 3 total | 3 available | 0 unavailable`

The application has been successfully restored to a healthy state with all intended replicas running.
