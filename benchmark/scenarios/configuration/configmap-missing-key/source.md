# ConfigMap key missing from environment provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/configure-pod-configmap.md`
- Git blob SHA: `5a7bbfd02019b027c74341d2861906a3e72e1f88`
- Relevant section: `Define container environment variables using ConfigMap
data` (lines 537–600 at the frozen revision)
- Ground truth: the container environment references a ConfigMap key that does
  not exist, so the container cannot be created
  (`CreateContainerConfigError`). A successful repair makes the application run
  with the intended `message` value while retaining the three-replica workload.
