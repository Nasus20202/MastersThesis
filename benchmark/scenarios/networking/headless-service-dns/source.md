# Headless Service for StatefulSet DNS provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/service.md`
- Git blob SHA: `6687659207aa5a8fc2c46a86a04e0393db14a8f6`
- Relevant section: `Headless Services` (line 880 at the frozen revision)
- Ground truth: the StatefulSet references a headless Service by `serviceName`,
  but that Service is absent, so per-Pod DNS records do not exist. A successful
  repair creates a `clusterIP: None` Service selecting the two StatefulSet Pods.
