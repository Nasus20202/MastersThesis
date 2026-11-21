---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/qos-guaranteed-requests/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# qos-guaranteed-requests · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl -n default patch deployment app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","resources":{"requests":{"cpu":"100m","memory":"128Mi"},"limits":{"cpu":"100m","memory":"128Mi"}}}]}}}}'`

## Final report

Fixed. The `default/app` Deployment had a container with `resources: {}` (no requests/limits), so its Pods were `BestEffort`.

**Change made** — patched only the `app` container's resources in the Deployment's pod template (one container, `nginx:1.31.5`):

```yaml
resources:
  requests: { cpu: 100m, memory: 128Mi }
  limits: { cpu: 100m, memory: 128Mi }
```

Requests equal limits for both CPU and memory on every container, which is what the QoS admission logic requires for `Guaranteed`. Values are well within the node's allocatable 12 CPU / ~31Gi memory, and the existing `RollingUpdate` strategy (`maxUnavailable: 25%`) kept capacity during the roll.

**Verified after the change (fresh observations):**

- QoS class of both running Pods: `Guaranteed` (`app-5568cd875-lft7k`, `app-5568cd875-xcrjw` → `status.qosClass: Guaranteed`), with `requests = limits = {cpu:100m, memory:128Mi}`.
- Availability: Deployment `spec.replicas=2`, `status.replicas=2`, `updatedReplicas=2`, `readyReplicas=2`, `availableReplicas=2`, no `unavailableReplicas`; `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`; `rollout status` completed successfully.
- Old BestEffort ReplicaSet `app-575b7496bc` scaled to 0/0/0; only the new ReplicaSet remains.
- Application actually serving: HTTP 200 from `curl http://localhost:80/` inside a Pod.
