# Workload must run with Guaranteed QoS provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/quality-service-pod.md`
- Git blob SHA: `d9bbccfac72627604fc0d64fba73fb06362b9f5d`
- Relevant sections: `Create a Pod that gets assigned a QoS class of Guaranteed`
  (lines 50–107) and `Retrieve the QoS class for a Pod` (lines 250–262) at the
  frozen revision
- Ground truth: a workload whose containers declare no resources is BestEffort.
  To reach Guaranteed QoS every container must set both a CPU and a memory limit
  and an equal request for each resource; the Deployment must stay Ready.
