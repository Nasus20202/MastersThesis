---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/configmap-update-needs-reload/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# configmap-update-needs-reload · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl rollout restart deployment/app`

## Final report

**Root cause:** `configMap/app-config` was updated to `message: beta`, but the Deployment `app` consumes it via `configMapKeyRef` as an environment variable. Env vars are injected only at container creation, so the pre-existing Pods (ReplicaSet `app-845b8b8485`) kept reporting `alpha` — no Pod-template change meant no rollout.

**Change made:** Triggered a rolling update of the owning Deployment (`kubectl rollout restart deployment/app`), which only added the pod-template annotation `kubectl.kubernetes.io/restartedAt: 2026-11-21T11:50:08Z` and let the Deployment recreate its Pods. Nothing else was altered: `replicas: 3`, image `nginx:1.31.5`, and the `CONFIG_MESSAGE` → `app-config/message` reference are all unchanged, and the ConfigMap still reads `beta`.

**Verified outcome (fresh observations):**

- Pods now run from new ReplicaSet `app-5dbbfbcb6b` (old `app-845b8b8485` scaled to 0), all 1/1 Running, 0 restarts.
- `printenv CONFIG_MESSAGE` in each of the three running Pods: `beta`, `beta`, `beta` — the reported symptom (alpha) is gone.
- Deployment status matches the requested spec exactly: `spec.replicas=3`, `updatedReplicas=3`, `readyReplicas=3`, `availableReplicas=3`, `generation=2` = `observedGeneration=2`.
- ConfigMap of record unchanged and still `message: beta`.
