---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/rollback-bad-release/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl -n default rollout undo deploy/app --to-revision=1 2>`

## Final report

The rollout was stuck because the new release's pod template was unstartable.

**Evidence found**

- `deploy/app` (namespace `default`, spec `replicas: 3`) had rollout revision 2 pointing at ReplicaSet `app-76ffd9fd7f` with `image: busybox:1.37.0` and `command: ["sh","-c","exit 1"]`.
- Those pods exited immediately: `STATUS Error`, `Exit Code: 1`, `Restart Count: 4`, kubelet `BackOff restarting failed container`. `readyReplicas: 2`, `unavailableReplicas: 2`, so the deployment was stuck at `2/3` with the old ReplicaSet `app-575b7496bc` (`nginx:1.31.5`, no command override) still serving.
- Revision 1 of the same Deployment was the known-good `nginx:1.31.5` template, so the smallest justified change was to restore that template rather than patch the broken command.

**Change made**

```
kubectl -n default rollout undo deploy/app --to-revision=1
deployment.apps/app rolled back
deployment "app" successfully rolled out
```

**Verified outcome (fresh observations)**

- Template restored: `image=nginx:1.31.5`, `command` absent.
- Controller counts match the request: `READY 3/3`, `UP-TO-DATE 3`, `AVAILABLE 3`; status `readyReplicas/replicas/availableReplicas = 3/3/3`, `unavailableReplicas` empty, `generation=3 observedGeneration=3`.
- Pods: three `app-575b7496bc-*` pods all `1/1 Running`, 0 restarts; the failed `app-76ffd9fd7f` ReplicaSet is scaled to `0/0/0`.
- Conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- Direct request through the app: `HTTP/1.1 200 OK`, `Server: nginx/1.31.5`.
- State held stable on a re-check 20s later: still `3/3/3` ready.
