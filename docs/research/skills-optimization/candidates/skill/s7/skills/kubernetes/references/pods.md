# Kubernetes Pods

A Pod is the smallest deployable Kubernetes workload unit. Containers in a Pod share the Pod network namespace and can share mounted volumes. Each Pod has its own lifecycle and normally receives a cluster network address.

Init containers run before regular app containers and can gate application startup. App containers form the main workload. Ephemeral containers can be added for debugging and have different lifecycle guarantees from regular containers.

Pod phase (`Pending`, `Running`, `Succeeded`, `Failed`, `Unknown`) is a coarse summary. Conditions such as `PodScheduled`, `Initialized`, `ContainersReady`, and `Ready`, together with per-container waiting, running, and terminated states, provide more detail. A Pod can be `Running` while not `Ready`.

Startup, readiness, and liveness probes have different purposes. A startup probe delays the other probes until startup succeeds. Readiness controls whether a container is considered ready for serving. Liveness can cause the kubelet to restart a container after repeated failures.

Container status records restart counts and current or previous termination information. A waiting reason describes why a container has not started, while a terminated state records an ended container process including its exit code and reason.

Workload controllers own Pod templates. Editing or replacing a generated Pod does not change the controller template that determines future Pods.
