# Service selector mismatch provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/debug/debug-application/debug-service.md`
- Git blob SHA: `debcff5ee490b6e7b033a09f71563a5c481ff118`
- Relevant section: `Does the Service have any EndpointSlices?` (lines 420–462
  at the frozen revision)
- Ground truth: the Service selector does not match the healthy application
  Pods, leaving it without usable EndpointSlice backends. A successful repair
  restores both the three intended endpoints and traffic through the Service.
