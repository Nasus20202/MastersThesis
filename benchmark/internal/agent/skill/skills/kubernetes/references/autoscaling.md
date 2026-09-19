# Kubernetes Autoscaling

Autoscalers act on different signals and targets. The HorizontalPodAutoscaler changes workload replica counts from resource, custom, or external metrics according to configured targets and behavior policies. Its decisions depend on the availability and freshness of the metrics pipeline.

The VerticalPodAutoscaler can recommend or update container resource requests, and can also manage limits depending on its policy. Applying updated resources can require Pods to be recreated. Horizontal and vertical autoscaling can interact, especially when they react to the same resource signal, so their controlled dimensions should be chosen deliberately.

Node autoscaling components such as Cluster Autoscaler change cluster capacity rather than workload replica counts. They react to scheduling demand and provider or node-group constraints, and scale-down must account for whether workloads can be moved safely.

Pod autoscaling changes workload demand; node autoscaling changes available supply. Quotas, placement constraints, disruption policy, and provisioner limits remain independent constraints.