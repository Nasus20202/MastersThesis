# Pod stuck Terminating because of a finalizer provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/overview/working-with-objects/finalizers.md`
- Git blob SHA: `70cc66c12b0201026ae220a9196e73b8a62177ad`
- Relevant section: `How finalizers work` (lines 20–62 at the frozen revision)
- Ground truth: a finalizer carried by a standalone Pod makes the API server
  set `metadata.deletionTimestamp` and keep the object, while no controller
  removes the finalizer, so the Pod stays Terminating indefinitely. A
  successful repair removes the finalizer so the object can be deleted and
  creates a healthy replacement Pod named app.
