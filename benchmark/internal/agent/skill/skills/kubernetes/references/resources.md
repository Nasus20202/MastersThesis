# Kubernetes Resources

CPU and memory requests and limits affect different stages of workload execution. Requests influence scheduling and resource reservation. Limits constrain runtime consumption: CPU limits are enforced through throttling, while exceeding a memory limit can terminate the container through an out-of-memory event.

Pod Quality of Service classes are derived from container CPU and memory requests and limits. `Guaranteed` requires every container to have CPU and memory requests and limits, with each request equal to its corresponding limit. `BestEffort` requires no CPU or memory requests or limits on any container. Other Pods are `Burstable`.

During node-pressure eviction, QoS class is one input together with whether a Pod is using more than its requests and the Pod's priority; QoS alone does not define a complete eviction order.

`LimitRange` can provide defaults and enforce minimum or maximum values within a namespace. `ResourceQuota` constrains aggregate namespace resource consumption and object counts. Quota admission can reject an object before scheduling occurs.

Changing a request changes scheduling and reservation behavior. Changing a limit changes runtime enforcement. They should be treated as separate controls.