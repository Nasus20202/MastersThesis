# Roll back a bad release provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/workloads/controllers/deployment.md`
- Git blob SHA: `2ec9ba7c01996e278991f54886d4f9fd97c2c20b`
- Relevant section: `Rolling Back a Deployment`, in particular `Rolling Back to
a Previous Revision` (lines 378–634 at the frozen revision)
- Ground truth: a second revision of the Deployment changes the container image
  to a bad release that cannot become ready, so the rollout stalls with fewer
  available replicas than desired. A successful repair restores the Pod template
  of the last good revision (`nginx:1.31.5`) while keeping the Deployment's
  rollout history, for example with `kubectl rollout undo`. Deleting and
  recreating the Deployment restores availability but discards the history.
