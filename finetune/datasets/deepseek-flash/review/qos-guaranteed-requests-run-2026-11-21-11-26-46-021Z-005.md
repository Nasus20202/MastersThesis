---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/qos-guaranteed-requests/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# qos-guaranteed-requests · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl -n default patch deployment app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","resources":{"requests":{"cpu":"250m","memory":"128Mi"},"limits":{"cpu":"250m","memory":"128Mi"}}}]}}}}'`

## Final report

The application was the `default/app` Deployment (2 replicas, single `app` container, `resources: {}` → BestEffort). I set equal requests and limits for CPU and memory, which is what the kubelet requires for Guaranteed QoS.

**Change made** (smallest justified; only the container's `resources` block was modified):

```
kubectl -n default patch deployment app --type=strategic \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"app",
       "resources":{"requests":{"cpu":"250m","memory":"128Mi"},
                    "limits":{"cpu":"250m","memory":"128Mi"}}}]}}}}'
```

**Verification (fresh observations)**

- QoS class of the running Pods is `Guaranteed`, with requests equal to limits for both resources:
  - `app-7bdb575bbc-994jz`: QOS `Guaranteed`, cpu req/lim `250m`/`250m`, mem req/lim `128Mi`/`128Mi`
  - `app-7bdb575bbc-vxnh9`: QOS `Guaranteed`, cpu req/lim `250m`/`250m`, mem req/lim `128Mi`/`128Mi`
- Controller counts all match the requested replica count: `DESIRED 2, UPDATED 2, READY 2, AVAILABLE 2`; `rollout status` reported "successfully rolled out".
- Availability was preserved during the rollover: the Deployment's `maxUnavailable 25%` of 2 replicas resolves to 0, and events show the new ReplicaSet was scaled up and its Pods started/ready before the old ReplicaSet was scaled 2→1→0. No Pod was ever unavailable.
- The application still serves traffic: `curl http://127.0.0.1/` from each running Pod returned `HTTP 200` (port 80/http unchanged).
- Unrelated settings intact: replicas `2`, strategy `RollingUpdate` (maxUnavailable/maxSurge `25%`), image `nginx:1.31.5`, ports and labels unchanged. Node capacity is ample (requests now 950m/290Mi of 12 CPU / ~32Gi allocatable).
