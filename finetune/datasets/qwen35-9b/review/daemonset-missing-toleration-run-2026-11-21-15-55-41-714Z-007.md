---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/daemonset-missing-toleration/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: unsure: kubectl apply -f /tmp/fix-daemonset.yaml: file content and origin not shown; cannot confirm scope | second pass against the raw record: ok: Toleration benchmark=daemonset:NoSchedule copied from node taint; before and after counts match describe output."
---

# daemonset-missing-toleration · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl apply -f /tmp/fix-daemonset.yaml`

## Final report

## Summary

**Problem Identified:** The node `benchmark-daemonset-missing-toleration-0f3fc087-control-plane` had a taint `benchmark=daemonset:NoSchedule`, but the DaemonSet's pod template lacked a matching toleration. This prevented proper scheduling.

**Fix Applied:** Added a toleration to the DaemonSet:

```yaml
tolerations:
  - effect: NoSchedule
    key: benchmark
    operator: Equal
    value: "daemonset"
```

**Verification Results:**

| Metric                            | Before Fix | After Fix |
| --------------------------------- | ---------- | --------- |
| Desired Number of Nodes Scheduled | 0          | **1** ✓   |
| Current Number of Nodes Scheduled | 0          | **1** ✓   |
| Number Ready                      | 0          | **1** ✓   |
| Number of Nodes Misscheduled      | 1          | **0** ✓   |

The DaemonSet now runs successfully on all target nodes.
