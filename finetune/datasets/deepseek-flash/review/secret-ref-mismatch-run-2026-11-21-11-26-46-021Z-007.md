---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/secret-ref-mismatch/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# secret-ref-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl patch deploy app -n default --type=json -p '[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key","value":"password"}]'`

## Final report

## Root cause

Deployment `app` (namespace `default`, image `nginx:1.31.5`, 3 replicas) was stuck mid-rollout. Revision 2 of the pod template pointed its env var at a non-existent secret key:

```
APP_PASSWORD -> secretKeyRef{name: app-secret, key: passwd}
```

but `app-secret` only contains the key `password` (`app-secret` data was `{"password":"czNjcjN0"}`). The rollout's only new pod therefore sat in `CreateContainerConfigError`, with kubelet events repeating:

```
Warning Failed  Error: couldn't find key passwd in Secret default/app-secret
```

The Deployment showed `replicas: 4`, `unavailableReplicas: 1`, `updatedReplicas: 1`, and the new ReplicaSet `app-c6d597897` never became ready, so the rollout never completed.

## Fix

The secret was intact and the previously working ReplicaSet (`app-849cc7cbf4`) used the existing key `password`, so the faulty change was on the Deployment side. Smallest justified change — one field patched back to the existing key, leaving the secret and all other settings untouched:

```
kubectl patch deploy app --type=json \
  -p '[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key","value":"password"}]'
```

## Verification (fresh observations after the change)

- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out`
- Replica counts match the request: `READY 3/3`, `UP-TO-DATE 3`, `AVAILABLE 3`; status `replicas: 3`, `readyReplicas: 3`, `availableReplicas: 3`, no `unavailableReplicas`
- Deployment conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`; `observedGeneration 3 == generation 3`
- The old failing pod `app-c6d597897-zj7hp` is gone and its ReplicaSet scaled to `0/0/0`; no non-Running pods remain in `default`
- All 3 pods `1/1 Running` with `0` restarts after a further 15s wait
- Container actually resolved the secret: `kubectl exec ... echo $APP_PASSWORD` → `APP_PASSWORD=s3cr3t`, and nginx serves traffic: `http_code=200`
- Secret unchanged: still exactly `{"password":"czNjcjN0"}` — no keys added or modified.
