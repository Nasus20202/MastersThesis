---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/automount-token-disabled/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# automount-token-disabled · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `kubectl -n default patch deployment app --type=merge -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}' 2>`

## Final report

## Restored — verified healthy

**Symptom:** `deployment.apps/app` in `default` was stuck at `0/1` with pod `app-cf74c8f5f-694sd` in `Error` / CrashLoopBackOff (4 restarts).

**Evidence gathered**

Container logs showed the two-line failure:

```
cat: can't open '/var/run/secrets/kubernetes.io/serviceaccount/token': No such file or directory
wget: server returned error: HTTP/1.1 403 Forbidden
```

The Deployment's startup command authenticates to the API server using the projected SA token:
`wget ... --header "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config && exec sleep 3600`

The pod template carried `automountServiceAccountToken: false`, so the `token` file never existed → empty bearer header → 403 → `wget` non-zero → `exitCode: 1`.

I ruled out the other plausible causes before changing anything:

- **ServiceAccount `api-reader`** — no `automountServiceAccountToken: false` on the SA itself.
- **RBAC** — `Role/api-reader` grants `get` on `configmaps`, and `RoleBinding/api-reader` binds the `api-reader` SA (correct namespace/subject).
- **ConfigMap `app-config`** — exists with data `message: healthy`.

So authorization and the target object were fine; only the token projection was disabled.

**Change (one field, on the owning resource)**

```
kubectl -n default patch deployment app --type=merge \
  -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}'
deployment.apps/app patched
```

No RBAC, ConfigMap, image, or scheduling settings were altered.

**Verification of the original symptom being gone**

- Rollout: `deployment "app" successfully rolled out`.
- Replica counts match: `1 requested, 1 updated, 1 ready, 1 available`.
- Conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- New pod `app-f76b64bc8-8qkrg`: `1/1 Running`, `ready=true`, `restarts=0`, still `Ready` after ~106s.
- Token actually projected: `token`, `ca.crt`, `namespace` symlinks present under `/var/run/secrets/kubernetes.io/serviceaccount/`.
- App logic succeeds, not just pod health: PID 1 is `sleep 3600` (only reached if `wget` exited 0), and re-running the exact request from inside the pod returned `wget exit=0` (previously 403).
- No warning events since the fix; the last `BackOff`/`FailedScheduling` entries belong to the deleted pods from the broken revisions.
