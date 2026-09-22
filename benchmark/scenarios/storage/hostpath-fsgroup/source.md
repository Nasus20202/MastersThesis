# Non-root container cannot write a hostPath volume provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/security-context.md`
- Git blob SHA: `36c05cb0e2c45353594ebe631c6b3e9c67f61b73`
- Relevant sections: `Set the security context for a Pod` (lines 53–67) and
  `Discussion` (lines 827–834 at the frozen revision)
- Ground truth: the container runs as a non-root user while the hostPath
  directory is owned by root with restrictive permissions. hostPath volumes do
  not receive `fsGroup` ownership management, so the container cannot write and
  restarts. A successful repair makes the volume writable for the container
  (for example a root init container that chowns it, or `fsGroup` where the
  volume type supports it), after which the volume is writable and the Pod is
  Ready.
