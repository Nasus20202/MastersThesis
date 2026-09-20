# Kubernetes Lifecycle

Rollouts, deletion, and disruption are asynchronous operations with controller-specific semantics. Deployments normally use `RollingUpdate`, bounded by surge and unavailable settings, while other controllers such as StatefulSets and DaemonSets have their own update strategies. Rollout status is reflected through generations, conditions, revisions, and replica counts over time.

Deletion normally begins by setting a deletion timestamp and then allowing Kubernetes and controllers to complete graceful termination and cleanup. Finalizers can deliberately keep an object present until cleanup is complete. Node unavailability, storage detach work, dependent resources, or controller failures can also delay termination.

Owner references and propagation policy influence whether dependent objects are removed with their owner. Removing a generated child does not change the desired state stored by its owner.

PodDisruptionBudgets limit voluntary disruptions for selected Pods. They affect operations such as eviction and node drain, but do not prevent all involuntary workload failures.
