# Kubernetes scheduling

The scheduler matches a Pod's requests and placement constraints to a node's
available allocatable capacity and policies. A replica count expresses how
many Pods are desired; it does not create node capacity or explain why a Pod
is Pending.

For a Pending Pod, inspect:

- Pod events and the scheduler's unschedulable reason;
- CPU and memory requests, plus other requested resources;
- node capacity and allocatable values and currently requested capacity;
- taints and tolerations, node selectors/affinity and other placement rules;
- priority, preemption and topology or anti-affinity constraints when present.

Distinguish scheduling admission from later container startup or readiness.
Do not treat `nodeName` assignment as normal scheduler placement. Change the
constraint or request only when the evidence identifies it as the cause, and
re-check events and placement after the scheduler has had time to converge.
