# Kubernetes Authorization

Every API authorization decision depends on the requesting subject, verb, API group, resource or subresource, resource name where applicable, and namespace.

Subjects are users, groups, or ServiceAccounts. Pods normally run as a ServiceAccount selected by `serviceAccountName`, defaulting to the namespace's `default` ServiceAccount when none is specified. API verbs such as `get`, `list`, `create`, `update`, `patch`, and `delete` are evaluated against API resources; subresources such as `pods/log` and `pods/exec` have separate authorization checks.

A Role contains namespaced policy rules. A ClusterRole is cluster-scoped and can contain rules for namespaced or cluster-scoped resources. RoleBindings grant a Role or ClusterRole within one namespace, while ClusterRoleBindings grant a ClusterRole at cluster scope.

`kubectl auth can-i <verb> <resource> -n <namespace> --as=<subject>` evaluates a requested permission without performing the underlying operation. Error messages from denied API requests commonly identify the requesting subject, verb, resource, API group, and namespace.

## Repairing a workload permission

Start from the workload's `serviceAccountName` and namespace. Inspect the matching ServiceAccount, the RoleBindings in that namespace, and the Role or ClusterRole named by each applicable binding. Check cluster-wide bindings only when namespace-scoped bindings do not explain the permission. Avoid dumping every ClusterRoleBinding as the first diagnostic query.

If an existing binding grants the intended role but its subject names the wrong ServiceAccount, correct that subject instead of adding another Role or RoleBinding. Preserve other subjects and the existing `roleRef`. For a ServiceAccount subject, set `kind`, `name`, and `namespace`; leave `apiGroup` unset. The `apiGroup` field is used for User and Group subjects.

After the change, test the exact permission the workload needs with `kubectl auth can-i`. Also check any narrower scope or prohibited access named by the task.

For example, after inspecting the binding and confirming the target subject is element zero, update only its name:

```bash
kubectl patch rolebinding/my-app-access -n my-namespace --type=json \
  -p='[{"op":"replace","path":"/subjects/0/name","value":"my-app"}]'
```

Use the index from the live binding; do not assume the target is always the first subject.

Least privilege means granting only the verbs and resource scope required by the workload. Changing a binding from namespaced to cluster-wide scope or granting broad roles such as `cluster-admin` changes far more than one failed request.
