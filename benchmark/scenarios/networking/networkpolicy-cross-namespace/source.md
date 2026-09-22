# Cross-namespace NetworkPolicy does not apply provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/network-policies.md`
- Git blob SHA: `22f3515ce15eb97378d7b5bf70c5a8cfab2fdb77`
- Relevant sections: `Behavior of to and from selectors` (lines 142–207),
  `Targeting multiple namespaces by label` (lines 315–352) and `Targeting a
Namespace by its name` (lines 353–360) at the frozen revision
- Ground truth: the client is allowed by label, but the policy entry has only a
  `podSelector`, which selects Pods in the policy's own namespace. A successful
  repair adds a `namespaceSelector` that matches the client namespace, so that
  client is restored while clients in other namespaces remain denied.
