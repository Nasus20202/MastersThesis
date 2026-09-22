# Taint and toleration placement provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/scheduling-eviction/taint-and-toleration.md`
- Git blob SHA: `2819bcf9f0645c9d9486575115e23834be04b2d2`
- Relevant sections: `Concepts` and the toleration example (lines 29–70) and
  `Example Use Cases` for dedicated nodes (lines 245–255 at the frozen
  revision)
- Ground truth: the single node is tainted `dedicated=gpu:NoSchedule`, so only
  Pods with a matching toleration can be scheduled. A successful repair deploys
  a three-replica workload with that toleration while leaving the taint in
  place.
