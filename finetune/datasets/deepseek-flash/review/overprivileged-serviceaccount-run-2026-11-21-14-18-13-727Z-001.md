---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-14-18-13-727Z/overprivileged-serviceaccount/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: unsure: grants get on all configmaps in default; app reads only app-config, not minimal | second pass against the raw record: ok: kubectl patch Role api-reader to configmaps get (msg 23); msgs 25, 29, 33 show denials, in-pod 200 and api-ok; report matches."
---

# overprivileged-serviceaccount · deepseek-flash · run-2026-11-21-14-18-13-727Z · attempt 1

## Task

The application's service account has broader cluster permissions than the application uses. Reduce its permissions to the minimum the application needs without breaking the application.

Documented in:

- content/en/docs/reference/access-authn-authz/rbac.md: API objects > Aggregated ClusterRoles > Role examples, API objects > Referring to resources, ServiceAccount permissions

## Write commands

1. `kubectl patch role api-reader -n default --type=merge -p '{"rules":[{"apiGroups":[""],"resources":["configmaps"],"verbs":["get"]}]}' 2>`

## Final report

Done. The over-permission was in the namespaced `Role/api-reader` bound to `ServiceAccount default/api-reader` (the only SA used by `Deployment/app`); no extra ClusterRoleBinding existed for it — the remaining cluster-level entries (`clustertrustbundles` read, discovery endpoints) come from stock `system:serviceaccounts` bindings and were left untouched.

**What the app actually does** (from the Deployment spec): the container loops `GET https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config` with its SA token — nothing else (no other workload or object uses this SA).

**Change made** (one resource, one field):

```
before: rules: [{"apiGroups":[""],"resources":["configmaps","secrets","pods"],"verbs":["get","list","watch"]}]
after:  rules: [{"apiGroups":[""],"resources":["configmaps"],"verbs":["get"]}]
```

Applied via `kubectl patch role api-reader -n default --type=merge` (roleRef/subjects, the SA, and the Deployment were left intact; no restart or delete was needed since RBAC is evaluated per request).

**Verification after the change**

- App still works, from the pod's own credentials: `GET /api/v1/namespaces/default/configmaps/app-config` → HTTP 200; logs continued `api-ok` at 14:19:36, :46, :56, 14:20:06, :16, :26 (last minute: 6× `api-ok`, 0× `api-fail`).
- Removed capabilities now denied (both `kubectl auth can-i` as the SA and live in-pod requests): `get/list secrets` → `no` / HTTP **403**, `get/list pods` → `no` / HTTP **403**, `list/watch configmaps` → `no`, `get configmaps -n kube-system` → `no`, `get nodes` → `no`, `list pods --all-namespaces` → `no`.
- Retained capability: `get configmaps -n default` → `yes`.
- Controller health: desired=1, updated=1, ready=1, available=1; pod `app-786487cfdf-967xt` Ready=true, restarts=0.
