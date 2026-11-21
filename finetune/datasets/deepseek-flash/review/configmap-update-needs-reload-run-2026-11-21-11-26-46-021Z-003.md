---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/configmap-update-needs-reload/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# configmap-update-needs-reload · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl rollout restart deployment/app -n default 2>`

## Final report

Done. The symptom is gone and verified directly.

**Diagnosis**

- `configmap/app-config` in `default` already contained `message: beta` (the config update had landed).
- `deployment/app` injects that key as `CONFIG_MESSAGE` via `configMapKeyRef`. ConfigMap-sourced **env vars are resolved only at container creation**, so the already-running Pods kept the stale snapshot: `CONFIG_MESSAGE=alpha`.
- The Deployment spec itself was already correct (`CONFIG_MESSAGE=app-config/message`), so no config change was needed — the Pods simply had to be recreated by their persistent owner. Deleting Pods alone would have been transient (the ReplicaSet would recreate them from the same unchanged template), so I changed the owning Deployment instead.

**Change made (only one)**

- `kubectl rollout restart deployment/app -n default` → added a `restartedAt` pod-template annotation, producing a new ReplicaSet. Nothing else was modified: replicas still `3`, image still `nginx:1.31.5`, env source still `app-config/message`, `app-config` data still `beta`.

**Verification (fresh observations)**

- In-Pod value via `printenv CONFIG_MESSAGE` on every Pod: `app-7885df88f5-hj8wn -> beta`, `…-ndbxb -> beta`, `…-nf9l6 -> beta`.
- Controller convergence: `3 requested, 3 updated, 3 ready, 3 available`, `observedGeneration=2` = `generation=2`, revision `2`.
- Old ReplicaSet `app-845b8b8485` scaled to `0/0/0`; new `app-7885df88f5` at `3/3/3`; all Pods `1/1 Running`, 0 restarts.
- No new Warning events (the three listed `FailedScheduling` entries are ~2m54s old, from initial cluster bring-up, and did not recur).
