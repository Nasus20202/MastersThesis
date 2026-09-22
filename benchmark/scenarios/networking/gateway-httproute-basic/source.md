# Expose a Service with Gateway API provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/gateway.md`
- Git blob SHA: `5948d74a083200fdcd50cb58a5f4a02549b7b5ae`
- Relevant sections: `Gateway` (lines 91–130) and `HTTPRoute` (lines 131–167)
  at the frozen revision
- Ground truth: the namespace starts with a healthy Service and no Gateway API
  objects. A successful repair creates a Gateway with the `traefik` class and a
  listener whose port matches the Traefik `web` entrypoint, plus an HTTPRoute
  that attaches to that Gateway and forwards matching host/path requests to the
  Service, so the request is finally proxied to the application.
