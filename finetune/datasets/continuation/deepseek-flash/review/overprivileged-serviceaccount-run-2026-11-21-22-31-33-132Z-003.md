---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/overprivileged-serviceaccount/prompt/003.json
verdict: accepted
reason: 'Applies a Role limited to get on app-config, verified with auth can-i checks and api-ok logs.'
---
# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 3

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

1. `kubectl apply -f -`

## Final report

The application's service account has been restricted to the minimum permissions required for it to function.

**Observed Evidence & Diagnosis:**
- The application pod (`app`) runs under the `api-reader` ServiceAccount.
- The `api-reader` Role originally granted broad permissions: `get`, `list`, and `watch` on `configmaps`, `secrets`, and `pods`.
- Inspecting the deployment's command showed the application only performs a single action: an HTTP `GET` request to fetch a specific ConfigMap named `app-config` (`/api/v1/namespaces/default/configmaps/app-config`).

**Change Applied:**
I updated the `api-reader` Role to grant only the exact operation needed:
```yaml
rules:
- apiGroups: [""]
  resources: ["configmaps"]
  resourceNames: ["app-config"]
  verbs: ["get"]
```

**Verification:**
1. **Application Health:** The application logs were monitored post-change. The continuous loop now prints `api-ok` repeatedly, confirming the pod successfully fetches the configuration.
2. **Permission Boundaries:** Using `kubectl auth can-i`, the ServiceAccount now correctly shows:
   - `Get app-config`: **yes**
   - `Get other configmaps`: **no**
   - `List configmaps`: **no**
   - `Get secrets`: **no**
   - `Get pods`: **no**

The application remains fully operational while its cluster permissions have been reduced to the smallest justified scope.
