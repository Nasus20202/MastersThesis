---
name: kubernetes
description: Kubernetes concepts for interpreting resources, controllers, status, and failure symptoms.
---

# Kubernetes

Use this skill for Kubernetes domain knowledge.

Kubernetes is a declarative system built around desired state and reconciliation. Reason about resources through their API objects, controllers, relationships, current status, events, and the components responsible for reconciling them.

Keep resource ownership and abstraction boundaries in mind. Higher-level controllers manage lower-level resources, configuration is often distributed across related objects, and observed symptoms may originate in another part of the system.

Load a reference with `load_reference` when more detail about a relevant area is useful. For command details use the `kubectl` skill; for a general diagnosis and repair process use the `troubleshooting` skill.

## References

- `api.md` — API objects, metadata, spec, status and API conventions.
- `architecture.md` — control plane, nodes and Kubernetes components.
- `workloads.md` — workload controllers and workload patterns.
- `pods.md` — Pods, containers, probes and lifecycle.
- `images.md` — image names, pull policy, registry credentials and pull failures.
- `nodes.md` — nodes, kubelet and node conditions.
- `scheduling.md` — scheduling, placement, affinity, taints and topology.
- `resources.md` — requests, limits, QoS and resource management.
- `networking.md` — Services, DNS, EndpointSlices, ingress and network policy.
- `storage.md` — volumes, PersistentVolumes, claims and StorageClasses.
- `configuration.md` — ConfigMaps, Secrets and workload configuration.
- `authorization.md` — ServiceAccounts, RBAC and authorization.
- `security.md` — workload and cluster security concepts.
- `policies.md` — admission, policies and resource governance.
- `observability.md` — status, events, logs, metrics and debugging signals.
- `lifecycle.md` — rollout, deletion, disruption and object lifecycle.
- `autoscaling.md` — workload and cluster autoscaling.
- `extensions.md` — CRDs, operators, admission webhooks and API extensions.
