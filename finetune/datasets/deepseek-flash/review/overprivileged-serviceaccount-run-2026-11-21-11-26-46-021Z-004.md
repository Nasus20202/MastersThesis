---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/overprivileged-serviceaccount/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

The application's service account has broader cluster permissions than the application uses. Reduce its permissions to the minimum the application needs without breaking the application.

Documented in:

- content/en/docs/reference/access-authn-authz/rbac.md: API objects > Aggregated ClusterRoles > Role examples, API objects > Referring to resources, ServiceAccount permissions

## Write commands

1. `kubectl apply -f role.yaml 2>`

## Final report

## Summary

**Finding.** The only workload using ServiceAccount `default/api-reader` is `deployment/app` (verified by scanning every namespace's pods). Its permission came solely from namespaced `Role/api-reader` bound by `RoleBinding/api-reader`; there are **0** ClusterRoleBindings for this SA, so the "cluster-wide" breadth was a namespace-wide blanket read grant: `get,list,watch` on `configmaps`, `secrets`, `pods` — including readable secrets and pod/logs, none of which the app uses.

**What the app actually needs.** The container loop only issues one request:
`GET https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config`

**Change made (one object, the persistent owner of the faulty state).** Replaced `Role/api-reader`'s rules with the single needed rule, leaving the RoleBinding, ServiceAccount, Deployment, and ConfigMap untouched:

```yaml
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    resourceNames: ["app-config"]
    verbs: ["get"]
```

**Verification (fresh observations, after the change):**

| Check                                                   | Result             |
| ------------------------------------------------------- | ------------------ |
| `get configmaps/app-config` as the SA                   | **yes** (required) |
| `get configmaps`, `list configmaps`, `watch configmaps` | no / no / no       |
| `get configmaps/other-cm` (other name)                  | no                 |
| `list pods`, `get pods/log`                             | no                 |
| `get secrets`, `list secrets`, `delete pods`            | no                 |
| Same read in `kube-system`                              | no                 |
| `list nodes`, `get nodes`, `list persistentvolumes`     | no / no / no       |

**Original symptom gone / app still healthy:**

- In-pod API call returns `"kind": "ConfigMap"` with `"message": "healthy"` (HTTP 200).
- Negative in-pod test on `secrets` now returns `HTTP/1.1 403 Forbidden` — proving the narrowing is effective for the pod's own token.
- Ongoing loop: 2 new log lines appeared after a 22s wait, both `api-ok`; cumulative log universe is `19 api-ok`, **0 api-fail**.
- Deployment converged: `desired=1 updated=1 ready=1 available=1`, `generation=1 observedGeneration=1` — the app was never restarted or disrupted by the change.

**Constraints preserved:** namespace `default`, SA/Deployment names and binding unchanged, no unrelated configuration altered, and nothing was deleted or restarted as a fix.
