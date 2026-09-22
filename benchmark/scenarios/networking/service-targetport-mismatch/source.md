# Service target port mismatch provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/service.md`
- Git blob SHA: `6687659207aa5a8fc2c46a86a04e0393db14a8f6`
- Relevant section: `Defining a Service` / `targetPort` (lines 128–160 at the
  frozen revision)
- Ground truth: the Service selects the healthy Pods and has ready endpoints,
  but its `targetPort` does not match any container port, so traffic is not
  delivered. A successful repair restores a target port that routes to the
  application while keeping the three intended endpoints.
