# Republished image tag is not picked up provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/containers/images.md`
- Git blob SHA: `c984d76127520a9e7dc4326048734a9144c37a1c`
- Relevant section: `Updating images` / `Image pull policy` (the `IfNotPresent` default and its effect on mutable tags)
- Ground truth: with `imagePullPolicy: IfNotPresent` the kubelet reuses the
  locally cached image for an unchanged tag, so republishing the same tag does
  not take effect on the running Pods. A successful repair forces a fresh
  resolution of the tag (for example `imagePullPolicy: Always`) or references
  the published image by digest, so every replica runs the currently published
  build.
