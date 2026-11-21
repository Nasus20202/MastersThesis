---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/secret-ref-mismatch/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# secret-ref-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl -n default patch deploy app --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key","value":"password"}]' 2>`

## Final report

## Summary

**Symptom:** Deployment `default/app` was stuck mid-rollout — `3/3` old pods served traffic but the new ReplicaSet `app-c6d597897` had `0/1` ready, with its pod in `CreateContainerConfigError`. Rollout never converged (`updatedReplicas: 1`, `unavailableReplicas: 1`).

**Root cause (from observed evidence):**

- Pod event: `Warning Failed ... Error: couldn't find key passwd in Secret default/app-secret`
- The Deployment's new pod template referenced `secretKeyRef.key: passwd`:
  ```json
  {
    "name": "APP_PASSWORD",
    "valueFrom": { "secretKeyRef": { "key": "passwd", "name": "app-secret" } }
  }
  ```
- Secret `app-secret` only contains one key: `{"password":"czNjcjN0"}` (no `passwd`).
- All previously-running pods and the old ReplicaSet `app-849cc7cbf4` used `key: password`, confirming `password` is the correct/expected key and the new template introduced the mismatch.

**Change made (single, minimal):** patched only that one field on the Deployment that owns the faulty rollout state — the secret key reference back to the existing key:

```
kubectl -n default patch deploy app --type=json \
  -p='[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key","value":"password"}]'
```

The Secret was left untouched (no new key invented), and `replicas: 3`, image `nginx:1.31.5`, labels, and rollout strategy were preserved.

**Verified outcome:**

- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out`
- Counts match the requested replica count: `3 requested / 3 total / 3 updated / 3 ready / 3 available / 0 unavailable`
- Conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)` — rollout complete, revision 3
- Pods: 3 × `Running`, container `ready=true`, image `docker.io/library/nginx:1.31.5` (the intended image)
- Functional check through the app itself: `printenv APP_PASSWORD` inside every pod returns `s3cr3t`, i.e. the secret value now actually resolves at runtime (previously the container could not even be created)
- The failing ReplicaSet `app-c6d597897` (rev 2, key `passwd`) is scaled to `0` and retained as normal revision history.
