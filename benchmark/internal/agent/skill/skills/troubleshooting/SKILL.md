---
name: troubleshooting
description: General inspect-diagnose-repair-verify workflow. Load when the troubleshooting process is useful; use the kubernetes skill for subsystem semantics and focused references, and kubectl for command syntax.
---

# Troubleshooting

Follow this loop: inspect, identify the owner and cause, make the smallest durable change, and verify the outcome and constraints.

## 1. Inspect

Capture the symptom, the affected resources, and the constraints that must survive the repair (replicas, permissions, data, availability, configuration). Record object state, conditions, events, logs, and owners. Use the `kubectl` skill for command details and the `kubernetes` skill for Kubernetes semantics when needed. A command that succeeds can still report a wrong or stale object.

## 2. Identify the owner and cause

Trace the symptom through the dependency chain and stop at the first layer whose evidence disagrees with its expected state. Prefer one observation that eliminates a cause over many unrelated reads. Repair the resource that owns the faulty state, not a generated object it creates.

When the evidence identifies a Kubernetes subsystem, load the `kubernetes` skill if it has not already been loaded. Check its focused references and load the exact matching reference before a specialized repair when semantics, ownership, or constraints matter. Do not route domain questions to the `kubectl` skill and do not load unrelated references mechanically.

## 3. Make the smallest durable change

Change the owning resource with one minimal change that preserves unrelated settings and the constraints from step 1. Review the diff before applying. Do not treat deleting or restarting a transient object as a lasting fix.

## 4. Verify the outcome and constraints

Reinspect with fresh observations: first that the change took effect, then that the requested outcome holds and the original constraints survive. Bound waits with an explicit timeout. If verification fails, return to step 1 with the new evidence instead of stacking another change.
