---
name: scheduling
description: "Placing or repairing Pending and unschedulable Pods and DaemonSets: selectors, affinity, taints, tolerations and capacity."
---

# Scheduling

A Pending Pod is waiting on the scheduler. Read the scheduler's reason before changing anything, then remove only the constraint that blocks placement.

## Read the reason

```bash
kubectl describe pod <pod> -n <namespace>
```

The `FailedScheduling` event names the cause: no node matches the selector or affinity, no node tolerates the taints, insufficient capacity, or a volume that cannot attach. Act on the named cause.

## Repair by cause

- **Node selector or node affinity.** Compare `spec.nodeSelector` and the affinity rules with the actual node labels. Correct the label value the workload needs; adding a label to a node is allowed when the task asks for it. Do not delete a required selector to make the Pod fit.
- **Taints and tolerations.** To place a Pod on a tainted node, add the matching `toleration` to the workload's Pod template (`spec.template`). Never remove or edit the node taint to free space. A DaemonSet needs the toleration in its own `spec.template` like any other workload.
- **Anti-affinity.** A required `podAntiAffinity` with no eligible second node stays Pending. Check the topology key and rule against the node set instead of removing the rule.
- **Capacity.** Compare the Pod's `requests` with the node's allocatable and other Pods' requests. Reduce requests only if the task allows it and a stated floor is respected; never remove a limit to fit.
- **Volume attachment.** A Pod can be Pending because its PersistentVolumeClaim or volume is not attachable in the zone. Resolve the volume side (see the `storage` skill) rather than rescheduling blindly.

## Verify

The Pod is scheduled to an intended node and reaches Ready. Node labels and taints are unchanged, and for a DaemonSet the desired number of nodes has a Ready Pod. Read the fresh `spec.nodeName` and `status.conditions` instead of assuming the rollout succeeded.
