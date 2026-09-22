# StatefulSet stable per-Pod DNS provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/concepts/workloads/controllers/statefulset.md`
- Git blob SHA: `288a2beb54235b19d72ef5c712a230d4f3f501d3`
- Relevant section: `Stable Network ID` (lines 190–233 at the frozen revision)
- Ground truth: the StatefulSet names a governing Service in `serviceName`, but
  that headless Service does not exist, so the per-Pod DNS records
  `app-0.app.default.svc.cluster.local` and `app-1.app.default.svc.cluster.local`
  do not exist. A successful repair creates the `clusterIP: None` Service with a
  selector matching the StatefulSet Pods and the expected port, which restores
  the per-Pod DNS records.
