# Kubernetes observability

Use several evidence types for different layers of the system:

- Object status and conditions describe what a controller or component reports;
  compare them with the relevant generation and timestamp.
- Events explain recent scheduling, admission, image, mount, probe and
  controller transitions, but are transient and may be incomplete.
- Current container logs show process output; previous-container logs are
  relevant after a restart. Logs do not establish readiness or service routing.
- Metrics can reveal resource pressure, replica targets and latency when the
  cluster exposes the required metrics API or monitoring system. First verify
  that the metric source exists and is current.
- Audit or component logs may explain API and control-plane decisions when they
  are exposed by the environment; do not assume access to them.

Correlate namespace, object name, UID, container, node and time across evidence.
Prefer a focused observation that distinguishes hypotheses over a large dump.
Use bounded collection and preserve exit status. A missing event, empty log or
successful command is not, by itself, evidence of health.
