---
name: kubernetes
description: Reason about Kubernetes API objects, controllers, workloads, networking, storage, scheduling, security and cluster operations from observed state.
metadata:
  version: "1.1"
  domain: Kubernetes
  corpus: kubernetes/website
  corpus_revision: ea639c1d22a60365d07b78692b1b1a2eb866bd15
---

# Kubernetes

Use this skill to build an evidence-based model of Kubernetes state. Start with
object identity and scope, then compare desired state with observed status and
conditions. Trace ownership and references between API objects, and locate the
failure at the relevant layer: API or admission, controller, scheduling, node
execution, container lifecycle, networking, storage or policy.

Controllers reconcile asynchronously, and Pods are execution units rather than
the durable source of workload configuration. Treat namespaces, labels,
selectors, owner references, ports, identities and configuration references as
relationships to verify, not values to assume. Make changes at the authoritative
boundary, preserve unrelated settings and access scope, and verify the result
with fresh observations.

Load one or more focused references only when their subjects are relevant. For
cross-layer problems, load the smallest set that covers the observed path; do
not load the entire reference set by default.

| Area          | Reference          | Load when                                                           |
| ------------- | ------------------ | ------------------------------------------------------------------- |
| Foundation    | `api.md`           | Inspecting object identity, scope, metadata or API behaviour.       |
| Foundation    | `architecture.md`  | Localizing a failure across control-plane and node components.      |
| Workloads     | `workloads.md`     | Following controllers, ownership, replicas or rollouts.             |
| Execution     | `pods.md`          | Interpreting Pod lifecycle, containers, probes, logs or volumes.    |
| Execution     | `nodes.md`         | Checking node health, taints, cordons or kubelet/runtime state.     |
| Scheduling    | `scheduling.md`    | Explaining placement, Pending state or resource fit.                |
| Resources     | `resources.md`     | Interpreting requests, limits, quotas, QoS or eviction.             |
| Networking    | `services.md`      | Tracing selectors, ports, EndpointSlices or ready backends.         |
| Networking    | `networking.md`    | Checking DNS, NetworkPolicy, Ingress, Gateway or CNI paths.         |
| Storage       | `storage.md`       | Following PVC/PV binding, provisioning, attach or mount state.      |
| Configuration | `configuration.md` | Tracing ConfigMaps, Secrets or configuration propagation.           |
| Security      | `rbac.md`          | Checking identities, roles, bindings or authorization.              |
| Security      | `security.md`      | Checking security contexts, admission or Pod Security Admission.    |
| Policy        | `policies.md`      | Interpreting admission, disruption or namespace-level policies.     |
| Operations    | `observability.md` | Selecting and correlating events, logs, status or metrics.          |
| Operations    | `lifecycle.md`     | Explaining deletion, termination, finalizers or garbage collection. |
| Operations    | `autoscaling.md`   | Diagnosing HPA targets, metrics or replica convergence.             |
| Extensions    | `extensions.md`    | Working with CRDs, custom resources or operators.                   |
