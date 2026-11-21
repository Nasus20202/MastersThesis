---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/qos-guaranteed-requests/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# qos-guaranteed-requests · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 3

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl set resources deployment/app -n default --requests=cpu=100m,memory=128Mi --limits=cpu=100m,memory=128Mi`

## Final report

## Summary

Successfully reconfigured the workload to achieve **Guaranteed QoS**:

### Changes Made:

- Updated Deployment `app` in namespace `default` with resource guarantees:
  - **CPU**: request=100m, limit=100m (matching)
  - **Memory**: request=128Mi, limit=128Mi (matching)

### Verification Results:

| Pod Name            | QoS Class  | Status  | Ready |
| ------------------- | ---------- | ------- | ----- |
| app-5568cd875-lfw9n | Guaranteed | Running | true  |
| app-5568cd875-lksjf | Guaranteed | Running | true  |

### Deployment Status:

- **Replicas**: 2/2 desired
- **Available**: 2
- **Ready**: 2

Both pods now have the required resource constraints (equal requests and limits), achieving Guaranteed QoS while maintaining full availability.
