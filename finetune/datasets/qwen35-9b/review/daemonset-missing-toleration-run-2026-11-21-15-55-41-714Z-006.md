---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/daemonset-missing-toleration/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 6

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl patch daemonset app -n default --type=json -p='[{"op": "add", "path": "/spec/template/spec/tolerations", "value": [{"key": "benchmark", "operator": "Equal", "value": "daemonset", "effect": "NoSchedule"}]}]' 2>/dev/null`

## Final report

The issue has been resolved. Let me provide a final confirmation of the fix:

**Problem:**

- The node `benchmark-daemonset-missing-toleration-68c5393a-control-plane` had a taint `benchmark=daemonset:NoSchedule`
- The DaemonSet's pod template had no tolerations, preventing scheduling on this node
- Status showed: currentNumberScheduled=0, numberReady=0, numberMisscheduled=1

**Fix Applied:**

- Added toleration to the DaemonSet's pod template: `key: benchmark, operator: Equal, value: daemonset, effect: NoSchedule`

**Verification - Before vs After:**

| Metric                 | Before | After |
| ---------------------- | ------ | ----- |
| currentNumberScheduled | 0      | 1     |
| desiredNumberScheduled | 0      | 1     |
| numberReady            | 0      | 1     |
| numberMisscheduled     | 1      | 0     |
| numberAvailable        | -      | 1     |

**Current State:**

- DaemonSet `app` in namespace `default` is now scheduled and running on the single node it targets
- The running pod has the required toleration for `benchmark=daemonset:NoSchedule`
- Node status: Ready
- No pending or misscheduled pods

The DaemonSet is now ready on all nodes it should target.
