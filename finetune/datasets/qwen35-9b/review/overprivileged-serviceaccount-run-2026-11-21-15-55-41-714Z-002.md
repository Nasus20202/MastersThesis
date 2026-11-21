---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/overprivileged-serviceaccount/prompt/002.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: Sole write "kubectl apply -f /tmp/minimal-role.yaml" applies a file never written in sheet | second pass against the raw record: ok: msg 22 writes /tmp/minimal-role.yaml and applies it (Role now configmaps/get); msg 27-33 confirm api-ok logs and can-i results match the report.'
---

# overprivileged-serviceaccount · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

The application's service account has broader cluster permissions than the application uses. Reduce its permissions to the minimum the application needs without breaking the application.

Documented in:

- content/en/docs/reference/access-authn-authz/rbac.md: API objects > Aggregated ClusterRoles > Role examples, API objects > Referring to resources, ServiceAccount permissions

## Write commands

1. `kubectl apply -f /tmp/minimal-role.yaml`

## Final report

Permissions have been successfully reduced and verified. Here's the summary:

## Changes Made

**Modified Role `api-reader` in namespace `default`:**

| Permission Category | Before           | After          |
| ------------------- | ---------------- | -------------- |
| configmaps          | get, list, watch | **get only** ✓ |
| secrets             | get, list, watch | **removed**    |
| pods                | get, list, watch | **removed**    |

## Verification

1. **Application is running and healthy:**
   - Deployment `app`: 1/1 ready, 1 available
   - Pod status: Running with 0 restarts

2. **Application functionality confirmed:**
   - Logs show continuous "api-ok" output (successful configmap reads)

3. **Permissions reduced as required:**
   - `get configmaps` ✅ allowed
   - `list/watch configmaps` ❌ denied
   - `get secrets` ❌ denied
   - `get pods` ❌ denied

The application only needs to read the specific ConfigMap `app-config`, which now works with just GET permission on ConfigMaps. All unnecessary permissions (LIST, WATCH for ConfigMaps; all access to Secrets and Pods) have been removed while maintaining application functionality.
