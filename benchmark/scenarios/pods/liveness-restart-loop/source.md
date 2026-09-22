# Liveness probe restarts a healthy container provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md`
- Git blob SHA: `82fb8b2984920b3ea212eeb68b5990f3955eb08e`
- Relevant section: `Define a liveness HTTP request` (lines 110–191 at the
  frozen revision)
- Ground truth: the liveness probe targets a path and port the container does
  not serve, so every probe returns a failure code and the kubelet kills and
  restarts the otherwise healthy container. A successful repair makes the
  liveness probe check an endpoint the application answers, so the containers
  stop restarting and stay Ready.
