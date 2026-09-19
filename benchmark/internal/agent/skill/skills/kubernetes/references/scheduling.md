# Kubernetes Scheduling

Scheduling assigns unscheduled Pods to nodes. The scheduler evaluates node eligibility and placement constraints such as effective Pod requests, node selectors, node affinity and anti-affinity, taints and tolerations, topology spread constraints, priority, and available allocatable capacity.

Requests participate in scheduling; resource limits do not create schedulable capacity. Init-container requests contribute to a Pod's effective scheduling request according to Kubernetes' resource calculation rules.

Node labels, taints, readiness, and unschedulable state affect placement. Affinity and topology rules can also make a Pod unschedulable even when raw CPU and memory capacity is available.

ResourceQuota and LimitRange are admission controls rather than scheduler filters. They can reject or alter an object before the scheduler sees it.

Scheduling ends when a Pod is bound to a node. Image pulling, volume attachment and mounting, container startup, and readiness happen afterward on the selected node. Scheduler events and Pod conditions expose placement decisions and failures.