---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/rollback-bad-release/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl rollout undo deployment/app --to-revision=1`

## Final report

The application is restored and verified healthy.

## Root cause

Deployment `default/app` (3 replicas, `RollingUpdate` maxSurge 1 / maxUnavailable 1) had just been updated to **revision 2**, whose pod template was broken:

- image `busybox:1.37.0` with `command: ["sh","-c","exit 1"]` — the container exited immediately and crash-looped (`BackOff restarting failed container app`)
- The previous revision 1 ran `nginx:1.31.5` with no command override (working pods serving on port 80)

Because the new pods never reached Ready, the rollout stalled and only the 2 surviving old nginx pods were available (`availableReplicas: 2`, `unavailableReplicas: 2` of the intended 3).

## Change made

Single, smallest justified action — roll the Deployment back to its last known-good revision (no other spec fields touched):

```
kubectl rollout undo deployment/app --to-revision=1
```

## Verification (fresh observations)

- Rollout completed: `deployment "app" successfully rolled out`, condition `Progressing=True (NewReplicaSetAvailable)`, `Available=True (MinimumReplicasAvailable)`.
- Replica counts now match the requested count exactly: `kubectl get deploy app` → **READY 3/3, UP-TO-DATE 3, AVAILABLE 3**; status `replicas/readyReplicas/updatedReplicas/availableReplicas = 3/3/3/3` (previously 4/2/2/2).
- Running image is the intended one: all three pods (`app-575b7496bc-24mxm`, `-dcg8z`, `-f5dg6`) show `IMAGE=nginx:1.31.5`, `READY=true`, `RESTARTS=0`, `PHASE=Running`.
- Faulty ReplicaSet `app-76ffd9fd7f` (busybox) is scaled to `DESIRED 0 / CURRENT 0 / READY 0` and no pods from it remain (`No resources found`).
- Functional check, not just object status: in-pod request to the app returned `http_code=200` (re-confirmed on a second pod after convergence, 20s later) — the app is actually serving traffic again.

No other cluster settings, replica limits, or unrelated workloads were altered.
