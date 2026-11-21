---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/nonroot-securitycontext/prompt/005.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: kubectl patch deploy app sets runAsUser/runAsGroup 1000 (conventional non-root UID); exec shows uid=1000 and report matches."
---

# nonroot-securitycontext · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A workload in the cluster is rejected and never starts. Make it run as a non-root user.

Documented in:

- content/en/docs/tasks/configure-pod-container/security-context.md: Set the security context for a Pod, Set the security context for a Container

## Write commands

1. `kubectl -n default patch deploy app --type=merge -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}}}}}'`

## Final report

The workload now runs as a non-root user.

## Diagnosis

`deployment/app` in namespace `default` was stuck at 0/1 available. Its pod `app-56c6cdc6d6-4lp9b` was in `CreateContainerConfigError`, with the kubelet repeatedly reporting:

```
Error: container has runAsNonRoot and image will run as root
```

The Deployment's pod template had `securityContext.runAsNonRoot: true` but **no `runAsUser`**. Since the `busybox:1.37.0` image's default user is UID 0, the kubelet had no non-root UID to run as and rejected the container before starting it — the container was never created, so it could never become ready. I confirmed there was no PodSecurity Admission label on `default` and no PodSecurityPolicy, so the deployment's own `runAsNonRoot` field was the sole (and legitimate) enforcement of the non-root requirement.

## Change

Smallest justified fix: keep `runAsNonRoot: true` and pin an explicit non-root UID on the owning Deployment, so no unrelated settings were altered (image, command, args, strategy, replicas all untouched):

```
securityContext: {runAsNonRoot: true, runAsUser: 1000, runAsGroup: 1000}
```

## Verification

- **Original symptom gone:** new pod `app-77c6996468-mh2km` is `1/1 Running`, 0 restarts. Events show `Pulled → Created → Started` with **no** `Failed`/`runAsNonRoot` warnings on the new pod.
- **Actually non-root (not just accepted):** `kubectl exec ... id -u` → `uid=1000`, `groups=1000`; `ps` shows `PID 1 1000 sleep 3600`, i.e. the main process itself runs as UID 1000, and it was re-checked 17s later and still alive.
- **Controller counts match:** spec `replicas=1`, `updated=1`, `ready=1`, `available=1`, no `unavailableReplicas`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- **Constraint preserved:** effective pod securityContext is `{"runAsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}` — the non-root requirement is still enforced, now satisfiable by the runtime.
