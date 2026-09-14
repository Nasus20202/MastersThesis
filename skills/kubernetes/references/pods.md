# Kubernetes Pods

A Pod is the scheduling and execution unit. Its containers share a network
namespace and declared volumes but have separate process and container state.
Inspect its phase and conditions, container states and restart information
together; a phase alone is not a complete diagnosis.

- Use `describe` and recent events for scheduling, admission, mount and probe
  information.
- Use current logs and, when a container restarted, the previous-container
  logs. Keep the container name and namespace explicit when more than one is
  present.
- Compare image, command, arguments, environment, mounts and probe settings
  with the owning template. Check referenced ConfigMaps, Secrets and volumes
  without exposing secret values unnecessarily.
- Account for init containers and container ordering when the application
  container has not started. Ephemeral containers are for investigation and do
  not repair the owning workload template.
- Treat readiness as a routing signal and liveness/startup probes as lifecycle
  signals; a running container is not necessarily ready to receive traffic.

Use the owner and the Pod specification to decide where a durable change
belongs. Verify both Pod convergence and the user-visible behavior afterward.
