---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-23-35-22-480Z/overprivileged-serviceaccount/prompt/001.json
verdict: accepted
reason: 'Restricts the Role to get on configmaps, verified by the role output and api-ok logs.'
---
# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-23-35-22-480Z · attempt 1

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

1. `kubectl patch role api-reader -n default --type='json' -p='[{"op": "replace", "path": "/rules", "value": [{"apiGroups":[""],"resources":["configmaps"],"verbs":["get"]}]}]'`

## Final report

I investigated the environment and found the application's service account was indeed overprivileged.

**Investigation Findings:**
- The `app` Deployment in the `default` namespace runs as the `api-reader` service account.
- The Deployment's container executes a loop that fetches a specific ConfigMap:
  `wget ... https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config && echo api-ok || echo api-fail`
- The `api-reader` Role originally granted broad read access (`get`, `list`, `watch`) on `configmaps`, `secrets`, and `pods`.
- The application only ever performs a `get` on a single `configmap`. It does not need access to `secrets`, `pods`, or the `list`/`watch` verbs.

**Action Taken:**
I applied the minimal necessary permissions by updating the `api-reader` Role in the `default` namespace to:

```json
[
  {
    "apiGroups": [""],
    "resources": ["configmaps"],
    "verbs": ["get"]
  }
]
```

**Verification:**
1.  The Role's current rules now show only `get` on `configmaps`:
    ```bash
    [{"apiGroups":[""],"resources":["configmaps"],"verbs":["get"]}]
    ```
2.  I observed the application pod's logs to ensure the change didn't break functionality. The application continues to successfully access the API and logs `api-ok`:
    ```
    api-ok
    api-ok
    api-ok
    ```
The service account permissions have been successfully reduced to the minimum required, and the application is still running without errors.
