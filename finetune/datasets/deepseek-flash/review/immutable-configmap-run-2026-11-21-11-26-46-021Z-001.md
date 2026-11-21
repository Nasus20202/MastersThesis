---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/immutable-configmap/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# immutable-configmap · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 1

## Task

The application must run with the configuration value v2 but currently reports v1. Update the application configuration to v2 and make the running Pods use it.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Immutable ConfigMaps

## Write commands

1. `kubectl patch cm app-config --type=merge -p '{"data":{"message":"v2"}}' 2>`
2. `kubectl delete cm app-config --wait=true`
3. `kubectl create -f /tmp/app-config-v2.yaml`
4. `kubectl rollout restart deployment/app`

## Final report

Done. The application now runs with `v2`.

## Root cause

- `default/app-config` had `data.message: v1` **and** `immutable: true`. Patching was rejected with:
  `The ConfigMap "app-config" is invalid: data: Forbidden: field is immutable when 'immutable' is set`
- The Deployment's Pod template injects the value via `env` → `configMapKeyRef {name: app-config, key: message}`, which is resolved **only at Pod creation**. So even a fixed ConfigMap would not have affected already-running Pods.

## Change made (smallest justified)

1. Recreated the owning resource under the same name `app-config`, same key `message`, with value `v2`, **keeping `immutable: true`** (preserved the original constraint; only the data value changed). Delete+create is the only way to change an immutable ConfigMap's data, and it kept the Deployment's reference and all unrelated settings untouched.
2. `kubectl rollout restart deployment/app` so new Pods re-resolve the env var. No Deployment spec fields, image, replicas, or selector were altered.

## Verified outcome (fresh observations)

- ConfigMap: `name=app-config data.message=v2 immutable=true`
- Live value inside every running Pod (read from the process, not just the spec):
  - `app-64b9cc88bb-4l95q CONFIG_MESSAGE=v2`
  - `app-64b9cc88bb-nvfbr CONFIG_MESSAGE=v2`
  - `app-64b9cc88bb-v2k96 CONFIG_MESSAGE=v2`
- Controller convergence matches the requested count: `replicas=3 ready=3 updated=3 available=3`, `generation=2, observedGeneration=2`; new ReplicaSet `app-64b9cc88bb` at 3/3, old `app-845b8b8485` scaled to 0.
- All 3 Pods `Running 1/1` with 0 restarts; no new warning events after the rollout.

Pre-existing `FailedScheduling` warnings on `app-845b8b8485-*` are from the initial startup before the node was ready and are unrelated to this configuration issue.
