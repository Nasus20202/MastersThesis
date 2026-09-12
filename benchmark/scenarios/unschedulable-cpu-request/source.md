# Unschedulable CPU request provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/configuration/manage-resources-containers.md`
- Git blob SHA: `cbd80b3d4d61f92226570516bac9f94ceb9155a3`
- Relevant sections: `How Pods with resource requests are scheduled` (lines
  248–259) and the `FailedScheduling` / insufficient CPU example (lines
  582–608) at the frozen revision
- Ground truth: the new Pod template requests more CPU than any cluster node
  can supply. A successful repair makes the request schedulable while retaining
  the intended three-replica workload.
