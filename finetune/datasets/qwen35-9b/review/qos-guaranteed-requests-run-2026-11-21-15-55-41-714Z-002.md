---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/qos-guaranteed-requests/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# qos-guaranteed-requests · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl set resources deployment/app -n default --limits=cpu=100m,memory=128Mi --requests=cpu=100m,memory=128Mi`

## Final report

## Summary

I've successfully reconfigured the `app` workload in the `default` namespace to run with **Guaranteed QoS** while maintaining availability.

### Changes Made

Updated the deployment's resource requirements using:

```
kubectl set resources deployment/app -n default --limits=cpu=100m,memory=128Mi --requests=cpu=100m,memory=128Mi
```

### Verification Results

**Guaranteed QoS Confirmed:**

- Both pods now show `QoS Class: Guaranteed`
- Resources are set to identical requests and limits per container:
  - **CPU**: `requests: 100m` = `limits: 100m`
  - **Memory**: `requests: 128Mi` = `limits: 128Mi`

**Availability Confirmed:**

- Deployment status: `available: 2, ready: 2, replicas: 2`
- All 2 pods are running with Ready status (1/1):
  - `app-5568cd875-lrxgz` → Running, Ready
  - `app-5568cd875-zwx62` → Running, Ready

### Key Point

For Guaranteed QoS in Kubernetes, **every container** must have CPU and memory requests equal to their respective limits. This ensures the scheduler guarantees the pod its resources and the kubelet enforces them.
