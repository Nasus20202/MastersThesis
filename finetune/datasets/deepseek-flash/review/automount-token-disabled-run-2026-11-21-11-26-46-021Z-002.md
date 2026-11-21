---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/automount-token-disabled/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# automount-token-disabled · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `kubectl -n default patch deployment app --type=json -p='[{"op":"replace","path":"/spec/template/spec/automountServiceAccountToken","value":true}]' 2>`

## Final report

The workload is healthy. Here's what I found and changed.

## Root cause

The `default/app` Deployment's pod template explicitly set `automountServiceAccountToken: false`, but the container's startup command requires the projected service-account token:

```
cat /var/run/secrets/kubernetes.io/serviceaccount/token  -> No such file or directory
wget ... Authorization: Bearer $(cat .../token)          -> HTTP/1.1 403 Forbidden
```

With no token file, `cat` failed and the unauthenticated request was rejected with 403, so the container exited → `Error`/`BackOff` → never Ready (Deployment `Available=False`, `MinimumReplicasUnavailable`).

I confirmed the rest of the wiring was already correct and did not touch it: ServiceAccount `api-reader` exists, `Role/api-reader` grants `get configmaps`, and `Rolebinding/api-reader` binds it (`kubectl auth can-i get configmaps --as=system:serviceaccount:default:api-reader -n default` → `yes`).

## Change (smallest justified, applied to the persistent owner)

Patched only the one field on the Deployment `default/app`:

```
automountServiceAccountToken: false  ->  true
```

## Verification (fresh observations)

- **Token mount restored:** `exec` shows `token` symlink, `TOKEN_LEN=1240`.
- **The exact failing startup command now succeeds:** running the Deployment's `wget ... Authorization: Bearer $(cat .../token) .../configmaps/app-config` returns `WGET_OK`.
- **Deployment converged:** `desired=1 updated=1 ready=1 available=1 unavailable=(none)`, `observedGeneration=3/3`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- **Pod healthy after soak:** `1/1 Running`, `ready=true restarts=0`, image `busybox:1.37.0` (intended image), no new `BackOff` events (the only BackOff in history is the old failing pod `app-cf74c8f5f-nrm5b`).

The original crash-on-startup symptom is gone and the constraints (same image, same SA/RBAC, same startup behavior) are intact.
