---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/overprivileged-serviceaccount/prompt/002.json
verdict: accepted
reason: 'Restricts the api-reader Role to get on configmap app-config, verified with auth can-i and api-ok logs.'
---
# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 2

## Task

The application's service account has broader cluster permissions than the application uses. Reduce its permissions to the minimum the application needs without breaking the application.

Documented in:
- content/en/docs/reference/access-authn-authz/rbac.md: API objects > Aggregated ClusterRoles > Role examples, API objects > Referring to resources, ServiceAccount permissions

## Inherited stop

The first 15 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 0
stdout:
[{"apiGroups":[""],"resources":["configmaps","secrets"],"verbs":["get","list","watch"]}]
stderr:

Its writes:

1. `kubectl patch role api-reader -n default --type='json' -p='[{"op": "remove", "path": "/rules/0/resources/2"}]'`

## Teacher's write commands

1. `kubectl patch role api-reader -n default --type='json' -p='[{"op": "replace", "path": "/rules", "value": [{"apiGroups": [""], "resources": ["configmaps"], "verbs": ["get"], "resourceNames": ["app-config"]}]}]'`

## Final report

I identified the over-privileged service account and reduced its permissions to exactly what the application requires, then verified the application continues to function.

**Discovery:**
- The application runs as a Deployment named `app` in the `default` namespace using the `api-reader` service account.
- The application's inline command performs a specific API call: `GET /api/v1/namespaces/default/configmaps/app-config`.
- The `api-reader` ServiceAccount had a RoleBinding to a Role also named `api-reader`.
- The original Role granted `get`, `list`, and `watch` on `configmaps`, `secrets`, and `pods` — significantly broader than the application's actual usage.

**Change:**
I patched the `api-reader` Role to restrict permissions to the minimum required by the application:
- **Resources:** `configmaps`
- **Verbs:** `get`
- **Resource Names:** `app-config`

**Verification:**
- I confirmed the effective permissions using `kubectl auth can-i --list` for `system:serviceaccount:default:api-reader`. It now shows only `configmaps [app-config] [get]` alongside standard system discovery endpoints, with no other resource access. No ClusterRoleBindings exist for this service account.
- I checked the application logs after the change; the application successfully reached the API and printed `api-ok` on its most recent iterations.
- The deployment status confirms `1` ready of `1` desired replicas, and the pod `app-786487cfdf-jmn6z` is `1/1 Running`.
