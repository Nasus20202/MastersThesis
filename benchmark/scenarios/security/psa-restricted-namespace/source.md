# Pod rejected by Pod Security Admission provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md`
- Git blob SHA: `1bfce66d1d2df9df451a7a0c0e7d86b657f49b54`
- Relevant sections: `Requiring the baseline Pod Security Standard with namespace
labels` (lines 24–48) and `Applying to a single namespace` (lines 84–93) at
  the frozen revision
- Ground truth: the namespace enforces the `restricted` Pod Security Standard,
  so a Pod without the required security context is rejected at admission. A
  successful repair adds the restricted requirements to the Pod (non-root user,
  `RuntimeDefault` seccomp, no privilege escalation, all capabilities dropped)
  without weakening the namespace labels.
