# Container repeatedly killed by its memory limit provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/configure-pod-container/assign-memory-resource.md`
- Git blob SHA: `8bf3d4748f1a87bc10a4ab865e9f3ec22db34327`
- Relevant sections: `Specify a memory request and a memory limit` (lines 63–129)
  and `Exceed a Container's memory limit` (lines 130–235) at the frozen revision
- Ground truth: the container memory limit is below the working set of the
  application, so the kubelet terminates the container and it enters a restart
  loop. A successful repair raises the memory limit to a value the application
  fits in and sets a matching memory request, leaving the restart count stable.
