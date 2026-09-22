# NetworkPolicy blocks required traffic provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/network-policies.md`
- Git blob SHA: `22f3515ce15eb97378d7b5bf70c5a8cfab2fdb77`
- Relevant sections: `The NetworkPolicy resource` (lines 83–141) and `Default
deny all ingress traffic` (lines 214–223) at the frozen revision
- Ground truth: a default-deny ingress policy selects every Pod in the
  namespace, so the previously working client is blocked. A successful repair
  adds an ingress rule that allows only the intended client label on the server
  port, so that client is restored while all other clients remain denied.
