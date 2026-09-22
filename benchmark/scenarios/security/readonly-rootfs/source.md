# Read-only root filesystem breaks the application provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/security-context.md`
- Git blob SHA: `36c05cb0e2c45353594ebe631c6b3e9c67f61b73`
- Relevant sections: the `readOnlyRootFilesystem` security context field
  (line 41) and `Set the security context for a Container` (line 412 at the
  frozen revision)
- Ground truth: making the container root filesystem read-only prevents the
  application from writing its runtime files, so it restarts continuously. A
  successful repair keeps the root filesystem read-only and provides writable
  volumes for the paths the application needs.
