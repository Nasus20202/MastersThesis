# Application cannot resolve its peer Service provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/dns-pod-service.md`
- Git blob SHA: `c05cd0b6f9df1e4c798f3273ba4bd52f0db2cf4b`
- Relevant sections: `Namespaces of Services` (lines 28–56) and `Services` /
  `A/AAAA records` (lines 70–85) at the frozen revision
- Ground truth: the client container is configured with a peer host name that
  does not exist as a Service, so the query fails and the client never becomes
  Ready. A successful repair points the client at the resolvable Service name
  in the same namespace while keeping the two intended replicas.
