---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/daemonset-missing-toleration/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl patch daemonset app -n default -p '{"spec":{"template":{"spec":{"tolerations":[{"key":"benchmark","operator":"Exists","effect":"NoSchedule"}]}}}}'`

## Final report

## Issue Identified and Fixed

**Root Cause**: The node had a taint `benchmark=daemonset:NoSchedule` but the DaemonSet's pod template had no matching tolerations, preventing new pods from being scheduled.

**Fix Applied**: Added a toleration to the DaemonSet's pod template spec:

```yaml
tolerations:
  - key: benchmark
    operator: Exists
    effect: NoSchedule
```

**Verification Results**:

| Metric                            | Before Fix            | After Fix             |
| --------------------------------- | --------------------- | --------------------- |
| Desired Number of Nodes Scheduled | 0                     | 1                     |
| Current Number of Nodes Scheduled | 0                     | 1                     |
| Number of Nodes Misscheduled      | 1                     | 0                     |
| Pods Status                       | 1 Running / 0 Waiting | 1 Running / 0 Waiting |
| Pods Ready                        | 0                     | 1 (app-2rblj: 1/1)    |

The DaemonSet `app` in namespace `default` is now scheduled and ready on all nodes it should target (the single control-plane node with the `benchmark=daemonset:NoSchedule` taint).
