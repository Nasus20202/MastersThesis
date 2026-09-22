# Gateway HTTPRoute does not attach provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/services-networking/gateway.md`
- Git blob SHA: `5948d74a083200fdcd50cb58a5f4a02549b7b5ae`
- Relevant section: `HTTPRoute` (lines 131–167 at the frozen revision)
- Ground truth: the Gateway is programmed and the HTTPRoute initially attaches
  through its `parentRefs`. The injected fault repoints the route at a parent
  Gateway that does not exist, so the route is no longer attached and requests
  through the Gateway stop reaching the Service. A successful repair restores a
  `parentRefs` entry that names the existing Gateway while keeping the host and
  path routing intact.
