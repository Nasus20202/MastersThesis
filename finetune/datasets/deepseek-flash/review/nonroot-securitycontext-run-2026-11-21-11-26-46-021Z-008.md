---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/nonroot-securitycontext/prompt/008.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: kubectl patch deployment app sets runAsUser 1000; exec id shows uid=1000 and kubelet event x21 matches report."
---

# nonroot-securitycontext · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

A workload in the cluster is rejected and never starts. Make it run as a non-root user.

Documented in:

- content/en/docs/tasks/configure-pod-container/security-context.md: Set the security context for a Pod, Set the security context for a Container

## Write commands

1. `kubectl -n default patch deployment app --type=merge -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}}}}}' 2>`

## Final report

## Diagnosis

The `default` namespace Deployment `app` had a pod stuck in `CreateContainerConfigError` that never started. The kubelet event named the exact cause:

```
Warning  Failed  Error: container has runAsNonRoot and image will run as root
        (pod: "app-56c6cdc6d6-5249g_default", container: app)
```

The pod template's securityContext declared `runAsNonRoot: true` but specified no UID. The image `busybox:1.37.0` has no non-root `USER` in its config, so the runtime would run it as UID 0. `runAsNonRoot: true` alone is only a validator — it cannot pick a user — so the kubelet rejected the container before creating it (event repeated x21).

## Change made

One minimal patch to the Deployment, which owns the faulty pod, keeping the existing security intent (`runAsNonRoot: true`) intact:

```
kubectl -n default patch deployment app --type=merge \
  -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}}}}}'
```

No other fields (image, command, replicas, strategy) were touched, and no namespace policy needed to change — `default` has no Pod Security Admission enforcement labels and no ResourceQuota/LimitRange.

## Verified outcome

- **Symptom gone:** the failing pod `app-56c6cdc6d6-5249g` was replaced; zero pods with that ReplicaSet hash remain, and no further `runAsNonRoot`/`Failed` events were logged.
- **Controller replica counts match:** `deployment.apps/app` reports `READY 1/1, UP-TO-DATE 1, AVAILABLE 1`; `rollout status` returned success within the timeout.
- **Running, not restarting:** `app-77c6996468-wnr9k` is `1/1 Running`, `ready=true`, `restarts=0`, container state `running`.
- **Actually non-root (direct evidence):** `kubectl exec … -- id` returns `uid=1000 gid=1000 groups=1000`; effective pod securityContext is `{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000}`.

The workload now runs as a non-root user (UID 1000) while still enforcing `runAsNonRoot`.
