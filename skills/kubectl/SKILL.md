---
name: kubectl
description: Choose focused kubectl commands and interpret Kubernetes object output without guessing names, scope or relationships.
metadata:
  version: "1.0"
  scope: Kubernetes command-line inspection
---

# kubectl

Run kubectl through Bash using the provided kubeconfig and context. Discover
resource names, namespaces and labels from the cluster before operating on
them.

- Start with read-only `get`, `describe`, `logs` and event inspection. Use an
  explicit namespace when the task has a namespace scope.
- Use `-o yaml`, `-o json` or a narrow `-o jsonpath=...` expression when exact
  fields, conditions, labels or relationships matter. Treat table output as a
  summary, not the complete object.
- Follow ownership and references between objects before changing a generated
  object. Prefer the declarative owner and preserve fields unrelated to the
  observed problem.
- Use rollout/status and other bounded checks to observe convergence. Capture
  stderr and the exit status when a command fails.
- Use authorization queries only to test the identity and action in question;
  do not broaden permissions merely to make an operation pass.

When syntax or a field is uncertain, use `kubectl explain` or the command's
help output rather than inventing a field name.
