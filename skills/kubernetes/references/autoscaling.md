# Kubernetes autoscaling

An HPA reconciles a target workload's replica count from configured metrics and
constraints. It does not replace the workload controller and should not be
debugged by treating a current `replicas` value as the whole state.

- Verify the HPA's `scaleTargetRef`, namespace, min/max replicas and reported
  conditions and events. Confirm that the target exposes the expected scale
  subresource.
- Check current and desired replicas, metric type, target value, metric
  availability and the requests used by utilization metrics. A missing or stale
  metrics source is different from a target that is merely below or above its
  threshold.
- Account for stabilization, rate limits and other behavior settings before
  expecting an immediate change. Compare timestamps and fresh observations.
- Determine whether another controller or a human is also writing the replica
  field. Do not fight an HPA by editing the target workload's replicas directly.
- Treat VPA and external autoscalers as separate components when present; verify
  their ownership and supported behavior instead of assuming they are built in.

Verify the autoscaler status, target workload convergence and application-level
behavior after a change. Preserve configured bounds and unrelated metrics.
