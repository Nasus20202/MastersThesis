# Job exhausted its backoff limit provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/workloads/controllers/job.md`
- Git blob SHA: `9a236bbf64464afc0216c5537de6b307af25980d`
- Relevant section: `Pod backoff failure policy` (lines 505–536 at the frozen
  revision)
- Ground truth: the Job Pod command fails and the Job reaches its backoff
  limit, so it is marked Failed. Because a Job's Pod template is immutable, a
  successful repair replaces the Job with one whose container command exits
  successfully.
