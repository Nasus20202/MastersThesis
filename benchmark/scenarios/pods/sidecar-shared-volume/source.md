# Serve a page written by a sidecar through a shared volume provenance

This file is evaluator-only metadata. The benchmark runner exposes only the
`task` field from `scenario.yaml` to the model.

- Corpus: `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- Source path: `content/en/docs/tasks/access-application-cluster/communicate-containers-same-pod-shared-volume.md`
- Git blob SHA: `00da802420f3cfc2922ab66321dc0f60b09c14ec`
- Relevant section: `Creating a Pod that runs two Containers` (lines 20–112 at
  the frozen revision)
- Ground truth: the web container serves an `emptyDir` that contains nothing,
  so the root path returns no page. A successful repair adds a second container
  that writes an index page into the same `emptyDir` mounted by both containers,
  so the web server serves the page and both containers are Ready.
