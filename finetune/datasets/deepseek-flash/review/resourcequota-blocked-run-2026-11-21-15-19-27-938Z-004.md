---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-15-19-27-938Z/resourcequota-blocked/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# resourcequota-blocked · deepseek-flash · run-2026-11-21-15-19-27-938Z · attempt 4

## Task

The application rollout is not completing because the namespace ResourceQuota blocks additional Pods. Restore the application to three healthy replicas without removing the ResourceQuota.

Documented in:

- content/en/docs/concepts/policy/resource-quotas.md: How Kubernetes ResourceQuotas work

## Write commands

1. `kubectl patch resourcequota app-quota -n default --type merge -p '{"spec":{"hard":{"pods":"3"}}}' 2>`

## Final report

## Summary

**Diagnosis (observed evidence):**

- `kubectl get deploy app -n default` showed `2/3` ready; the ReplicaSet `app-575b7496bc` was `3 desired / 2 ready` with condition `ReplicaFailure=True`.
- The Deployment status and ReplicaSet events carried the exact cause:
  `pods "app-575b7496bc-..." is forbidden: exceeded quota: app-quota, requested: pods=1, used: pods=2, limited: pods=2`
- `kubectl get resourcequota app-quota -n default` showed `pods: 2/2` — a hard ceiling of 2 pods against a Deployment whose `spec.replicas` legitimately requests 3.

So the blocker was the quota's hard limit, not the application: the ReplicaSet could never create the 3rd pod.

**Change made (smallest justified, quota preserved):**

```
kubectl patch resourcequota app-quota -n default --type merge -p '{"spec":{"hard":{"pods":"3"}}}'
```

Result: `resourcequota/app-quota patched`; `spec.hard.pods` went from `"2"` → `"3"`. The ResourceQuota object was **not** removed or disabled, and the Deployment's desired replica count was left untouched.

**Verification (fresh observations after the change):**

| Check                    | Observed value                                                                           |
| ------------------------ | ---------------------------------------------------------------------------------------- |
| Deployment spec replicas | 3                                                                                        |
| Deployment status        | `replicas=3 updated=3 ready=3 available=3`, `unavailableReplicas` empty                  |
| Conditions               | `Progressing=True (NewReplicaSetAvailable)`, `Available=True (MinimumReplicasAvailable)` |
| ReplicaSet               | `app-575b7496bc  DESIRED 3 / CURRENT 3 / READY 3`                                        |
| Pods                     | `5jx8x`, `wmt2s`, `xhrb5` — all `1/1 Running`, 0 restarts                                |
| Pods serving traffic     | HTTP `200` from each of the three pods                                                   |
| ResourceQuota            | still present: `app-quota`, hard `pods=3`, used `pods=3`                                 |
| FailedCreate events      | 0 in the 60s since the patch (the only ones are pre-patch, ≤ `limited: pods=2`)          |

The requested replica count (3) now matches updated (3), ready (3), and available (3), the original quota-forbidden symptom is gone, and the ResourceQuota remains enforced in the namespace.
