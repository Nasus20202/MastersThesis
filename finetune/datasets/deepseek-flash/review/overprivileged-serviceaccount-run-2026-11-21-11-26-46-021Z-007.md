---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/overprivileged-serviceaccount/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The application's service account has broader cluster permissions than the application uses. Reduce its permissions to the minimum the application needs without breaking the application.

Documented in:

- content/en/docs/reference/access-authn-authz/rbac.md: API objects > Aggregated ClusterRoles > Role examples, API objects > Referring to resources, ServiceAccount permissions

## Write commands

1. `kubectl apply -f role-min.yaml`

## Final report

Done. Here's what I found and changed, with the evidence proving each check.

## Diagnosis

- Only one workload uses the service account: deployment `default/app` (pod `app-786487cfdf-ssbhj`, `serviceAccountName: api-reader`).
- The app's container loop does exactly one API call every 10s: `GET https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config`.
- `api-reader` had no ClusterRoleBinding — its only grant was `RoleBinding default/api-reader` → `Role default/api-reader`, but that Role was far broader than usage:
  - resources: `configmaps`, **`secrets`**, **`pods`**
  - verbs: `get`, **`list`**, **`watch`** (list/watch on configmaps were unused, and list on secrets/pods is a notable read escalation)

## Change (smallest justified)

Replaced the Role's rule with only the call the app makes — a single `get` on the one ConfigMap it reads (recorded the prior Role at `/tmp/role-backup.yaml` first). The ServiceAccount and RoleBinding were left untouched.

```yaml
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["app-config"]
    verbs: ["get"]
```

`kubectl apply` result: `role.rbac.authorization.k8s.io/api-reader configured`.

## Verification

**Reduced permissions (authorization checks as the SA):**

| query                                    | before | after   |
| ---------------------------------------- | ------ | ------- |
| `get configmaps/app-config` (ns default) | yes    | **yes** |
| `list` / `watch configmaps`              | yes    | **no**  |
| `get` / `list secrets`                   | yes    | **no**  |
| `get` / `list pods`                      | yes    | **no**  |
| any of the above cluster-wide (`-A`)     | –      | **no**  |

`auth can-i --list` for the SA now shows only `configmaps [app-config] [get]`; everything else listed is the baseline `system:discovery` / `system:public-info-viewer` access that `system:authenticated` gets cluster-wide, not something this SA was granted.

**App not broken (live in-pod calls with the real projected token):**

- `GET .../configmaps/app-config` → **OK**
- `GET .../secrets` → **403 Forbidden** (was allowed before)
- `GET .../pods` → **403 Forbidden** (was allowed before)

**Sustained health after the change:** 8 consecutive `api-ok` log lines spanning 13:23:58 → 13:25:09 (all post-change); the single `api-fail` in the log is timestamped 13:21:08, i.e. pre-change startup. Deployment status matches the request exactly: **1 requested / 1 total / 1 updated / 1 ready / 1 available**, pod `1/1 Running` with 0 restarts, no warning events.
