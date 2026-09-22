# Node selector matches no node provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/scheduling-eviction/assign-pod-node.md`
- Git blob SHA: `2e3b71391a91f17b64577d2e9b25c582f8523654`
- Relevant section: `nodeSelector` (lines 72–100 at the frozen revision)
- Ground truth: the new Pod template requires a node label the cluster nodes do
  not carry, so replacement Pods stay Pending. A successful repair removes the
  unsatisfiable constraint while retaining the intended three-replica workload.
