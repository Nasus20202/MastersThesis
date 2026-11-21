---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/immutable-configmap/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# immutable-configmap · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

The application must run with the configuration value v2 but currently reports v1. Update the application configuration to v2 and make the running Pods use it.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Immutable ConfigMaps

## Write commands

1. `kubectl patch cm app-config -n default --type merge -p '{"data":{"message":"v2"}}' 2>`
2. `kubectl delete cm app-config -n default`
3. `kubectl create -f /tmp/app-config-v2.yaml`
4. `kubectl rollout restart deploy/app -n default`

## Final report

## Diagnosis

- `deploy/app` (namespace `default`, 3 replicas) injects `CONFIG_MESSAGE` from ConfigMap `app-config` key `message`.
- The ConfigMap had `data.message: v1` **and** `immutable: true`, so a normal update is rejected:
  `The ConfigMap "app-config" is invalid: data: Forbidden: field is immutable when 'immutable' is set`.
- Pods therefore reported `v1`, and since the value is injected as an env var at container start, editing the ConfigMap alone could never be picked up by running Pods — a replacement of the immutable object plus a rollout was required.

## Change made (smallest justified)

1. Deleted and recreated only `app-config`, preserving its `immutable: true` property and the `message` key, with the new value:
   ```
   configmap "app-config" deleted from default namespace
   configmap/app-config created
   ```
2. Restarted the Pods so the env var is re-resolved: `kubectl rollout restart deploy/app` → `deployment "app" successfully rolled out`.
   Deployment spec, image, replicas, selector and resource settings were left untouched.

## Verification (fresh observations)

- ConfigMap: `message = v2`, `immutable = true`.
- Each running Pod now reports the new value:
  ```
  app-fc799765f-r8hgz -> v2
  app-fc799765f-v8zjk -> v2
  app-fc799765f-wgx4k -> v2
  ```
- Controller converged: `replicas=3 updated=3 ready=3 available=3 generation=2 observed=2`; old ReplicaSet `app-845b8b8485` scaled to 0, new `app-fc799765f` at 3/3, all Pods `Running` with 0 restarts.
- No new warning events after the change (the only Warnings are pre-existing scheduling messages at cluster start, older than the new ReplicaSet).

Reported symptom (app reporting `v1`) is gone: all three running Pods report `v2`.
