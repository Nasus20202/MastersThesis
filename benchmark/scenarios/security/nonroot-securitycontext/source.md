# Container rejected because it runs as root provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/security-context.md`
- Git blob SHA: `36c05cb0e2c45353594ebe631c6b3e9c67f61b73`
- Relevant sections: `Set the security context for a Pod` (lines 53–162) and
  `Set the security context for a Container` (lines 412–466) at the frozen
  revision
- Ground truth: the Pod sets `runAsNonRoot: true` but does not set a numeric
  user, and the image runs as root, so the kubelet refuses to start the
  container. A successful repair sets an explicit non-zero `runAsUser` so the
  Pod starts with a non-root effective UID.
