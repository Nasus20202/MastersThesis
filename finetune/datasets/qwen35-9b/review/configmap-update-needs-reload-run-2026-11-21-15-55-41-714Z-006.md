---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/configmap-update-needs-reload/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# configmap-update-needs-reload · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 6

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl rollout restart deployment/app -n default 2>`

## Final report

The issue has been resolved. Here's what happened and the verification:

**Root Cause:** The ConfigMap `app-config` was updated to have `message: beta`, but the existing Pods were created when the ConfigMap had `alpha`. Kubernetes environment variables from ConfigMaps are not dynamically refreshed for running containers - they're baked in at container creation time.

**Fix Applied:** Ran `kubectl rollout restart deployment/app -n default` which triggered a rolling update, causing all pods to be recreated with fresh ConfigMap references.

**Verification:**

- Deployment status: READY=3/3, UP-TO-DATE=3, AVAILABLE=3
- All 3 Pods now report `CONFIG_MESSAGE=beta`:
  - `app-6f4f4cf8fc-9ggxb`: CONFIG_MESSAGE=beta
  - `app-6f4f4cf8fc-lg4dd`: CONFIG_MESSAGE=beta
  - `app-6f4f4cf8fc-wgkgp`: CONFIG_MESSAGE=beta

The running Pods now use the updated configuration value.
