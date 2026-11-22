---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-20-57-03-137Z/automount-token-disabled/prompt/001.json
verdict: accepted
reason: 'Re-enabling token automounting is the minimal in-scope fix, and the reported 1/1 Deployment, running pod and empty logs match the observed checks.'
---
# automount-token-disabled · deepseek-flash · run-2026-11-21-20-57-03-137Z · attempt 1

## Task

A workload in the cluster keeps crashing on startup and never becomes Ready. Restore it to a healthy state.

Documented in:
- content/en/docs/tasks/configure-pod-container/configure-service-account.md: Use the default service account to access the API server > Opt out of API credential automounting, ServiceAccount token volume projection

## Inherited stop

The first 21 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 1
stdout:

stderr:
Error from server (NotFound): pods "app-cf74c8f5f-qgj6g" not found

Its writes:

1. `kubectl rollout restart deployment app -n default`

## Teacher's write commands

1. `kubectl patch deployment app -n default --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/automountServiceAccountToken", "value": true}]'`

## Final report

The workload has been restored to a healthy state. Here is the diagnosis and resolution:

**Root Cause**
The `app` Deployment in the `default` namespace had `automountServiceAccountToken: false` set in its pod template. However, the container's startup command relies on the mounted service account token to authenticate with the Kubernetes API:
```yaml
args:
  - 'wget -q -O /dev/null --no-check-certificate --header "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://kubernetes.default.svc/api/v1/namespaces/default/configmaps/app-config && exec sleep 3600'
```
Because token automounting was disabled, `/var/run/secrets/kubernetes.io/serviceaccount/token` did not exist. Pod logs confirmed this with:
- `cat: can't open '/var/run/secrets/kubernetes.io/serviceaccount/token': No such file or directory`
- `wget: server returned error: HTTP/1.1 403 Forbidden`

Without the token, the `wget` request was unauthenticated, resulting in a 403 Forbidden, causing the container to exit with an error on every startup attempt.

**Fix Applied**
I applied the smallest justified change by patching the Deployment to re-enable service account token automounting:
```bash
kubectl patch deployment app -n default --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/automountServiceAccountToken", "value": true}]'
```

**Verification**
- The Deployment now reports `1/1` READY, `1` UP-TO-DATE, and `1` AVAILABLE.
- The new pod `app-f95d84c77-8nc95` is `1/1 Running` with `0` restarts.
- Pod logs are empty (exit code 0), confirming the startup command completed successfully without errors and the container transitioned to the `sleep 3600` phase.
