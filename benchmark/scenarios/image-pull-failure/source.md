# Image pull failure provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/containers/images.md`
- Git blob SHA: `c984d76127520a9e7dc4326048734a9144c37a1c`
- Relevant section: `ImagePullBackOff` (lines 176–188 at the frozen revision)
- Ground truth: the Deployment references an image that cannot be pulled. A
  successful repair restores a pullable application image while retaining the
  intended three-replica workload.
