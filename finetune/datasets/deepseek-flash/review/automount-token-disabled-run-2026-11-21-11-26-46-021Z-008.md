---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/automount-token-disabled/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# automount-token-disabled · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Write commands

1. `kubectl -n default patch deploy app --type=json -p '[{"op":"replace","path":"/spec/template/spec/automountServiceAccountToken","value":true}]'`

## Final report

## Root cause

Deployment `default/app` sets `spec.template.spec.automountServiceAccountToken: false` while its container command authenticates to the API server using `/var/run/secrets/kubernetes.io/serviceaccount/token`:

```
wget ... --header "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
  https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config && exec sleep 3600
```

Container logs confirmed both failure steps:

```
cat: can't open '/var/run/secrets/kubernetes.io/serviceaccount/token': No such file or directory
wget: server returned error: HTTP/1.1 403 Forbidden
```

So `wget` exited non-zero → `sh -c` exited 1 → CrashLoopBackOff, `Available=False (MinimumReplicasUnavailable)`, restarts 3.

I checked the other plausible causes first and ruled them out:

- ServiceAccount `api-reader` exists; Role `api-reader` grants `get` on `configmaps`, bound via RoleBinding `api-reader` to that SA, and `kubectl auth can-i get configmaps/app-config -n default --as=system:serviceaccount:default:api-reader` → **yes**.
- ConfigMap `default/app-config` exists (`message: healthy`).
- Image `busybox:1.37.0` pulled fine ("already present on machine"); no scheduling/volume issues.

The only missing piece was the projected token, which `automountServiceAccountToken: false` (added in Deployment revision 2) suppressed.

## Change made (single field on the owning Deployment)

```
kubectl -n default patch deploy app --type=json \
  -p '[{"op":"replace","path":"/spec/template/spec/automountServiceAccountToken","value":true}]'
```

Nothing else was touched: image `busybox:1.37.0`, command/args, `serviceAccountName: api-reader`, Role/RoleBinding, ConfigMap, and `replicas: 1` are all unchanged.

## Verification (fresh observations)

- `rollout status`: `deployment "app" successfully rolled out`.
- Deployment status: `replicas=1 updated=1 ready=1 available=1`; conditions `Available=True/MinimumReplicasAvailable`, `Progressing=True/NewReplicaSetAvailable`.
- Pod `app-f76b64bc8-rnntn`: `1/1 Running`, `restarts=0`, state `running` — re-checked ~50s later, still `restarts=0` (no new BackOff events; last events are only Normal Pulled/Created/Started).
- Functional check inside the running container — the exact startup logic now succeeds:
  ```
  /var/run/secrets/kubernetes.io/serviceaccount/token -> ..data/token   (present)
  API_OK   (exit=0)
  ```
  i.e. the authenticated GET of `configmaps/app-config` returned successfully, so the container stays up in `sleep 3600` instead of exiting.
