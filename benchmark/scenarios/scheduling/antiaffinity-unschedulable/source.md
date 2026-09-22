# Unsatisfiable required pod anti-affinity provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/scheduling-eviction/assign-pod-node.md`
- Git blob SHA: `2e3b71391a91f17b64577d2e9b25c582f8523654`
- Relevant section: `Inter-pod affinity and anti-affinity` (lines 236–307 at the
  frozen revision), especially `Types of Inter-pod Affinity and Anti-affinity`
  and `Scheduling Behavior`
- Ground truth: the Deployment requires that no two replicas with label
  `app=app` share a hostname, but the cluster has a single schedulable node, so
  every replacement Pod is rejected by the scheduler. A successful repair
  removes the unsatisfiable required rule (for example by making it preferred)
  while keeping three replicas Ready.
