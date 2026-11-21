---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/secret-ref-mismatch/prompt/005.json
verdict: accepted
reason: "accepted by the researcher for now (forceful write only); first review of a replacement attempt against the raw record by a Haiku model: force: kubectl apply -f /tmp/app-deployment.yaml -n default --force (msg 22) printed configured; plain apply sufficed; report accurate."
---

# secret-ref-mismatch · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 5

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl edit deployment app -n default`
2. `kubectl apply -f /tmp/app-deployment.yaml -n default --force`

## Final report

## Summary

The Kubernetes application rollout issue has been resolved. Here's what was found and fixed:

### Root Cause

The deployment `app` in namespace `default` had a **secret key reference mismatch**. The deployment template incorrectly referenced `key: passwd` but the secret `app-secret` only contained a key named `password`. This caused one new pod to fail with: `Error: couldn't find key passwd in Secret default/app-secret`, resulting in 4 total pods (3 from old config + 1 failing).

### Fix Applied

Updated the deployment configuration by changing:

```yaml
key: passwd # Wrong - key doesn't exist
```

to:

```yaml
key: password # Correct - matches actual secret key
```

### Verification Results

**Before fix:**

- Desired replicas: 3, Total pods: 4, Available: 3 (1 failing)
- One pod in `CreateContainerConfigError` state

**After fix:**

- **Desired replicas: 3** ✅
- **Ready replicas: 3** ✅
- **Available replicas: 3** ✅
- All 3 pods running with status `Running` and READY=1/1
- No failed or pending pods remaining

The deployment rollout is now complete and the application is healthy.
