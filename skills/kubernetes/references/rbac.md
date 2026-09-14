# Kubernetes authorization

Resolve authorization from the identity outward. Identify the Pod's
ServiceAccount, then inspect the referenced Role or ClusterRole and the
RoleBinding or ClusterRoleBinding that grants it. Namespace scope matters for
both the subject and the binding; `roleRef` identifies the granted role.

- Check the exact verb, API group, resource and resource name involved.
- Use `kubectl auth can-i` with the relevant identity and namespace to test the
  observed operation, and compare the result with the declared binding.
- Prefer a namespaced Role and RoleBinding when cluster-wide access is not
  required. Retain existing restrictions and add only the smallest missing
  rule or binding.
- Treat a successful authorization check as evidence about one action, not as
  proof that every related action is permitted.

Do not resolve an authorization problem by granting broad administrator access
or by replacing the workload identity without evidence.
