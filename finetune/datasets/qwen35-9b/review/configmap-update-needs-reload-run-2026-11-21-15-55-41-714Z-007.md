---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/configmap-update-needs-reload/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# configmap-update-needs-reload · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl rollout restart deployment/app -n default`

## Final report

The issue has been resolved. Here's what I found and fixed:

**Investigation:**

- Running Pods in namespace `default` use environment variable `CONFIG_MESSAGE` from ConfigMap `app-config`
- The ConfigMap already showed `message: beta` (updated value)
- But running pods reported `CONFIG_MESSAGE=alpha` (old baked-in value)
- This happens because Kubernetes bakes environment variables into the Pod spec at creation time; updates to ConfigMaps don't propagate automatically to existing Pods

**Fix Applied:**

- Triggered a rolling restart of deployment `app`: `kubectl rollout restart deployment/app -n default`
- The Deployment controller recreated all 3 pods with the new container specification referencing the updated ConfigMap

**Verification:**
All 3 pods now report the correct value:

| Pod                  | CONFIG_MESSAGE |
| -------------------- | -------------- |
| app-56649f77fc-96wvn | beta ✓         |
| app-56649f77fc-brv98 | beta ✓         |
| app-56649f77fc-z6xzb | beta ✓         |

The running Pods are now using the updated "beta" configuration value.
