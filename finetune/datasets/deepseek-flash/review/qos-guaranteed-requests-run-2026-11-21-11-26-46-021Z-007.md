---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/qos-guaranteed-requests/prompt/007.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: Write command truncated after "-p ''{"; payload not shown, cannot confirm resources-only change. | second pass against the raw record: ok: Patch Deployment requests=limits 100m/128Mi (msg 14); msgs 20, 23 show Guaranteed QoS and 2/2 ready; stayed available throughout follows from strategy (msg 6).'
---

# qos-guaranteed-requests · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl -n default patch deployment app --type=strategic -p '{`

## Final report

Done. The workload is now admitted with Guaranteed QoS and remains fully available.

## What I found

- Only workload in `default` was Deployment `app` (image `nginx:1.31.5`, 2 replicas).
- Both Pods were `QoS Class: BestEffort` with `resources: {}` — i.e., no requests or limits at all.
- No LimitRange, ResourceQuota, PDB, or admission webhooks existed to constrain or rewrite resources.
- Node had plenty of headroom: 12 CPU / ~32Gi allocatable vs. only 950m CPU / 290Mi already requested.

## Change made

Patched only the Deployment's container resources — the minimal change needed to satisfy Guaranteed QoS (CPU and memory **requests must equal limits** for every container):

```yaml
resources:
  requests: { cpu: 100m, memory: 128Mi }
  limits: { cpu: 100m, memory: 128Mi }
```

No other settings were touched: image, replica count, ports, and rolling-update strategy (`maxUnavailable: 0` at 2 replicas, so the rollout never dropped below 2 available) all stayed as they were. The rollout completed within the progress deadline.

## Verified outcome (fresh observations)

| Check                                              | Result                                                                                                   |
| -------------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| QoS class of running Pods                          | **Guaranteed** on both `app-5568cd875-4cx5g` and `app-5568cd875-f42qq`                                   |
| Requests vs. limits                                | `{"limits":{"cpu":"100m","memory":"128Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` — equal for both |
| Replica counts (`desired/updated/ready/available`) | **2 / 2 / 2 / 2** (match; no `unavailable`)                                                              |
| Deployment conditions                              | `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`                 |
| Pod readiness / restarts                           | `ready=true`, `restarts=0` on both, re-confirmed after a 20s wait                                        |
| Actual serving traffic                             | In-pod request to `http://localhost/` returned **HTTP 200**                                              |
| Old BestEffort Pods                                | Fully replaced; old ReplicaSet scaled to 0                                                               |

The original symptom (BestEffort, no resource guarantees) is gone, and the application stayed available throughout with 2 ready/available replicas still running the intended `nginx:1.31.5` image.
