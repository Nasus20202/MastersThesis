---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/configmap-update-needs-reload/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# configmap-update-needs-reload · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 4

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl delete pod -n default app-845b8b8485-6mdvg app-845b8b8485-fvfrk app-845b8b8485-zkcx2 --wait --timeout=60s`

## Final report

## Summary

**Issue:** The application configuration ConfigMap was updated to "beta", but running Pods still reported `CONFIG_MESSAGE=alpha`.

**Root Cause:** Kubernetes injects ConfigMap values via environment variable references (`<set to the key 'message' of config map 'app-config'>`) at pod startup. When a ConfigMap is updated, existing pods do not automatically pick up the new value—they need to be recreated.

**Fix Applied:** Deleted all 3 running pods in the deployment, which triggered the Deployment controller to recreate them using the current ConfigMap with "beta".

**Verification:**

- **ConfigMap `app-config` value:** `beta` (confirmed via `kubectl get configmap`)
- **All 3 pods now report:** `CONFIG_MESSAGE=beta` (previously all showed `alpha`)
- Pods: `app-845b8b8485-rjmz4`, `app-845b8b8485-v9jjl`, `app-845b8b8485-wxjrx`

The originally reported symptom (pods reporting "alpha") has been resolved. All running pods now use the updated "beta" configuration value.
