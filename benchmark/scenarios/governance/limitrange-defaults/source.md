# LimitRange provides default resources provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/policy/limit-range.md`
- Git blob SHA: `b4cbcdc40394a4ff9a00af2ba215442a0a928cbb`
- Relevant section: LimitRange defaults and admission behavior (lines 24–51 at
  the frozen revision)
- Ground truth: a LimitRange supplies default container requests and limits in
  the namespace. A successful repair deploys the two-replica workload without
  explicit resources and the Pods receive the defaults (`100m` requests,
  `200m` limits).
