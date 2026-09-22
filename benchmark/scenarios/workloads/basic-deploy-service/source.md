# Deploy a workload and expose it with a Service provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tutorials/kubernetes-basics/expose/expose-intro.md`
- Git blob SHA: `0d2008a52f3abe15ea2c40a399c8bcfe255aef20`
- Relevant sections: `Services and Labels` (lines 75–95) and `Step 1: Creating
a new Service` (lines 96–173) at the frozen revision
- Ground truth: the namespace is empty. A successful repair creates a
  two-replica nginx Deployment and a ClusterIP Service that selects its Pods on
  port 80, so the Service proxies the nginx welcome page.
