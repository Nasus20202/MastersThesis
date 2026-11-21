---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/overprivileged-serviceaccount/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: unsure: grants get on all configmaps in default; app reads only app-config, not minimal | second pass against the raw record: ok: kubectl patch Role to configmaps get (msg 31); msgs 34, 41 show 403s and fresh api-ok lines; no ClusterRoleBinding (msg 15) as reported."
---

# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

The application's service account has broader cluster permissions than the application uses. Reduce its permissions to the minimum the application needs without breaking the application.

Documented in:

- content/en/docs/reference/access-authn-authz/rbac.md: API objects > Aggregated ClusterRoles > Role examples, API objects > Referring to resources, ServiceAccount permissions

## Write commands

1. `kubectl -n default patch role api-reader --type=json -p '[{"op":"replace","path":"/rules","value":[{"apiGroups":[""],"resources":["configmaps"],"verbs":["get"]}]}]' 2>`

## Final report

## What I found

The application is `deployment/app` in namespace `default`, running as service account `default/api-reader`; its only API interaction is one `GET` per 10s on `/api/v1/namespaces/default/configmaps/app-config` (busybox `wget` with the projected SA token).

The over-privilege was entirely in `Role/api-reader` (bound by `RoleBinding/api-reader` → `ServiceAccount/default/api-reader`). Verified no `ClusterRoleBinding` referenced the account, so there was nothing cluster-scoped beyond the default `system:authenticated` discovery roles (confirmed: `get nodes => no`, `get configmaps/pods/secrets -n kube-system|-n kube-public => no`). Within its namespace it held far more than used:

```
rules:
- resources: [configmaps, secrets, pods]   verbs: [get, list, watch]
```

→ effective `get/list/watch` on `secrets` and `pods`, plus `list/watch` configmaps, none of which the app performs.

## Change made (one resource, rules only)

Patched `Role/api-reader` in `default` to the single permission the app actually exercises:

```yaml
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get"]
```

Nothing else was touched: `RoleBinding/api-reader`, the ServiceAccount, the Deployment (still `serviceAccountName: api-reader`), `configmap/app-config`, and the running pod were left as-is.

## Verification (fresh observations)

Effective permissions as the SA (`kubectl auth can-i`):

- `get configmaps/app-config -n default` → **yes** (3/3 checks)
- `get secrets`, `list secrets`, `get pods`, `list pods`, `list configmaps`, `watch configmaps`, `create configmaps`, `delete pods` → **no** (all previously `yes`)
- `kube-system`/`kube-public` configmaps/secrets/pods → **no**; `get nodes` → **no** (no cluster-scope grants)

Direct API calls from inside the running pod with its real token:

- `GET .../configmaps/app-config` → `200`, returned the ConfigMap body (`"name": "app-config"`, `"message": "healthy"`)
- `GET .../secrets` → `403 Forbidden`; `GET .../pods` → `403 Forbidden`

Application still working (original symptom check):

- Pod logs, timestamped after the change: `api-ok` at 12:48:14, 12:48:24, 12:48:34, 12:48:44, 12:48:54, 12:49:04, 12:49:14 — no `api-fail`
- `deployment/app`: spec replicas `1`, ready `1`, updated `1`, available `1`, unavailable none; pod `app-786487cfdf-kzf24` `1/1 Running` with `RESTARTS 0` (same pod, no restart) and no warning events.
