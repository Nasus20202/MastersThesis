---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/qos-guaranteed-requests/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# qos-guaranteed-requests · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl patch deployment app -n default --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","resources":{"requests":{"cpu":"200m","memory":"128Mi"},"limits":{"cpu":"200m","memory":"128Mi"}}}]}}}}'`

## Final report

## Diagnosis

- The workload in the `default` namespace is `deployment.apps/app` (2 replicas, image `nginx:1.31.5`).
- Its pod template had `resources: {}` — no requests, no limits — so both Pods were admitted with `qosClass: BestEffort` (confirmed via `.status.qosClass` before the change).
- Guaranteed QoS requires every container in every pod to have CPU **and** memory requests set exactly equal to their limits. There was no `LimitRange`/`ResourceQuota` in the cluster that would override or block this, and the node has ample capacity (`12` CPU, `~32Gi` memory allocatable, only `950m` CPU / `290Mi` memory requested cluster-wide).

## Change (single, minimal edit to the owning Deployment)

Patched only the `app` container's `resources` on `deployment/app` — no other settings (image, ports, probe-less config, rolling-update strategy, replica count) were touched:

```
requests: {cpu: 200m, memory: 128Mi}
limits:   {cpu: 200m, memory: 128Mi}   # requests == limits → Guaranteed
```

## Verification (fresh observations)

- Rollout: `deployment "app" successfully rolled out`.
- Replica counts match the requested count: `2 desired | 2 updated | 2 ready | 2 available | 0 unavailable`; old ReplicaSet `app-575b7496bc` scaled to 0, new `app-7b4d55f559` at 2/2.
- QoS per pod: `app-7b4d55f559-664ll` and `app-7b4d55f559-dcn2c` both report `QOS=Guaranteed`, `READY=true`, `RESTARTS=0` (re-checked ~20s later — still Guaranteed, 0 restarts).
- Persisted on the owning resource: `kubectl get deploy app -o jsonpath` shows `requests={"cpu":"200m","memory":"128Mi"} limits={"cpu":"200m","memory":"128Mi"}`.
- Availability proven functionally, not just by status: an in-pod request through each Pod returned `http_code=200` — both `app-7b4d55f559-664ll -> 200` and `app-7b4d55f559-dcn2c -> 200`.
- No scheduling failures or warning events; node allocated resources now include the app's `200m`/`128Mi` per Pod.
