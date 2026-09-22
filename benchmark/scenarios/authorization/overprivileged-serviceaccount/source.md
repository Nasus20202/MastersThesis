# Over-privileged service account provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/reference/access-authn-authz/rbac.md`
- Git blob SHA: `c269a6149f5c3ca8aa4817015e159e0aa4446ea2`
- Relevant sections: `Role examples` (lines 352–439), `Referring to resources`
  (lines 168–255) and `ServiceAccount permissions` (lines 1074–1160) at the
  frozen revision
- Ground truth: the application only reads a named ConfigMap, but its Role also
  grants verbs and resources it never uses. A successful repair reduces the Role
  to a single `get` rule on `configmaps` while the in-cluster API call continues
  to succeed.
