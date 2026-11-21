---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/secret-ref-mismatch/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# secret-ref-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl patch deploy app -n default --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","env":[{"name":"APP_PASSWORD","valueFrom":{"secretKeyRef":{"name":"app-secret","key":"password"}}}]}]}}}}' 2>`

## Final report

## Root cause

The `app` Deployment's rollout (revision 2) referenced a **non-existent Secret key**: the pod template asked for `secretKeyRef {name: app-secret, key: passwd}`, but `app-secret` only contains the key `password`. The kubelet therefore failed the container with:

```
Error: couldn't find key passwd in Secret default/app-secret   (CreateContainerConfigError)
```

Revision 1 of the same Deployment (old ReplicaSet `app-849cc7cbf4`, still serving) used `key: password`, confirming `passwd` was a typo introduced by the rollout. The bad pod blocked progress: `updatedReplicas: 1`, `unavailableReplicas: 1`, deployment stuck at 4 pods / 3 ready.

## Fix applied

One minimal change to the owning resource — the Deployment's pod template — correcting the Secret reference back to the existing key. The Secret itself was **not** modified (it is the source of truth and stayed at `resourceVersion 363`, data `{"password":"czNjcjN0"}`); no pods were deleted and no other settings (image, replicas, strategy, resources) were touched.

```
kubectl patch deploy app -n default --type=strategic -p \
 '{"spec":{"template":{"spec":{"containers":[{"name":"app","env":[{"name":"APP_PASSWORD","valueFrom":{"secretKeyRef":{"name":"app-secret","key":"password"}}}]}]}}}}'
→ deployment.apps/app patched
```

## Verification (fresh observations)

- `kubectl rollout status deploy/app`: **"deployment 'app' successfully rolled out"**.
- Replica counts match the request exactly: `NAME app  READY 3/3  UP-TO-DATE 3  AVAILABLE 3` (spec `replicas: 3`).
- Conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)` — no longer `ReplicaSetUpdated`/stuck.
- The config-error pod was replaced: RS `app-c6d597897` scaled `1 → 0`, pod `app-c6d597897-zw7bq` deleted; the Deployment's corrected template now matches RS `app-849cc7cbf4` (3/3 desired/current/ready), so the rollout converged and revision advanced to 3.
- All three running pods report `1/1 Running`, ready=true, image `nginx:1.31.5`, and `secretKeyRef.key=password`.
- The env var is actually injected at runtime: `kubectl exec … printenv APP_PASSWORD` → `s3cr3t` on all three pods (`app-849cc7cbf4-9nmt9`, `-js862`, `-z2lrj`), proving the Secret reference resolves rather than just that the pod object looks healthy.

No errors remain in the recent event stream; the only warnings are the historical `passwd` failures from before the fix.
