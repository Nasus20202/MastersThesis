# Missing RBAC binding provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/reference/access-authn-authz/rbac.md`
- Git blob SHA: `c269a6149f5c3ca8aa4817015e159e0aa4446ea2`
- Relevant sections: `RoleBinding examples` (lines 120–160), named-resource
  restrictions through `resourceNames` (lines 197–229), and ServiceAccount
  guidance (lines 1074–1130) at the frozen revision
- Ground truth: the namespaced Role grants only `get` on `app-config`, but its
  binding names the wrong ServiceAccount. A successful repair restores the
  intended subject without granting list access to ConfigMaps or read access to
  Secrets.
