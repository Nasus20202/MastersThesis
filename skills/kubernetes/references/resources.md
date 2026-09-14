# Kubernetes resources and eviction

Resource requests influence scheduling; limits constrain runtime usage. Always
inspect the effective Pod specification because a LimitRange or another default
may have added values that are not obvious in the workload template.

- Compare CPU, memory and extended-resource requests with node allocatable
  capacity and current placement. Replica count does not add capacity.
- Interpret limits with the container state and node evidence. `OOMKilled`,
  node eviction and application-level failures are different causes even when
  all appear as a restart.
- Check namespace `ResourceQuota` and `LimitRange` when creation is rejected or
  defaults appear unexpectedly. Quota admission can fail before a controller
  creates a Pod.
- Use QoS classification and pressure evidence to explain eviction risk, not as
  a substitute for checking the actual event and container state.
- Preserve requests, limits and namespace policy unless the evidence identifies
  a resource mismatch. Raising a limit can move the failure to a node or make
  scheduling impossible.

Re-check admission, placement, container state and workload convergence after a
resource change. Keep unrelated containers and resource settings unchanged.
