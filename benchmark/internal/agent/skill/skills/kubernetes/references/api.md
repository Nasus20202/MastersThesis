# Kubernetes API Objects

Kubernetes resources are represented as API objects with common fields such as `apiVersion`, `kind`, and `metadata`, plus type-specific desired and observed state. Many workload resources use `spec` for desired state and `status` for observations reported by controllers or kubelets.

`metadata` carries identity and relationship fields such as `name`, `namespace`, `uid`, labels, annotations, owner references, generation, and finalizers. Names are unique within the scope defined for a resource type; UIDs distinguish object instances across creation and deletion. Labels support selection and grouping, while annotations carry non-identifying metadata.

API groups and versions identify resource schemas, for example `apps/v1` for Deployments and `v1` for core resources such as Pods and Services. Controllers can expose an `observedGeneration` or similar status field to indicate which desired-state generation has been processed.

Some resource types are namespaced and others are cluster-scoped. Whether one object may reference another object in a different namespace is defined by the specific API field; many built-in workload references, including Pod references to ConfigMaps and Secrets, are same-namespace only.

Subresources expose narrower operations with their own API and authorization semantics, for example `pods/log`, `pods/exec`, and `deployments/scale`.