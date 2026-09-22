# ConfigMap update not visible to running Pods provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/configuration/configmap.md`
- Git blob SHA: `aa3e6ac3c18b995a2057bd1f8ca19eb6861606e7`
- Relevant section: `Mounted ConfigMaps are updated automatically` (lines
  191–208, in particular line 205, at the frozen revision)
- Ground truth: the ConfigMap is consumed as a container environment variable,
  and environment values are resolved when the container is created. Updating
  the ConfigMap to `beta` therefore does not change the running Pods, which
  still report `alpha`. A successful repair restarts the consumers so they read
  `beta` while the Deployment stays ready.
