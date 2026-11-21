---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/secret-ref-mismatch/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# secret-ref-mismatch · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl edit deployment/app -n default --dry-run=client -o yaml`
2. `kubectl patch deployment app -n default --type=json -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key", "value": "password"}]'`

## Final report

## Summary

**Root Cause Identified:**  
The deployment `app` in namespace `default` had a secret key reference mismatch:

- The deployment's environment variable `APP_PASSWORD` was configured to read from secret key `passwd`
- The actual secret `app-secret` only contained a key named `password` (singular)

This caused the pod creation to fail with `CreateContainerConfigError`, leaving the rollout stuck.

**Fix Applied:**  
Updated the deployment's environment variable configuration to reference the correct secret key:

```bash
kubectl patch deployment app -n default --type=json \
  -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key", "value": "password"}]'
```

**Verification Results:**

| Metric                               | Before Fix                        | After Fix               |
| ------------------------------------ | --------------------------------- | ----------------------- |
| Status                               | Stuck (1/3 updating)              | Successfully rolled out |
| Replicas                             | 3                                 | 3                       |
| Ready Replicas                       | 3                                 | 3                       |
| Available Replicas                   | 3                                 | 3                       |
| Updated Replicas                     | 1                                 | **3**                   |
| Both pods old ReplicaSet (c6d597897) | ERROR: CreateContainerConfigError | Terminated              |

All pods are now Running and ready. The application is in a healthy state.
