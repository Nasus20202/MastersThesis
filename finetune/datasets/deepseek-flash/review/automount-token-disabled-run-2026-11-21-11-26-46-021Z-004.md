---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/automount-token-disabled/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# automount-token-disabled · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `kubectl -n default patch deployment app --type=strategic -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}' 2>`

## Final report

The workload is healthy and the fix is verified end-to-end.

## Root cause

The `app` Deployment in namespace `default` set `automountServiceAccountToken: false`, so the service-account token was never projected into the pod. The container's startup command reads `/var/run/secrets/kubernetes.io/serviceaccount/token` to authenticate to the API:

- `cat` failed with `No such file or directory`, so the Bearer token was empty;
- the API call returned `HTTP/1.1 403 Forbidden`;
- the `&&` chain exited non-zero, so the container crash-looped (`Error`, exit 1, restart backoff) and never became Ready.

The token mount was the only defect — RBAC (`auth can-i get configmaps` for `system:serviceaccount:default:api-reader` = `yes`) and the target `app-config` ConfigMap (`message: healthy`) were already correct.

## Change made

A single strategic patch to the owning Deployment, enabling the token on its pod template:

```
kubectl -n default patch deployment app --type=strategic \
  -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}'
```

## Verification (fresh observations)

- **Rollout:** `deployment "app" successfully rolled out`.
- **Deployment replica counts match:** desired=1, updated=1, ready=1, available=1, unavailable=none; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- **Pod:** `app-f76b64bc8-wgm2p` is `1/1 Running` with `0` restarts and `ready=true`.
- **Token present:** the token symlink resolves and is 1240 bytes.
- **Actual symptom path works:** replaying the app's own authenticated request inside the pod returns `HTTP/1.1 200 OK` with the body containing `"message": "healthy"` — the previous 403 is gone.

The original symptom (crash-loops, never Ready) is resolved, and unrelated settings and constraints were left intact.
