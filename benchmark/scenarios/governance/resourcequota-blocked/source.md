# ResourceQuota blocks replicas provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/policy/resource-quotas.md`
- Git blob SHA: `a7436db1377af334e45b1ebd02e5ea52c067d167`
- Relevant sections: `How Kubernetes ResourceQuotas work` (lines 31–100) and
  the admission-enforcement description of a `403 Forbidden` rejection
  (line 49 at the frozen revision)
- Ground truth: the namespace ResourceQuota allows only two Pods, so the third
  replica is rejected at admission. A successful repair raises the quota's Pod
  limit while keeping the quota in place.
