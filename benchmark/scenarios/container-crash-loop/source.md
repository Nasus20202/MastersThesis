# Container crash loop provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/debug/debug-application/debug-running-pod.md`
- Git blob SHA: `287d4d6d0b86376d4ef771d105edbf5d82dba972`
- Relevant section: `Copying a Pod while changing its command`, including the
  crashing-command and `CrashLoopBackOff` example (lines 514–562 at the frozen
  revision)
- Ground truth: invalid nginx arguments make each new container exit. A
  successful repair restores a valid application process while retaining the
  intended three-replica workload.
