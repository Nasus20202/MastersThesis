# Kubernetes Authorization

Every API authorization decision depends on the requesting subject, verb, API group, resource or subresource, resource name where applicable, and namespace.

Subjects are users, groups, or ServiceAccounts. Pods normally run as a ServiceAccount selected by `serviceAccountName`, defaulting to the namespace's `default` ServiceAccount when none is specified. API verbs such as `get`, `list`, `create`, `update`, `patch`, and `delete` are evaluated against API resources; subresources such as `pods/log` and `pods/exec` have separate authorization checks.

A Role contains namespaced policy rules. A ClusterRole is cluster-scoped and can contain rules for namespaced or cluster-scoped resources. RoleBindings grant a Role or ClusterRole within one namespace, while ClusterRoleBindings grant a ClusterRole at cluster scope.

`kubectl auth can-i <verb> <resource> -n <namespace> --as=<subject>` evaluates a requested permission without performing the underlying operation. Error messages from denied API requests commonly identify the requesting subject, verb, resource, API group, and namespace.

Least privilege means granting only the verbs and resource scope required by the workload. Changing a binding from namespaced to cluster-wide scope or granting broad roles such as `cluster-admin` changes far more than one failed request.