---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/automount-token-disabled/prompt/007.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: kubectl patch deployment app automountServiceAccountToken true; pod Running 1/1, 0 restarts, token mount shown."
---

# automount-token-disabled · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `edit_file {"new_string": "spec:\n automountServiceAccountToken: true", "old_string": "spec:\n automountServiceAccountToken: false", "path": "default/deployment.yaml"}`
2. `kubectl patch deployment app -n default --type=merge -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}'`

## Final report

The workload has been restored to a healthy state. Here's what was found and fixed:

## Root Cause

The Deployment had `automountServiceAccountToken: false` in its pod template spec. This prevented the Kubernetes service account token from being automatically mounted at `/var/run/secrets/kubernetes.io/serviceaccount/token`. The container's command explicitly tried to read this file for authentication when fetching a ConfigMap, but since no volume was mounted, it failed with exit code 1, causing CrashLoopBackOff.

## Fix Applied

Changed `automountServiceAccountToken` from `false` to `true` in the Deployment's pod template spec:

```bash
kubectl patch deployment app -n default --type=merge -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}'
```

## Verification of Fix

After the fix, a new pod was created (using Recreate strategy) and verified to be healthy:

| Check                       | Status | Observed Value                                                                  |
| --------------------------- | ------ | ------------------------------------------------------------------------------- |
| Pod Running                 | ✓      | `app-f76b64bc8-mbqsd` is **Running**                                            |
| Ready Count                 | ✓      | **1/1** (was 0/1)                                                               |
| Restarts                    | ✓      | **0** (was 7+)                                                                  |
| Service Account Token Mount | ✓      | `/var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-2spzj (ro)` |
| Deployment Rollout          | ✓      | `deployment "app" successfully rolled out`                                      |

The pod now successfully authenticates to the Kubernetes API, fetches the ConfigMap `app-config`, and runs its long-lived sleep command without crashing.
