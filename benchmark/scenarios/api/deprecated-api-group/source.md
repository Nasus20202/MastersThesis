# Manifest uses a removed API version provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/reference/using-api/deprecation-guide.md`
- Git blob SHA: `7289d518f0b43b0c81d73da753261752d68e66aa`
- Relevant sections: `Deployment` (lines 332–348) and `Migrate to non-deprecated
APIs` (lines 395–408) at the frozen revision
- Ground truth: the `apps/v1beta1` Deployment API is removed and no longer
  served, so the manifest cannot be applied. A successful repair rewrites the
  same workload as `apps/v1` and applies it, giving three Ready replicas.
