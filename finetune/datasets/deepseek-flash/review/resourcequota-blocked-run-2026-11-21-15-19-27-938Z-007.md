---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-15-19-27-938Z/resourcequota-blocked/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# resourcequota-blocked · deepseek-flash · run-2026-11-21-15-19-27-938Z · attempt 7

## Task

The application rollout is not completing because the namespace ResourceQuota blocks additional Pods. Restore the application to three healthy replicas without removing the ResourceQuota.

Documented in:

- content/en/docs/concepts/policy/resource-quotas.md: How Kubernetes ResourceQuotas work

## Write commands

1. `kubectl -n default patch resourcequota app-quota --type merge -p '{"spec":{"hard":{"pods":"3"}}}'`

## Final report

## Root cause

The namespace ResourceQuota `default/app-quota` had `spec.hard.pods: "2"`, while `Deployment/app` requires 3 replicas. The ReplicaSet `app-575b7496bc` was stuck at 2 current/3 desired, with repeated `FailedCreate` events:

```
pods "app-575b7496bc-..." is forbidden: exceeded quota: app-quota, requested: pods=1, used: pods=2, limited: pods=2
```

## Change made

Kept the ResourceQuota in place and raised only its pod ceiling, the minimum adjustment needed for the declared replica count:

```
kubectl -n default patch resourcequota app-quota --type merge -p '{"spec":{"hard":{"pods":"3"}}}'
→ resourcequota/app-quota patched
```

No Deployment, Pod, or other quota fields were modified or deleted.

## Verification (fresh observations)

- **Deployment replicas match**: `spec.replicas=3`, `updatedReplicas=3`, `readyReplicas=3`, `availableReplicas=3`, `unavailableReplicas` absent → `3/3` READY / UP-TO-DATE / AVAILABLE; conditions `Available=True/MinimumReplicasAvailable`, `Progressing=True/NewReplicaSetAvailable`. `rollout status` returned "deployment \"app\" successfully rolled out".
- **Pods healthy and intended image**: all three pods `app-575b7496bc-8g6g4`, `-grh68`, `-h9zw8` have `image=nginx:1.31.5`, `ready=true`, `restartCount=0`, status `Running`.
- **Application serves traffic**: `curl http://127.0.0.1/` from a pod returned `http_code=200`.
- **Quota preserved and not over-committed**: `app-quota   pods: 3/3` (hard `3`, used `3`).
- **Symptom gone**: last `FailedCreate`/quota-exceeded event is `15:27:42Z`; the third pod was created at `15:28:03Z` and no quota rejection has occurred since.
