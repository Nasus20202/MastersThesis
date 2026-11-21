---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/rollback-bad-release/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl -n default rollout undo deployment/app --to-revision=1 2>`

## Final report

## Diagnosis

`deployment/app` (replicas: 3, strategy maxUnavailable 1/maxSurge 1) was on revision 2, whose pod template was broken:

- **Image:** `busybox:1.37.0`
- **Command:** `sh -c exit 1` → container terminated with exit code 1 every start (`BackOff` events, RESTARTS 4, `0/1 Error`)
- Result: only 2 old pods available (`2/3` ready, `2 unavailable`, rollout stuck at "ReplicaSet app-76ffd9fd7f is progressing")

Revision 1's template was `nginx:1.31.5` with no failing command, and its ReplicaSet `app-575b7496bc` still had 2 healthy pods.

## Change made

Rolled the Deployment back to its last good template (the persistent owner of the bad state), leaving the rollout strategy, replicas, and probes untouched:

```
kubectl -n default rollout undo deployment/app --to-revision=1
# deployment.apps/app rolled back
```

## Verification (fresh observations)

- `rollout status`: **deployment "app" successfully rolled out**
- Deployment: `READY 3/3`, `UP-TO-DATE 3`, `AVAILABLE 3`; status `replicas=3, updatedReplicas=3, readyReplicas=3, availableReplicas=3`, no `unavailableReplicas`; `observedGeneration=generation=3`
- Conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`
- Pods: all three `app-575b7496bc-*` are `1/1 Running` with 0 restarts; the broken `app-76ffd9fd7f` ReplicaSet was scaled to `0/0`
- Running template is the intended one: `image=nginx:1.31.5`, `command=` (empty)
- Traffic check inside a running pod: `curl http://127.0.0.1:80/` returned **HTTP 200**, confirming the app actually serves requests, not just that objects look healthy

The application is restored to 3/3 available replicas running the known-good `nginx:1.31.5` image, and the rollout has completed.
