---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/nonroot-securitycontext/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# nonroot-securitycontext · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A workload in the cluster is rejected and never starts. Make it run as a non-root user.

Documented in:

- content/en/docs/tasks/configure-pod-container/security-context.md: Set the security context for a Pod, Set the security context for a Container

## Write commands

1. `kubectl patch deployment app -n default --type=merge -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}}}}}' 2>`

## Final report

## Diagnosis

The workload is `deployment/app` in namespace `default` (image `busybox:1.37.0`, command `sh -c "exec sleep 3600"`).

Observed evidence:

- Pod `app-56c6cdc6d6-f7sjf` was stuck in `CreateContainerConfigError`, never `Running`.
- Kubelet event, repeated 19 times: `Error: container has runAsNonRoot and image will run as root (pod: "app-56c6cdc6d6-f7sjf_default(...)", container: app)`.
- Deployment pod template had `securityContext: {runAsNonRoot: true}` with **no** `runAsUser`.
- The `busybox:1.37.0` image declares no non-root `USER`, so it defaults to UID 0. With `runAsNonRoot: true` but no UID given, the kubelet cannot satisfy the constraint and refuses to create the container.

The cause was a half-configured non-root restriction: the _check_ was enabled without the _identity_ that would satisfy it. The prior ReplicaSet `app-787bf6bd59` had an empty `securityContext` (running as root), so simply reverting would have produced a running pod that violates the stated requirement to run as non-root. The namespace `default` has no Pod Security Admission labels, so no external policy forced either choice.

## Change

Single merge patch to the owning Deployment, keeping the non-root restriction intact and supplying the missing identity:

```
spec.template.spec.securityContext: {runAsNonRoot: true, runAsUser: 1000, runAsGroup: 1000}
```

No other fields (image, command, args, strategy, replicas, resources) were modified.

## Verification

Original symptom gone — the container is created and running instead of being rejected:

- Deployment rollout: `deployment "app" successfully rolled out`; new ReplicaSet `app-77c6996468`, old failing pod deleted.
- Events for the new pod: `Created` → `Started`, with no `Failed` / `runAsNonRoot` error.
- Pod `app-77c6996468-hbq9t`: `1/1 Running`, `restarts=0`, `ready=true`, stable after a 20s settle window.
- Controller counts match the request: `requested=1 updated=1 ready=1 available=1` (no unavailable replicas).

Non-root outcome proven by inspecting the live container, not just the spec:

- `kubectl exec ... -- id` → `uid=1000 gid=1000 groups=1000`
- `id -u` → `1000` (i.e. not 0)
- Effective pod `securityContext` → `{"runAsGroup":1000,"runAsNonRoot":true,"runAsUser":1000}`

The workload now runs with UID 1000 while `runAsNonRoot: true` remains enforced, so the non-root requirement holds both in configuration and at runtime.
