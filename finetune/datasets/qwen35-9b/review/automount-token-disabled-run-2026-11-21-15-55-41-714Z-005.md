---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/automount-token-disabled/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# automount-token-disabled · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 5

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `edit_file {"new_string": "spec:\n progressDeadlineSeconds: 600\n replicas: 1\n revisionHistoryLimit: 10\n selector:\n matchLabels:\n app: app\n strategy:\n type: Recreate\n template:\n metadata:\n labels:\n app: app\n spec:\n automountServiceAccountToken: true", "old_string": "spec:\n progressDeadlineSeconds: 600\n replicas: 1\n revisionHistoryLimit: 10\n selector:\n matchLabels:\n app: app\n strategy:\n type: Recreate\n template:\n metadata:\n labels:\n app: app\n spec:\n automountServiceAccountToken: false", "path": "default/replicaset.yaml"}`
2. `kubectl patch deployment app -n default --type=merge -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}'`

## Final report

## Summary

**Root Cause Identified**:
The Deployment's pod template had `automountServiceAccountToken: false`, which prevented the service account token from being mounted at `/var/run/secrets/kubernetes.io/serviceaccount/token`. The container's entrypoint tried to read this token for HTTP authorization via `$($(cat /var/run/secrets/kubernetes.io/serviceaccount/token))`, causing wget to fail with exit code 1 and triggering `CrashLoopBackOff`.

**Fix Applied**:
Changed `automountServiceAccountToken: false` → `true` in the Deployment specification using `kubectl patch`.

**Verification of Resolution**:

- **Before**: Pod `app-cf74c8f5f-w2jn7` was in `CrashLoopBackOff` with 6 restarts and exit code 1
- **After**: New pod `app-f76b64bc8-45d2r` is `Running` with `Ready: True`, 0 restarts
- Deployment status: READY=1/1, UP-TO-DATE=1, AVAILABLE=1
- Volume mount now includes `kube-api-access-ftk82` which provides the service account token projection
- Container successfully completed the wget operation and is now sleeping indefinitely as intended
