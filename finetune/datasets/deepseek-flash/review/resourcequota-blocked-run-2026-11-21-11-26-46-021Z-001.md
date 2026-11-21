---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/resourcequota-blocked/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# resourcequota-blocked · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 1

## Task

The application rollout is not completing because the namespace ResourceQuota blocks additional Pods. Restore the application to three healthy replicas without removing the ResourceQuota.

Documented in:

- content/en/docs/concepts/policy/resource-quotas.md: How Kubernetes ResourceQuotas work

## Write commands

1. `kubectl patch resourcequota app-quota -n default --type=merge -p '{"spec":{"hard":{"pods":"3"}}}'`

## Final report

## Root cause

`ResourceQuota/app-quota` in namespace `default` hard-capped `pods: 2`, while `Deployment/app` is declared with `spec.replicas: 3`. The ReplicaSet `app-575b7496bc` could only create 2 pods and repeatedly failed the third with:

```
pods "app-575b7496bc-rfh5t" is forbidden: exceeded quota: app-quota, requested: pods=1, used: pods=2, limited: pods=2
```

which surfaced on the Deployment as `ReplicaFailure=True (FailedCreate)`, `Available=False (MinimumReplicasUnavailable)`, and a stalled rollout.

## Change made (single, minimal)

The owning persistent object is the ResourceQuota, so I raised only its pod ceiling to match the app's declared replica count — the quota itself and all other settings were left in place:

```
kubectl patch resourcequota app-quota -n default --type=merge -p '{"spec":{"hard":{"pods":"3"}}}'
```

## Verification (fresh observations)

- Quota still present and now satisfied: `app-quota  pods: 2/3` → `status.hard.pods=3`, `status.used.pods=3` (not deleted, not bypassed).
- Replica convergence matches the requested count: Deployment `app` `READY 3/3`, `UP-TO-DATE 3`, `AVAILABLE 3`; `spec.replicas=3`, `generation=2`, `status.observedGeneration=2`, `readyReplicas/availableReplicas=3`; ReplicaSet `app-575b7496bc` `DESIRED 3 / CURRENT 3 / READY 3`.
- `kubectl rollout status` completed: `deployment "app" successfully rolled out`.
- Conditions now `Progressing=True (NewReplicaSetAvailable)` and `Available=True (MinimumReplicasAvailable)`; the `ReplicaFailure` condition is gone and the latest events are only `SuccessfulCreate`/`Scheduled`/`Pulled`/`Created`/`Started` — no `FailedCreate`.
- Functional check on all three pods (image `nginx:1.31.5`, port 80): `http=200` from each of `app-575b7496bc-l2stp`, `app-575b7496bc-lwbcb`, `app-575b7496bc-xsx2j`.

Note for future rollouts: with `maxSurge: 25%` (rounds to 1) the strategy may momentarily want 4 pods, so a later version change could again hit the cap at 3; I did not pre-emptively widen the quota beyond the declared replica count, per the smallest-change constraint.
