---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-15-19-27-938Z/resourcequota-blocked/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# resourcequota-blocked · deepseek-flash · run-2026-11-21-15-19-27-938Z · attempt 8

## Task

The application rollout is not completing because the namespace ResourceQuota blocks additional Pods. Restore the application to three healthy replicas without removing the ResourceQuota.

Documented in:

- content/en/docs/concepts/policy/resource-quotas.md: How Kubernetes ResourceQuotas work

## Write commands

1. `kubectl -n default patch resourcequota app-quota --type=merge -p '{"spec":{"hard":{"pods":"3"}}}'`

## Final report

The rollout is fixed and verified.

**Diagnosis (observed):**

- Deployment `default/app` requested `spec.replicas: 3` but reported `2/3` ready; ReplicaSet `app-575b7496bc` showed `ReplicaFailure: True / FailedCreate` with repeated events: `pods "app-575b7496bc-…" is forbidden: exceeded quota: app-quota, requested: pods=1, used: pods=2, limited: pods=2`.
- `ResourceQuota default/app-quota` had `spec.hard.pods: "2"` with `used: 2` — the only pods in the namespace were the two healthy `app=app` pods, so no quota was being wasted by stale/completed pods; the cap itself was simply one lower than the deployment's desired replica count.
- No LimitRange, node pressure, taint, or image problem existed (image `nginx:1.31.5` pulled and running, node `Ready`, no taints), so the quota was the single blocking cause.

**Change (smallest justified, quota preserved):**

```
kubectl -n default patch resourcequota app-quota --type=merge -p '{"spec":{"hard":{"pods":"3"}}}'
→ resourcequota/app-quota patched  (hard pods: "2" → "3")
```

The `ResourceQuota` object was kept in place; only its pod ceiling was raised to exactly the required replica count. No other quota field, deployment, or workload setting was touched.

**Verification (fresh observations after the change):**

- Controller counts now match the request: `desired=3 updated=3 ready=3 available=3`, `unavailable` empty; conditions `Progressing=True (NewReplicaSetAvailable)`, `Available=True (MinimumReplicasAvailable)`; `kubectl get deploy` shows `app 3/3 3 3`.
- ReplicaSet `app-575b7496bc` is `DESIRED 3 / CURRENT 3 / READY 3`.
- Three pods `Running` and `READY true`: `app-575b7496bc-5br7m`, `-7prg6`, and the newly admitted `-gnvgm`, each with image `nginx:1.31.5` (the intended image), scheduled on `benchmark-resourcequota-blocked-…-control-plane`.
- Traffic check: `curl` inside each pod against port 80 returned HTTP `200` for all three.
- Quota state is enforced and consistent: `app-quota  pods: 3/3` (hard 3, used 3), still present in the namespace.
- Events after the patch show `SuccessfulCreate`, `Scheduled`, `Pulled`, `Created`, `Started` for `app-575b7496bc-gnvgm`; no new `FailedCreate`/quota-exceeded events occurred after the patch.
