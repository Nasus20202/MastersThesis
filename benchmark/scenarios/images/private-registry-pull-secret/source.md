# Image from an authenticated registry provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/pull-image-private-registry.md`
- Git blob SHA: `dd2c31eaa4053c7c6827a34435999b7fc87fb531`
- Relevant section: `Create a Secret by providing credentials on the command line` / `Create a Pod that uses your Secret`
- Ground truth: the image lives in an authenticated registry, so the workload
  needs a `kubernetes.io/dockerconfigjson` Secret and an `imagePullSecrets`
  reference; without it the kubelet fails the pull with `ImagePullBackOff`. A
  successful repair creates the credential Secret, references it from the
  Deployment, and brings all three replicas to a Ready state.
