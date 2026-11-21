---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/nonroot-securitycontext/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# nonroot-securitycontext · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

A workload in the cluster is rejected and never starts. Make it run as a non-root user.

Documented in:

- content/en/docs/tasks/configure-pod-container/security-context.md: Set the security context for a Pod, Set the security context for a Container

## Write commands

1. `kubectl -n default patch deployment app --type=merge -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}}}}}'`

## Final report

**Cause:** The `default/app` Deployment set `securityContext.runAsNonRoot: true` but specified no non-root user, while `busybox:1.37.0` defaults to UID 0. The kubelet therefore rejected the container: `Error: container has runAsNonRoot and image will run as root` (21 failed attempts, pod stuck in `CreateContainerConfigError`).

**Change (smallest justified, on the owning Deployment `default/app`):**

```yaml
spec.template.spec.securityContext:
  runAsNonRoot: true # kept
  runAsUser: 1000 # added
  runAsGroup: 1000 # added
```

No other fields (image, command/args, replicas, strategy) were modified; the `runAsNonRoot: true` constraint is preserved rather than weakened.

**Verification (fresh observations):**

- Pod `app-77c6996468-pwrql`: `1/1 Running`, no restarts; old failing pod `app-56c6cdc6d6-hh5jq` was removed by the rollout, and subsequent events show `Created`/`Started` with no further `runAsNonRoot` failures.
- Actual runtime identity inside the container: `uid=1000 gid=1000 groups=1000`, and `ps` shows `PID 1 USER 1000 COMMAND sleep 3600` — the workload process is genuinely non-root, not just admitted.
- Deployment replica counts all match the requested 1: `spec.replicas=1, updatedReplicas=1, readyReplicas=1, availableReplicas=1, unavailableReplicas=0`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.

The previously rejected workload now starts and runs as a non-root user (UID 1000) with the non-root enforcement intact.
