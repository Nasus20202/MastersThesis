# Required ConfigMap update rejected as immutable provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/configuration/configmap.md`
- Git blob SHA: `aa3e6ac3c18b995a2057bd1f8ca19eb6861606e7`
- Relevant section: `Immutable ConfigMaps` (lines 299–327 at the frozen
  revision)
- Ground truth: the ConfigMap is marked `immutable`, so its `data` cannot be
  mutated to the required value `v2`; the documented recovery is to delete and
  recreate the ConfigMap. Because the consumers read it as an environment
  variable, the recreated value only takes effect after the Pods are restarted.
  A successful repair recreates the ConfigMap with `v2` and rolls the consumers,
  so the running Pods report `v2`.
