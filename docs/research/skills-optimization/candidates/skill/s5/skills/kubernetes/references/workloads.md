# Kubernetes Workloads

Workload controllers own Pod templates and reconcile children from them. The ownership chain is recorded in `metadata.ownerReferences`:

```text
Deployment -> ReplicaSet -> Pod
StatefulSet -> Pod and volume claim
DaemonSet -> Pod on each eligible node
Job -> Pod
CronJob -> Job -> Pod
```

A Deployment manages rolling updates through ReplicaSets: each template change creates a new ReplicaSet revision while the old one scales down according to the update strategy and availability limits. `rollout history` lists these revisions. Desired, updated, available, and ready replica counts describe different stages of that convergence and normally differ during an update.

A StatefulSet gives each Pod a stable ordinal identity (`pod-0`, `pod-1`) with stable storage claims and ordered update semantics. Pods are not interchangeable: identity and claim names persist across rescheduling.

A DaemonSet places one Pod per eligible node, governed by node labels, taints, and affinity rather than a replica count. A Job runs Pods to completion and tracks successes and failures against a completion count; a CronJob creates Jobs on a schedule and is additionally governed by concurrency policy, deadlines, and history limits.

The durable object is the controller holding the template, not any generated child. A change to a Pod alone is overwritten at the next reconciliation; the template defines what returns.
