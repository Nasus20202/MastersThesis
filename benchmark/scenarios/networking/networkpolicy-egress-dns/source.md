# Default-deny egress blocks DNS and a dependency provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/network-policies.md`
- Git blob SHA: `22f3515ce15eb97378d7b5bf70c5a8cfab2fdb77`
- Relevant sections: `Behavior of to and from selectors` (lines 142–207) and
  `Default deny all egress traffic` (lines 234–249, including the caution that a
  default deny-all egress policy also blocks DNS) at the frozen revision
- Ground truth: a default-deny egress policy selects every Pod in the
  namespace, so the application can neither resolve names nor reach its
  dependency. A successful repair allows egress only to the dependency on its
  port and to the cluster DNS service on UDP/TCP 53, restoring name resolution
  and dependency access while keeping other egress denied.
