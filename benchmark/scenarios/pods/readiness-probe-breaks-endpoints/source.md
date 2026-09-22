# Readiness probe breaks endpoints provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md`
- Git blob SHA: `82fb8b2984920b3ea212eeb68b5990f3955eb08e`
- Relevant section: `Define readiness probes` (lines 340–380 at the frozen
  revision)
- Ground truth: the readiness probe targets a path the application does not
  serve, so the Pods never become Ready and the Service loses its endpoints. A
  successful repair restores a probe the application answers, bringing back the
  three intended endpoints.
