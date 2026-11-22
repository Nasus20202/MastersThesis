---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-20-57-03-137Z/overprivileged-serviceaccount/prompt/002.json
verdict: accepted
reason: 'Role is reduced to get on configmaps in default, consistent with the report.'
---
# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-20-57-03-137Z · attempt 2

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

1. `kubectl apply -f api-reader-role.yaml`
2. `kubectl apply -f api-reader-role.yaml`

## Final report

To reduce the service account's permissions to the minimum required by the application, I inspected the workload, its API usage, and the associated RBAC resources.

### Investigation
- Found a service account named `api-reader` in the `default` namespace.
- The `RoleBinding/api-reader` bound it to `Role/api-reader`, which originally granted `get`, `list`, and `watch` on `configmaps`, `secrets`, and `pods`.
- Verified there were no `ClusterRoleBindings` granting cluster-wide permissions to this service account.
- Inspected the running application pod (`app-786487cfdf-clsfv`) and saw that its only API interaction is a `GET` request to `/api/v1/namespaces/default/configmaps/app-config`.

### Remediation
I updated `Role/api-reader` to remove the unused `pods` and `secrets` resources, as well as the `list` and `watch` verbs. The Role now contains only the minimum required permission:

```yaml
rules:
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
```

### Verification
- Checked the application logs after the change and observed `api-ok`, confirming it can still successfully read the configmap:
  ```text
  api-ok
  api-ok
  api-ok
  api-ok
  ```
- Checked the deployment and confirmed it is healthy and matched the desired replica count:
  ```text
  NAME   READY   UP-TO-DATE   AVAILABLE   AGE
  app    1/1     1            1           64s
  ```
- Confirmed no other bindings reference the `api-reader` service account.

The service account is now restricted to the exact API access exercised by the application (`get` on `configmaps` in its namespace) without breaking functionality.
