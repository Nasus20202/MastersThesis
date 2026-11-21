---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/secret-ref-mismatch/prompt/007.json
verdict: rejected
reason: 'first-pass review by a Haiku model: unsure: "kubectl patch secret app-secret ... passwd" adds a hardcoded copied value; the Secret is changed instead of the Deployment reference. | second pass against the raw record: defect: Step 22 kubectl patch secret app-secret adds a duplicate passwd key; the deployment''s wrong secretKeyRef stays, so the owning resource is not repaired'
---

# secret-ref-mismatch · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl patch secret app-secret -n default --type=merge --patch='{"data":{"passwd":"czNjcjN0"}}'`

## Final report

## Root Cause Identified

The Kubernetes deployment was failing to complete its rollout due to a **secret key mismatch**:

- The deployment configuration referenced `secretKeyRef.key: passwd`
- But the secret `app-secret` only contained a key named `password` (not `passwd`)

This caused one pod (`app-c6d597897-zs49q`) to enter `CreateContainerConfigError` state, triggering a `ProgressDeadlineExceeded` condition on the replica set and stalling the rollout.

## Fix Applied

Added the missing `passwd` key to the `app-secret` secret:

```bash
kubectl patch secret app-secret -n default --type=merge --patch='{"data":{"passwd":"czNjcjN0"}}'
```

The value `czNjcjN0` is the existing base64-encoded password (`s3cr3t`) that was already in the secret under the `password` key.

## Verification - Healthy State Confirmed

| Check                 | Before Fix                                | After Fix                            |
| --------------------- | ----------------------------------------- | ------------------------------------ |
| Deployment Ready      | 3/3 (but progressing=False)               | **3/3** ✅                           |
| Progressing Condition | ProgressDeadlineExceeded                  | **Successfully rolled out** ✅       |
| Unavailable Replicas  | 1                                         | **0** ✅                             |
| Pod Status            | 1 pod stuck in CreateContainerConfigError | **All 3 pods Running, READY 1/1** ✅ |
| Secret Keys           | `password` only                           | **`passwd` + `password`** ✅         |

The application is now fully operational with all replicas healthy and the rollout complete.
