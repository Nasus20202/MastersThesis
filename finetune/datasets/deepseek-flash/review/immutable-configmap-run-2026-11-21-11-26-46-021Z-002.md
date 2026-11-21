---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/immutable-configmap/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# immutable-configmap · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

The application must run with the configuration value v2 but currently reports v1. Update the application configuration to v2 and make the running Pods use it.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Immutable ConfigMaps

## Write commands

1. `kubectl delete configmap app-config -n default --wait=true`
2. `kubectl create -f app-config-v2.yaml`
3. `kubectl rollout restart deployment/app -n default`

## Final report

Done — the app now runs with `v2` in the live Pods.

## Root cause

`deployment/app` injects `CONFIG_MESSAGE` from `configMapKeyRef{name: app-config, key: message}`. That ConfigMap held `message: v1` **and was marked `immutable: true`**, so an in-place `kubectl patch/edit` of the data is rejected by the API server — a new ConfigMap object (same name/key) was required.

## What I changed

1. Recreated the ConfigMap with the same name/key so the Deployment's existing reference stayed valid:
   - deleted `default/app-config` (old UID `c6d1379c…`, `message: v1`)
   - created `default/app-config` (new UID `d3c0248e…`) with `data.message: v2`, keeping `immutable: true`
2. `kubectl rollout restart deployment/app` — this was needed because env vars from `configMapKeyRef` are resolved only at container start, so the existing Pods would have kept `v1` forever.

## Verification (fresh observations)

- ConfigMap: `kubectl get cm app-config -n default -o jsonpath={.data.message}` → `v2`.
- Every running Pod reads v2 (`kubectl exec … printenv CONFIG_MESSAGE`):
  - `app-8487884bb4-bmcvn -> v2`
  - `app-8487884bb4-gsr9s -> v2`
  - `app-8487884bb4-swnwq -> v2`
- Rollout complete and replica counts match the requested count: `spec.replicas=3, updated=3, ready=3, available=3`, `generation=2, observedGeneration=2`; old ReplicaSet `app-845b8b8485` scaled to `0/0/0`.
- Constraints preserved: image still `nginx:1.31.5`, 3 replicas, and the env source is unchanged — still `configMapKeyRef{name: app-config, key: message}` (no other settings touched).
