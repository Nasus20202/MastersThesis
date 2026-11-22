---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-20-57-03-137Z/overprivileged-serviceaccount/prompt/003.json
verdict: accepted
reason: 'Role is narrowed to get on configmaps with resourceNames app-config, matching the reported auth checks.'
---
# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-20-57-03-137Z · attempt 3

## Task

The application's service account has broader cluster permissions than the application uses. Reduce its permissions to the minimum the application needs without breaking the application.

Documented in:
- content/en/docs/reference/access-authn-authz/rbac.md: API objects > Aggregated ClusterRoles > Role examples, API objects > Referring to resources, ServiceAccount permissions

## Inherited stop

The first 17 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 1
stdout:

stderr:
The Role "api-reader" is invalid: 
* rules[0].resources: Required value: resource rules must supply at least one resource
* rules[0].verbs: Required value

Its writes:

1. `kubectl patch role api-reader --namespace default -p '{"rules": [{"apiGroups": ["", "apps", ""]}]}'`

## Teacher's write commands

1. `kubectl patch role api-reader -n default -p '{"rules":[{"apiGroups":[""],"resources":["configmaps","secrets"],"verbs":["get","list","watch"]}]}'`
2. `kubectl patch role api-reader -n default -p '{"rules":[{"apiGroups":[""],"resources":["configmaps"],"resourceNames":["app-config"],"verbs":["get"]}]}'`

## Final report

The application's service account has been reduced to the minimum permissions required without breaking the application.

### Diagnosis
- The `app` deployment in the `default` namespace uses the `api-reader` service account.
- The application's sole function is to fetch the `app-config` ConfigMap from the `default` namespace using its service account token.
- The original `api-reader` Role was overly broad, granting `get`, `list`, and `watch` on `configmaps`, `secrets`, and `pods` across the entire `default` namespace.

### Change Applied
- Patched `Role/api-reader` in the `default` namespace to remove all unnecessary permissions.
- The Role now only allows the `get` verb on `configmaps` and is further restricted by `resourceNames` to only the specific `app-config` ConfigMap.

### Verification
- **Application Health:** The `app` deployment shows `1/1` ready replicas, and its recent logs confirm successful API access (`api-ok`).
- **Permissions Check (`kubectl auth can-i`):**
  - `get configmaps/app-config` → `yes`
  - `get secrets` → `no`
  - `get pods` → `no`
  - `list configmaps` → `no`
  - No `ClusterRoleBindings` exist for this service account, confirming no broader cluster-level permissions remain.
