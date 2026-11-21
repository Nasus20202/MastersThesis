---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/daemonset-missing-toleration/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl patch daemonset app -n default --type=merge -p '{"spec":{"template":{"spec":{"tolerations":[{"key":"benchmark","operator":"Equal","value":"daemonset","effect":"NoSchedule"}]}}}}' 2>`

## Final report

## Summary

**Problem Identified:**
The DaemonSet `app` in namespace `default` had a node tainting issue:

- The cluster node had taint `benchmark=daemonset:NoSchedule`
- The DaemonSet pod template was missing the corresponding toleration
- This caused 1 misscheduled node (`Number of Nodes Misscheduled: 1`)

**Fix Applied:**
Added the missing toleration to the DaemonSet via patch:

```yaml
tolerations:
  - key: "benchmark"
    operator: "Equal"
    value: "daemonset"
    effect: "NoSchedule"
```

**Verification - All Checks Passed:**

| Metric                       | Before Fix                  | After Fix     | Status   |
| ---------------------------- | --------------------------- | ------------- | -------- |
| Desired Nodes Scheduled      | 0                           | 1 ✓           | Fixed    |
| Current Nodes Scheduled      | 0                           | 1 ✓           | Fixed    |
| Number of Misscheduled Nodes | 1                           | 0 ✓           | Fixed    |
| Pods Running                 | 1/1 Ready                   | 1/1 Running ✓ | Verified |
| Pod Tolerations              | Missing benchmark/daemonset | Present ✓     | Fixed    |

The DaemonSet now runs correctly on all targeted nodes (the single control-plane node in this cluster).
