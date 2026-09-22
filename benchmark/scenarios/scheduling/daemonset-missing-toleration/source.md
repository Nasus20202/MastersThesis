# DaemonSet missing a node toleration provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/workloads/controllers/daemonset.md`
- Git blob SHA: `090c5477b28b1ec3579eb1019fd928b3d4efbb34`
- Relevant section: `Taints and tolerations` (lines 152–185 at the frozen
  revision)
- Ground truth: the node carries a custom `NoSchedule` taint that the DaemonSet
  Pod template does not tolerate, so the DaemonSet controller no longer targets
  the node and the DaemonSet is not ready on the whole cluster. A successful
  repair adds a matching toleration to the Pod template and leaves the node
  taint in place.
