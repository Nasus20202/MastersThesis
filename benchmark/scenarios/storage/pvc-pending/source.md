# PersistentVolumeClaim stuck Pending provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/storage/persistent-volumes.md`
- Git blob SHA: `ea0880a3f91d33553f6630a5f350330429a41cd1`
- Relevant sections: `Lifecycle of a volume and claim` / `Binding` (lines
  60–110) and `PersistentVolumeClaims` / `Class` (lines 867–897 at the frozen
  revision)
- Ground truth: the claim requests a StorageClass (`fast-ssd`) that has no
  matching PersistentVolume and no provisioner, so it remains Pending and the
  consumer Pod cannot start. A successful repair provides the requested class or
  delivers a bound claim, after which the consumer Pod becomes Ready.
