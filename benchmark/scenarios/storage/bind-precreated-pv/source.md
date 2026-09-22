# Bind a claim to a pre-created PersistentVolume provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/storage/persistent-volumes.md`
- Git blob SHA: `ea0880a3f91d33553f6630a5f350330429a41cd1`
- Relevant sections: PersistentVolume and PersistentVolumeClaim concepts
  (lines 37–80) and `PersistentVolumes typed hostPath` (lines 1007–1011 at the
  frozen revision)
- Ground truth: a manually provisioned PersistentVolume is available and must
  be matched by a PersistentVolumeClaim with the same storage class, access
  mode and capacity, then consumed by a Pod. A successful repair delivers a
  Bound claim and a Ready Pod mounting it.
