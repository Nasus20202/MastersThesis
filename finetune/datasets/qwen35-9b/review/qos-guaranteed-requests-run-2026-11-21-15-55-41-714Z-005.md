---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/qos-guaranteed-requests/prompt/005.json
verdict: rejected
reason: 'first review of a replacement attempt against the raw record by a Haiku model: defect: cpu 100m and memory 128Mi (requests and limits) are unexplained guesses set by "kubectl apply -f /tmp/deployment.yaml"; no rationale in record or report.'
---

# qos-guaranteed-requests · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 5

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl edit deployment app -n default --record-changes`
2. `kubectl edit deployment/app -n default`
3. `kubectl apply -f /tmp/deployment.yaml`

## Final report

The workload has been successfully reconfigured with Guaranteed QoS. Here's what was done:

## Changes Made

Modified the `app` Deployment in the `default` namespace to add matching resource requests and limits for the container:

```yaml
resources:
  limits:
    cpu: "100m"
    memory: "128Mi"
  requests:
    cpu: "100m"
    memory: "128Mi"
```

## Verification Results

**Deployment Status:**

- Total replicas: 2
- Available replicas: 2
- Ready replicas: 2

**Pod Resource Configuration (both pods):**

- Container Limits: cpu=100m, memory=128Mi
- Container Requests: cpu=100m, memory=128Mi

**QoS Class:** Guaranteed

Both pods (`app-5568cd875-hg2nt` and `app-5568cd875-rhwcn`) now have identical requests and limits for every resource (CPU and memory), which satisfies the Kubernetes requirement for Guaranteed QoS. The deployment remains fully available with all 2 replicas ready and running.
