# StatefulSet volume claim template cannot bind provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/workloads/controllers/statefulset.md`
- Git blob SHA: `288a2beb54235b19d72ef5c712a230d4f3f501d3`
- Relevant sections: `Components` / `Volume Claim Templates` (lines 144–154)
  and `Stable Storage` (lines 234–243 at the frozen revision)
- Ground truth: the `volumeClaimTemplates` entry requests a StorageClass
  (`fast-ssd`) with no provisioner and no matching PersistentVolume, so each
  generated claim stays Pending and the StatefulSet Pods never start. A
  successful repair supplies the class or corrects the claim template, after
  which the claims bind and the Pods become Ready.
