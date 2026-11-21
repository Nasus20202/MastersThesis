---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/configmap-update-needs-reload/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# configmap-update-needs-reload · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl rollout restart deployment/app -n default 2>/dev/null`

## Final report

The fix was successful. Here's the summary:

## Problem

The Kubernetes ConfigMap `app-config` contained `message: beta`, but all running pods showed `CONFIG_MESSAGE=alpha`. This occurred because environment variables injected via `configMapKeyRef` are baked into the pod spec at **creation time**, not updated dynamically.

## Solution

Executed `kubectl rollout restart deployment/app` to trigger recreation of pods with the updated ConfigMap reference.

## Verification

Before fix (3 pods):

```
CONFIG_MESSAGE=alpha  # all three pods
```

After fix (3 new pods, pod-756c795569-*):

```
app-756c795569-d9bh7:   CONFIG_MESSAGE=beta
app-756c795569-h46m2:   CONFIG_MESSAGE=beta
app-756c795569-jz89d:   CONFIG_MESSAGE=beta
```

All 3 pods now report `beta`, matching the updated ConfigMap value. The deployment shows 3/3 replicas ready and available as expected.
