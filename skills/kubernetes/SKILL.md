---
name: kubernetes
description: Interpret Kubernetes resources, control loops, networking, scheduling, configuration and authorization from observed cluster state.
metadata:
  version: "1.0"
  domain: Kubernetes
  corpus: kubernetes/website
  corpus_revision: ea639c1d22a60365d07b78692b1b1a2eb866bd15
---

# Kubernetes

Use this skill to build an evidence-based model of Kubernetes state. API
objects express desired state; controllers reconcile it; Pods are execution
units whose status, events and logs show observed state. Namespaces, labels,
selectors, owner references, ports and identities are relationships to inspect,
not values to assume.

Distinguish the declarative object that owns a result from the Pods or other
objects it creates. Check status and conditions, recent events, dependencies,
configuration references, node placement and authorization according to what
the evidence makes relevant. Make changes at the owning boundary, preserve
least privilege and verify the resulting state with fresh observations.

Load a focused reference only when its subject is relevant:

| Reference | Load when |
| --- | --- |
| `workloads.md` | Following controllers, ownership, replicas or rollouts. |
| `pods.md` | Interpreting Pod lifecycle, containers, probes, logs or volumes. |
| `services.md` | Tracing service discovery, selectors, ports or EndpointSlices. |
| `scheduling.md` | Explaining placement, Pending state or resource fit. |
| `rbac.md` | Checking identities, roles, bindings or authorization. |
| `configuration.md` | Tracing ConfigMaps, Secrets or configuration propagation. |
