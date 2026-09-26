# Kubernetes Architecture

The control plane keeps the cluster's state and issues decisions; nodes run workloads. The API server is the only entry point for state: every read and write passes through it, backed by etcd. The scheduler assigns unscheduled Pods to nodes. Controller processes watch objects and drive them toward their declared spec. The cloud controller integrates provider-specific node, route, and load-balancer behavior where a provider exists.

Nodes run an agent (kubelet), a container runtime, and a network proxy (kube-proxy). The kubelet registers the node, reports capacity and conditions through heartbeats, and reconciles the Pods assigned to it: pulling images, mounting volumes, starting containers, and running probes. It does not schedule; it executes assignments made by the scheduler.

All components follow the same pattern: watch the API for relevant objects, compare desired state with observed state, act, and report back through status, conditions, and events. Control is therefore asynchronous and eventually consistent: an accepted write is only a request, and convergence is reported later through generations, conditions, and child objects.

In local clusters such as `kind`, the same components run containerized, often with a single control-plane node. The architecture is unchanged; only capacity, networking, and storage providers differ.
