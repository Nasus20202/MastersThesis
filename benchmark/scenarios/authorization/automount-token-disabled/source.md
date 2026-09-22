# Service account token not mounted provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/configure-service-account.md`
- Git blob SHA: `ee8ce93e5145584610ed0bd35d1ab99aca9d9d8a`
- Relevant sections: `Opt out of API credential automounting` (lines 67–101) and
  `ServiceAccount token volume projection` (lines 381–483) at the frozen
  revision
- Ground truth: `automountServiceAccountToken: false` means the Pod gets no API
  credentials, so the application's call to the API server fails and the
  container restarts. A successful repair mounts the token again (or uses a
  projected token) while RBAC continues to allow the call.
