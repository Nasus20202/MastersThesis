---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/automount-token-disabled/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# automount-token-disabled · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `kubectl -n default patch deployment app --type=merge -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}'`

## Final report

The workload is healthy. Summary of what I found and the single change made:

## Diagnosis

- Crashing pod: `default/app-cf74c8f5f-bljkd` (Deployment `app`), `Error`, 4 restarts, `Ready=False`, exit code 1.
- Container logs:
  ```
  cat: can't open '/var/run/secrets/kubernetes.io/serviceaccount/token': No such file or directory
  wget: server returned error: HTTP/1.1 403 Forbidden
  ```
  The startup gate fetches `configmaps/app-config` with a bearer token, but the token file didn't exist.
- Ownership/evidence chain: Deployment `app` sets `spec.template.spec.automountServiceAccountToken: false` while using `serviceAccountName: api-reader` and reading the token path. The token was never projected into the pod (`describe pod` showed `Volumes: <none>`).
- I ruled out the other plausible causes rather than assuming:
  - ServiceAccount `api-reader` has no `automountServiceAccountToken` override (`nil`, inherits `true`) — the SA was not the blocker.
  - `kubectl auth can-i get configmaps -n default --as=system:serviceaccount:default:api-reader` → `yes`; Role `api-reader` grants `configmaps: [get]` and the RoleBinding binds it to that SA, so the 403 was purely the anonymous request sent without a token, not an RBAC gap.
  - The target ConfigMap `app-config` exists in `default`.

## Repair (minimal, on the owning resource)

Patched only the Deployment's pod template flag, leaving image, args, RBAC, ServiceAccount, and strategy untouched:

```
kubectl -n default patch deployment app --type=merge \
  -p '{"spec":{"template":{"spec":{"automountServiceAccountToken":true}}}}'
```

## Verification (fresh observations)

- Deployment rollout converged: requested `replicas=1`, and status `replicas=1 updatedReplicas=1 readyReplicas=1 availableReplicas=1 unavailableReplicas=(none)`; `READY 1/1`, `AVAILABLE 1`.
- New pod `app-f76b64bc8-x2lm4`: `1/1 Running`, `restarts=0`, `ready=true`, and it stayed that way across a further 45s check (no new `BackOff` events).
- Direct check of the originally failing operation inside the running pod:
  - `/var/run/secrets/kubernetes.io/serviceaccount/{token,ca.crt,namespace}` now present (projected volume mounted).
  - The exact startup `wget` against `https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config` returned `wget_exit=0` and body containing `"message": "healthy"`.
  - PID 1 is `sleep 3600`, i.e. the startup gate passed and the app moved on to its steady-state command instead of exiting 1.
