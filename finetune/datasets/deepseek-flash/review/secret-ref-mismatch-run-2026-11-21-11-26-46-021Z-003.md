---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/secret-ref-mismatch/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# secret-ref-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl patch deployment app -n default --type=json -p '[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key","value":"password"}]' 2>`

## Final report

The application is healthy. Summary of what I observed and changed:

## Diagnosis

- Deployment `default/app` was stuck mid-rollout: `replicas=4` but `updatedReplicas=1` and `unavailableReplicas=1`, with the Progressing condition stuck on `ReplicaSetUpdated`.
- The new ReplicaSet `app-c6d597897` pod `app-c6d597897-4l566` was in `CreateContainerConfigError`.
- Kubelet event: `Error: couldn't find key passwd in Secret default/app-secret` (repeated 14 times).
- The Secret `app-secret` exists and contains exactly one key: `password: czNjcjN0` (`s3cr3t`). The previously working ReplicaSet `app-849cc7cbf4` (revision 1) referenced key `password`; the new pod template (revision 2) referenced a non-existent key `passwd`.
- So the fault lived in the Deployment's pod template (a bad secret key reference), not in the Secret — the Secret's content was correct and unchanged throughout.

## Change (single, minimal)

Patched only that one field on the persistent owner, the Deployment:

```
/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key: passwd -> password
```

No other settings (image, replicas, strategy, Secret data) were touched.

## Verification (fresh observations)

- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out`.
- Replica counts match the requested count exactly: `replicas=3, updatedReplicas=3, readyReplicas=3, availableReplicas=3`, `unavailableReplicas` empty, `observedGeneration=3` equals `generation=3`.
- Conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable: ReplicaSet "app-849cc7cbf4" has successfully progressed.)`.
- Pods: 3/3 `Running` and `1/1 Ready`, zero restarts; the broken ReplicaSet `app-c6d597897` scaled down `1 → 0` and its failing pod was removed.
- The env reference actually resolves at runtime inside a running pod: `APP_PASSWORD=s3cr3t` (matches the Secret's decoded value), with `secretKey=password` and `image=nginx:1.31.5` — i.e., the intended image is the one running.
- Container logs show nginx started normally (no config errors).
