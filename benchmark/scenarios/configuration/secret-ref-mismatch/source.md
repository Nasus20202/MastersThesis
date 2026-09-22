# Pod references a missing Secret key provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md`
- Git blob SHA: `430bc1b0dde00c2f6011762fc938e4b4664fb169`
- Relevant section: `Define container environment variables using Secret data` /
  `Define a container environment variable with data from a single Secret`
  (lines 243–282 at the frozen revision)
- Ground truth: the container environment references a Secret key that does not
  exist, so the container cannot be created (`CreateContainerConfigError`). A
  successful repair makes the referenced Secret key hold the intended value
  (either by correcting the reference or supplying the key), after which the
  application runs.
