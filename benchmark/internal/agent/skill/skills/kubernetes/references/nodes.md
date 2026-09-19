# Kubernetes Nodes

A Node object represents a worker machine. The kubelet on that machine registers or updates node state and reports capacity, allocatable resources, addresses, and health conditions.

`capacity` is the node's total resource inventory, while `allocatable` is the portion available for Pods after system reservations. Node conditions include `Ready`, `MemoryPressure`, `DiskPressure`, and `PIDPressure`. The node's unschedulable flag prevents ordinary new scheduling onto the node while leaving existing Pods in place.

The kubelet runs Pods that have been assigned to its node: it coordinates image pulling, volumes, container startup, probes, and status reporting through the configured container runtime and other node components.

Node labels, taints, allocatable capacity, and readiness influence scheduling. Node pressure and kubelet eviction can also affect Pods after scheduling, so placement and runtime node health should be treated as separate stages.

DaemonSets, system Pods, and other workloads consume allocatable resources just like application workloads.