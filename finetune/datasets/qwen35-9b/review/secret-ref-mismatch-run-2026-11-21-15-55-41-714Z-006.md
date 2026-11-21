---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/secret-ref-mismatch/prompt/006.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: "kubectl edit deployment app --record" content not shown; "kubectl apply -f /tmp/deployment.yaml" has no listed edit to that file. | second pass against the raw record: ok: Step 32 kubectl apply of exported deployment with secretKeyRef passwd changed to password; only the deployment changed, report matches outputs'
---

# secret-ref-mismatch · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 6

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl edit deployment app -n default --record`
2. `kubectl apply -f /tmp/deployment.yaml -n default`

## Final report

## Summary

**Root Cause Identified:**
The Deployment referenced a secret key that didn't exist: `key: passwd`, but the actual secret `app-secret` contained only `password` (different spelling). This caused one pod (`app-c6d597897-8sjvp`) to fail with `CreateContainerConfigError`.

**Fix Applied:**
Edited the Deployment `app` in namespace `default` to change the secret key reference from `passwd` to `password`, matching the actual secret.

**Verification - Before Fix:**

- Deployment status: `3 desired | 1 updated | 4 total | 3 available | 1 unavailable`
- Broken ReplicaSet: `app-c6d597897 (1/1 replicas created)` with one pod stuck
- Pod error: `Error: couldn't find key passwd in Secret default/app-secret`

**Verification - After Fix:**

```
Replicas:               3 desired | 3 updated | 3 total | 3 available | 0 unavailable
```

- All 3 pods running: `app-849cc7cbf4-*` with READY 1/1
- Broken ReplicaSet terminated: `app-c6d597897 (0/0 replicas)`
- Deployment conditions: Available=True, Progressing=True (NewReplicaSetAvailable)

The application rollout is now complete and healthy.
