# Expose a Service through an Ingress path provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/ingress.md`
- Git blob SHA: `6324ee2e50f99a9fb9d2bbdcb900f3d91927850c`
- Relevant sections: `Ingress rules` (lines 119–138), `Path types` (lines
  183–206) and `Ingress backed by a single Service` (lines 392–419) at the
  frozen revision
- Ground truth: the namespace starts with a healthy Service and no Ingress. A
  successful repair creates an Ingress with the Traefik ingress class whose
  host/path rule forwards to the Service, so a request with the matching Host
  and path is proxied to the application.
