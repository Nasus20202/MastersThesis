---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/automount-token-disabled/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# automount-token-disabled · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `kubectl patch deployment app -n default --type=json -p='[{"op": "remove", "path": "/spec/template/spec/automountServiceAccountToken"}]' 2>`

## Final report

The workload has been restored to a healthy state. Here's the verification:

**Root Cause Identified:**
The deployment `app` in namespace `default` had `automountServiceAccountToken: false` in its pod template specification, which prevented the service account token from being mounted at `/var/run/secrets/kubernetes.io/serviceaccount/token`. Without this token, the container's `wget` command failed with "No such file or directory", causing exit code 1 and triggering CrashLoopBackOff.

**Fix Applied:**
Removed `automountServiceAccountToken: false` from the deployment spec using a patch.

**Verification of Outcome:**

- **Before fix:** Pod `app-cf74c8f5f-9ldcd` was in `CrashLoopBackOff` with 6 restarts, READY 0/1
- **After fix:** Pod `app-5478984887-6x5xz` is now `Running` with 0 restarts, READY 1/1

**Deployment Status Confirmed Healthy:**

- Replicas: 1 desired | 1 available (was 0 unavailable before)
- Condition Available: True (was False before)
- All pods in namespace are Running with no non-running resources

The original constraint (restoring the workload to a healthy state) is verified by observing that:

1. The pod READY count matches the desired replicas (1/1 = 1)
2. The deployment status shows Available: True
3. No CrashLoopBackOff or other error states exist
