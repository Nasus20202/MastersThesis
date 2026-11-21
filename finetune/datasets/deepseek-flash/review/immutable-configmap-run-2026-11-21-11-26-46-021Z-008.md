---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/immutable-configmap/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# immutable-configmap · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

The application must run with the configuration value v2 but currently reports v1. Update the application configuration to v2 and make the running Pods use it.

Documented in:

- content/en/docs/concepts/configuration/configmap.md: Immutable ConfigMaps

## Write commands

1. `kubectl delete cm app-config -n default`
2. `kubectl create -f app-config-v2.yaml`
3. `kubectl rollout restart deployment/app -n default`

## Final report

Done. Root cause and verified fix:

**Root cause:** `default/app-config` was marked `immutable: true` with `data.message: v1`. Immutable ConfigMaps reject in-place updates (`field is immutable`), so the value could not simply be edited. The Deployment `app` consumed it via `configMapKeyRef` into the `CONFIG_MESSAGE` env var, and env vars are only read when a container starts — so even a corrected ConfigMap wouldn't affect the already-running Pods.

**Change made (smallest justified, persisting object-level fix):**

1. Recreated the persistent ConfigMap in place with the same name/namespace, preserving immutability, changing only the data value (`v1` → `v2`):
   - `app-config` → `data.message: v2`, `immutable: true` (new uid `1bb3e336-…`, resourceVersion `765`). Deployment reference (`configMapKeyRef → app-config/message`) left untouched.
2. `kubectl rollout restart deployment/app` so Pods were recreated and re-read the env var; the resulting ReplicaSet is `app-6b464c68b5`.

**Verified outcome (fresh observations):**

- All three Pods report the intended value: `app-6b464c68b5-hzmnq/ntq6k/wx5g6` each return `CONFIG_MESSAGE=v2` via `printenv` (previously `v1`).
- Deployment replica counts match exactly: `3 requested / 3 updated / 3 ready / 3 available`, `READY 3/3`, pods `1/1 Running` with `0` restarts; old ReplicaSet scaled to 0, so no stale v1 Pods remain.
- Constraints intact: ConfigMap still `immutable=true` and still owned by the same workload reference; image (`nginx:1.31.5`), replicas (3), and all other settings unchanged. Events show only normal rollout activity with no errors.
