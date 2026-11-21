---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/immutable-configmap/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# immutable-configmap · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 3

## Task

The application must run with the configuration value v2 but currently reports v1. Update the application configuration to v2 and make the running Pods use it.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Immutable ConfigMaps

## Write commands

1. `kubectl patch configmap app-config -n default -p '{"data":{"message":"v2"}}'`
2. `kubectl delete configmap app-config -n default 2>`
3. `kubectl apply -f - <<'EOF'`
4. `kubectl rollout restart deployment/app -n default 2>`

## Final report

## Summary

Successfully updated the application configuration from **v1** to **v2**.

### Changes Made:

1. Deleted the immutable ConfigMap `app-config` (which contained `message: v1`)
2. Created a new ConfigMap `app-config` with `message: v2`
3. Triggered deployment rollout to recreate pods

### Final State Verification:

- **ConfigMap**: `app-config` now contains `message: v2` ✓
- **Pods**: All 3 pods are Running and using the new configuration
- **Configuration Values**: All pods show `CONFIG_MESSAGE=v2` ✓

| Pod Name             | Status  | CONFIG_MESSAGE |
| -------------------- | ------- | -------------- |
| app-594f656886-gtbmr | Running | v2             |
| app-594f656886-mh2cf | Running | v2             |
| app-594f656886-twwd8 | Running | v2             |

The application now runs with configuration value **v2** as required.
