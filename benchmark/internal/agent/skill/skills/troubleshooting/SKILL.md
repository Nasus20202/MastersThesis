---
name: troubleshooting
description: General diagnosis and repair workflow for structured troubleshooting.
---

# Troubleshooting

Diagnose from the task's evidence, change the resource that owns the faulty state, and verify recovery before reporting success.

## 1. Scope and diagnose

Start with resource names and namespaces supplied by the task. If they are missing, discover them with the narrowest useful query and expand the search only when evidence requires it. Avoid repeating cluster-wide listings. Check current state, relevant events or logs, and the owning resource.

Load the relevant supporting guidance before relying on it:

- Load the `kubectl` skill before forming a nontrivial mutation or wait command.
- Load the `kubernetes` skill when diagnosis depends on controller ownership, reconciliation, or resource status.
- For RBAC repairs, load the `authorization.md` reference before selecting subjects, verbs, resources, or scope.

Trace the symptom to the first resource whose observed state conflicts with the intended state. Prefer a focused observation over unrelated reads. Repair the owning resource rather than a generated Pod or other transient object.

## 2. Repair

Make one minimal declarative change to the owning resource and preserve unrelated fields and task constraints. Avoid interactive `kubectl edit`. Do not delete or restart a controller-managed Pod as a substitute for repairing its owner. For RBAC, grant only the required verbs on the required resources and namespace; avoid broad listing or Secret permissions unless the task evidence requires them.

## 3. Verify

Use fresh observations to confirm both that the change took effect and that the requested outcome and original constraints hold. In this benchmark, tool calls have a 60-second limit; bound waits to 30 seconds and do not start an unbounded watch. If verification fails, use the new evidence to diagnose further instead of stacking speculative changes.

For a Deployment, compare desired replicas (`spec.replicas`) with both `status.updatedReplicas` and `status.readyReplicas`. Do not report full recovery until both status counts equal the desired count. A successful patch or ready old replicas alone does not establish that the rollout completed.
