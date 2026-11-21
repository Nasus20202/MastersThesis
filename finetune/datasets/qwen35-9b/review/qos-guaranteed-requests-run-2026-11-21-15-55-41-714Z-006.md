---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/qos-guaranteed-requests/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# qos-guaranteed-requests · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 6

## Task

The application currently runs without any resource guarantees. Reconfigure the workload so its Pods are admitted with Guaranteed QoS while remaining available.

Documented in:

- content/en/docs/tasks/configure-pod-container/quality-service-pod.md: Create a Pod that gets assigned a QoS class of Guaranteed, Retrieve the QoS class for a Pod

## Write commands

1. `kubectl set resources deployment/app -n default --requests=cpu=100m,memory=128Mi --limits=cpu=100m,memory=128Mi`
2. `kubectl run -n default --rm --timeout=10s test-curl --image=curlimages/curl --restart=Never -- http://default/app 2>/dev/null`
3. `kubectl run -n default --rm --restart=Never test-check --image=curlimages/curl -- timeout 5 curl -s http://app-5568cd875-tj2zh:80`

## Final report

## Summary

### Change Made

Updated the `Deployment/app` in namespace `default` with matching resource requests and limits for Guaranteed QoS:

- **CPU**: 100m request / 100m limit
- **Memory**: 128Mi request / 128Mi limit

This triggered a rolling update, recreating the pods with the new configuration.

### Verified Outcomes

| Check              | Observed Value | Requirement                | Status  |
| ------------------ | -------------- | -------------------------- | ------- |
| QoS Class          | `Guaranteed`   | Guaranteed                 | ✅ Pass |
| CPU Limit          | 100m           | Equals request (100m)      | ✅ Pass |
| CPU Request        | 100m           | Matches limit (100m)       | ✅ Pass |
| Memory Limit       | 128Mi          | Equals request (128Mi)     | ✅ Pass |
| Memory Request     | 128Mi          | Matches limit (128Mi)      | ✅ Pass |
| Available Replicas | 2/2            | All desired replicas ready | ✅ Pass |

The workload now has **Guaranteed QoS** with both pods running and the deployment maintaining full availability.
