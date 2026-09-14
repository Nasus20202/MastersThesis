# Kubernetes workloads

Workload controllers express desired state and create or manage lower-level
objects. A Deployment commonly manages a ReplicaSet, which manages Pods;
StatefulSet, DaemonSet, Job and CronJob use different ownership and rollout
semantics. Do not treat a Pod as the durable source of workload configuration.

- Deployment and ReplicaSet changes create revisions and converge through a
  rollout; compare desired, updated, available and ready replicas.
- StatefulSet manages stable identities and may coordinate ordering and
  persistent claims. Check its update policy and per-Pod identity before
  changing replicas or deleting a Pod.
- DaemonSet targets eligible nodes rather than a fixed replica count. Inspect
  node selectors, affinity, taints and tolerations when coverage is unexpected.
- Job and CronJob express completion and scheduling behavior, not continuous
  service availability. Check completions, parallelism, backoff, deadlines and
  the Jobs created by a CronJob.

When tracing a workload:

1. Identify the relevant controller and inspect its desired and observed
   status.
2. Follow `metadata.ownerReferences` to connect controllers, ReplicaSets and
   Pods.
3. Compare the controller's pod template with the actual Pod specification
   and conditions.
4. Use rollout status and controller events to distinguish convergence from a
   stalled or failed rollout.

An edit to a generated Pod may be replaced by its owner and is therefore not a
durable repair. Preserve update strategy, selectors and unrelated template
fields when changing the owning object.
