# Kubernetes nodes

A Node represents a place where a kubelet and container runtime can execute
Pods. A node being registered or `Ready` does not mean that every workload can
fit or that its network and storage paths are healthy.

- Check node conditions such as readiness, memory pressure, disk pressure and
  PID pressure, along with recent transitions and events.
- Compare `spec.unschedulable`, taints and labels with the Pod's tolerations,
  selectors and affinity. A cordon prevents new placement but does not stop
  existing Pods.
- Distinguish capacity from allocatable resources and from the requests already
  assigned to Pods. Resource fit is a scheduling question; runtime pressure or
  eviction is a node-execution question.
- For a placed Pod that is not running, inspect kubelet/runtime-related events,
  image access, volumes and container state before changing placement.
- Treat drain, deletion and runtime restarts as disruptive operations. Use them
  only when the evidence and task scope justify them, and verify the resulting
  node and workload state.

Preserve node labels, taints and unrelated workloads. A node-level repair should
not be used to conceal a malformed Pod specification or an application failure.
