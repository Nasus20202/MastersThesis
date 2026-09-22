# Slow-starting application killed before it serves provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md`
- Git blob SHA: `82fb8b2984920b3ea212eeb68b5990f3955eb08e`
- Relevant section: `Protect slow starting containers with startup probes`
  (lines 302–339 at the frozen revision)
- Ground truth: the application needs about twenty seconds before it serves,
  but the liveness probe starts after a short delay with a small
  `failureThreshold * periodSeconds`, so the kubelet kills the container during
  initialization. A successful repair adds a startup probe whose budget covers
  the worst-case startup time, after which a liveness probe the running
  application answers takes over, so the container starts without restarts and
  becomes Ready.
