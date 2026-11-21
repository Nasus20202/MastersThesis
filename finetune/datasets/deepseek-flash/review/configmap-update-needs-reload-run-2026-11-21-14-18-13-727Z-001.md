---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-14-18-13-727Z/configmap-update-needs-reload/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# configmap-update-needs-reload · deepseek-flash · run-2026-11-21-14-18-13-727Z · attempt 1

## Task

The application configuration was updated to beta, but the running Pods still report alpha. Make the running Pods use the updated value.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Using ConfigMaps > Using ConfigMaps as files from a Pod > Mounted ConfigMaps are updated automatically

## Write commands

1. `kubectl rollout restart deployment/app -n default`

## Final report

## Root cause

`app-config` already contained the updated value (`message: beta`), and the Deployment `app` consumes it as `CONFIG_MESSAGE` via `configMapKeyRef`. Environment variables sourced from a ConfigMap are resolved **only when the container starts**, and the Deployment's pod template was unchanged (`generation=1`, single ReplicaSet `app-845b8b8485`), so nothing ever replaced the running Pods — all three still showed `CONFIG_MESSAGE=alpha`.

## Change made

Rolled the owning persistent resource, the `app` Deployment (`kubectl rollout restart deployment/app -n default`), which stamps `kubectl.kubernetes.io/restartedAt: 2026-11-21T14:19:12Z` on the pod template and creates a new ReplicaSet:

- Deployment `app` generation `1` → `2`, new ReplicaSet `app-f47cb9bf9`; old RS `app-845b8b8485` scaled to 0 desired/0 current.
- Nothing else touched: ConfigMap `app-config` still `message: beta`, replica count still `3`, image still `nginx:1.31.5`, rolling-update strategy unchanged.

## Verified outcome

- Rollout status: `deployment "app" successfully rolled out`; replicas `spec=3, replicas=3, updated=3, ready=3, available=3` (match).
- All three new Pods `Running 1/1`, `RESTARTS 0`.
- Direct in-pod observation of the reported symptom, on every Pod:
  - `app-f47cb9bf9-8ntcq` → `CONFIG_MESSAGE=beta`
  - `app-f47cb9bf9-lgsq2` → `CONFIG_MESSAGE=beta`
  - `app-f47cb9bf9-xnsff` → `CONFIG_MESSAGE=beta`

The running Pods now report `beta`, matching the updated ConfigMap.
