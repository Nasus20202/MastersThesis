---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/nonroot-securitycontext/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# nonroot-securitycontext · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A workload in the cluster is rejected and never starts. Make it run as a non-root user.

Documented in:

- content/en/docs/tasks/configure-pod-container/security-context.md: Set the security context for a Pod, Set the security context for a Container

## Write commands

1. `kubectl -n default patch deploy app --type=merge -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}}}}}' 2>`

## Final report

The workload now starts and runs as a non-root user.

**Root cause**

The `default/app` Deployment's pod template set `securityContext.runAsNonRoot: true` but declared no `runAsUser`. The image `busybox:1.37.0` defaults to uid 0, so the kubelet refused to create the container. The pod sat in `Pending` with `CreateContainerConfigError`, and the event repeated 25+ times:

```
Warning  Failed  spec.containers{app}: Error: container has runAsNonRoot and image will run as root
```

Deployment status showed `MinimumReplicasUnavailable` with `0/1` ready.

**Change made** (one persistent object, the owning Deployment, minimal edit — `runAsNonRoot: true` preserved):

```
kubectl -n default patch deploy app --type=merge -p \
  '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}}}}}'
```

**Verification (fresh observations)**

| Check                                                              | Result                                                                                                                      |
| ------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------- |
| Deployment replicas: desired / updated / ready / available         | `1 / 1 / 1 / 1` (`kubectl get deploy` → `READY 1/1 AVAILABLE 1`)                                                            |
| Pod state                                                          | `app-77c6996468-cxnzs` `Running`, `ready=true`, `restarts=0`, container state `running` since 12:03:43                      |
| Effective identity inside the container (`kubectl exec ... -- id`) | `uid=1000 gid=1000 groups=1000`                                                                                             |
| Process owner (`ps`)                                               | PID 1 `sleep 3600` running as user `1000` (not root)                                                                        |
| Applied securityContext on the pod                                 | `{"runAsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}`                                                                  |
| Recurrence of the failure                                          | No new `Failed`/`CreateContainerConfigError` events; the last one (113s ago) belongs to the old, now-deleted ReplicaSet pod |

The non-root requirement is still enforced (`runAsNonRoot: true` kept), and the container is confirmed to execute as uid 1000 rather than merely being admitted.
