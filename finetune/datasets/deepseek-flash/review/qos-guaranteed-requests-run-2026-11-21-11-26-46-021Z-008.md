---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/qos-guaranteed-requests/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# qos-guaranteed-requests · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl patch deploy app -n default --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","resources":{"requests":{"cpu":"100m","memory":"128Mi"},"limits":{"cpu":"100m","memory":"128Mi"}}}]}}}}'`

## Final report

## Root cause

The `app` Deployment in `default` had a single container with an empty resource block (`resources: {}`), so its Pods were admitted as **BestEffort** — no cpu/memory requests or limits, hence no resource guarantees.

## Change made

Patched the persistent owner (the Deployment, not the Pods) so the container declares requests equal to limits, which is exactly the condition the kubelet/API server uses to assign Guaranteed QoS:

```
kubectl patch deploy app -n default --type=strategic \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","resources":
      {"requests":{"cpu":"100m","memory":"128Mi"},
       "limits":{"cpu":"100m","memory":"128Mi"}}}]}}}}'
```

Values were chosen small relative to the node (allocatable 12 CPU / ~32Gi) so both replicas schedule with headroom, and only the `resources` field was touched — no LimitRange, ResourceQuota, or admission policy exists in this cluster to conflict with it.

## Verification (fresh observations after the change)

- **QoS class achieved:** both Pods report `qosClass=Guaranteed` (`app-5568cd875-5vh52`, `app-5568cd875-n479c`); each container shows `requests={"cpu":"100m","memory":"128Mi"}` and `limits={"cpu":"100m","memory":"128Mi"}` — identical, which is the Guaranteed requirement. No init containers or sidecars exist whose resources could break the class.
- **Availability preserved, counts match:** Deployment status `replicas=2, updated=2, ready=2, available=2, unavailable=<none>` — requested replica count equals updated/ready/available. `rollout status` reported "successfully rolled out" (exit 0), and the rollout event log shows the surge-first sequence (`app-5568cd875` scaled 1→2 before the old ReplicaSet went 2→1→0), so no moment had zero ready replicas.
- **Application actually serving, not just "healthy" objects:** `curl http://localhost/` from _inside each_ new Pod returned `http_code=200` with the nginx welcome page — confirming the tighter cgroup limits (100m CPU) didn't starve the container.
- **Constraints unchanged:** image still `nginx:1.31.5`, `replicas=2`, `strategy=RollingUpdate` with `maxSurge=25%`/`maxUnavailable=25%` — only the resource guarantees were added; nothing was deleted or restarted as a shortcut, and the old ReplicaSet was retired by the normal controller rollout.
