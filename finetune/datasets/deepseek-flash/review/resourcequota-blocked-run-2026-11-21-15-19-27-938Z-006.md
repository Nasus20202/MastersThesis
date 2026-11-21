---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-15-19-27-938Z/resourcequota-blocked/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# resourcequota-blocked · deepseek-flash · run-2026-11-21-15-19-27-938Z · attempt 6

## Task

The application rollout is not completing because the namespace ResourceQuota blocks additional Pods. Restore the application to three healthy replicas without removing the ResourceQuota.

Documented in:

- content/en/docs/concepts/policy/resource-quotas.md: How Kubernetes ResourceQuotas work

## Write commands

1. `kubectl patch resourcequota app-quota -n default --type merge -p '{"spec":{"hard":{"pods":"3"}}}'`

## Final report

Resolved. The ResourceQuota was the only blocker.

**What was wrong**

- `resourcequota/app-quota` in `default` had `hard.pods: "2"` and `used.pods: "2"`.
- `deployment/app` requests 3 replicas; its ReplicaSet repeatedly emitted `FailedCreate: ... exceeded quota: app-quota, requested: pods=1, used: pods=2, limited: pods=2`, leaving `2/3` ready with `Available=False (MinimumReplicasUnavailable)`.

**Change made (smallest justified, quota retained)**

- Patched only the pod allowance: `kubectl patch resourcequota app-quota -n default --type merge -p '{"spec":{"hard":{"pods":"3"}}}'`.
- The quota still exists and enforces a hard cap of 3 pods; no other limits or settings were touched.

**Verified outcome**

- Deployment: `READY 3/3`, `UP-TO-DATE 3`, `AVAILABLE 3`; conditions `Progressing=True (NewReplicaSetAvailable)` and `Available=True (MinimumReplicasAvailable)`.
- ReplicaSet `app-575b7496bc`: desired 3, current 3, ready 3.
- Quota status: `hard pods: 3`, `used pods: 3` (`app-quota` still present).
- Direct request through a running app pod: `nginx HTTP 200`, confirming the three replicas are serving.
