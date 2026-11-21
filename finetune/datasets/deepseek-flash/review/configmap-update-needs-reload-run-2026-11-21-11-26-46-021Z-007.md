---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/configmap-update-needs-reload/prompt/007.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: kubectl rollout restart deployment/app; outputs show 3 Pods report beta, ConfigMap still beta, image unchanged."
---

# configmap-update-needs-reload · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl -n default rollout restart deployment/app`

## Final report

## Root cause

Deployment `app` injects `CONFIG_MESSAGE` via `env.valueFrom.configMapKeyRef` (`app-config`/`message`). Kubernetes resolves that reference **only when the container is created**, so the three original Pods kept the `alpha` snapshot even though the ConfigMap had already been updated to `beta`. The persistent resource holding the stale state was the Deployment's Pod template/generation; the ConfigMap itself was already correct and needed no change.

## Change made (smallest justified)

Triggered one rolling update of the Deployment's Pod template so the controller recreated the Pods and they re-read the ConfigMap:

```
kubectl -n default rollout restart deployment/app
```

No other field was modified — image, replica count, env source and ConfigMap contents are all untouched.

## Verification (direct observations)

- **Symptom gone** — all 3 running Pods now report `beta`:
  - `pod/app-657f995fc7-d9z5d -> beta`, `...-n45jb -> beta`, `...-tkq7v -> beta`; uniq count = `3 beta`, zero `alpha`.
- **Controller reconciliation** — requested `3 | 3` current, `updated=3`, `ready=3`, `available=3` (all match).
- **Constraints/config preserved** — image still `nginx:1.31.5`; env still `CONFIG_MESSAGE=app-config/message`; ConfigMap `app-config.data.message` still `beta`.
- **Rollout converged** — new ReplicaSet `app-657f995fc7` at 3/3; old `app-845b8b8485` scaled to 0, no leftover stale Pods.

The running Pods now use the updated `beta` value, and the Deployment remains the persistent owner of that state for future replica recreation.
