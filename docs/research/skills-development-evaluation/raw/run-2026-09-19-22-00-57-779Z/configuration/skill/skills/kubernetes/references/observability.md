# Kubernetes Observability

Kubernetes exposes four signal families with different coverage and retention. Status fields and conditions are the current reconciled view: generations, replica counts, Pod phases and conditions, and serving state. They answer "what does the system believe now" and say nothing about history.

Events are recent control-plane explanations attached to objects: scheduling decisions, pull and mount outcomes, probe results, and controller actions, each with a timestamp, reason, and count. They expire and are not an audit log; correlation with Pod start times, restart counts, and rollout revisions is what makes them meaningful.

Logs are process output: `kubectl logs` for the current run, `--previous` for the last terminated run, per container. They describe what the process did, never why the control plane acted. Metrics (resource usage, custom metrics, API-server and kubelet indicators) quantify behavior over time but require a metrics pipeline to exist.

A successful API response only proves the request was accepted. Desired state, controller conditions, events, logs, and functional client observations each cover a different layer; agreement across layers is the evidence, and any single layer can look healthy while another carries the failure.
