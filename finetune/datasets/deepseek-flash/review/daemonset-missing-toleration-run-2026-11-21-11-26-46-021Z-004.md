---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/daemonset-missing-toleration/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl patch daemonset app -n default --type='json' -p='[{"op":"add","path":"/spec/template/spec/tolerations","value":[{"key":"benchmark","operator":"Equal","value":"daemonset","effect":"NoSchedule"}]}]' 2>`

## Final report

## Root cause

`default/app` is a DaemonSet with no node selector (so it should cover every node), but it had **no tolerations**. The only node in the cluster carries a custom taint:

```
benchmark-daemonset-missing-toleration-...-control-plane
taints=[{"key":"benchmark","value":"daemonset","effect":"NoSchedule"}]
```

The DaemonSet controller therefore considered the node ineligible, evicted/ignored its pod (`desiredNumberScheduled: 0`, `numberMisscheduled: 1`), and the node-level function was absent from the whole cluster. A stale `app-d2w2n` pod was left over from before the taint was applied and was not the durable fix.

## Change made (DaemonSet only, nodes untouched)

Added a single, minimal toleration to the DaemonSet's pod template:

```json
{
  "key": "benchmark",
  "operator": "Equal",
  "value": "daemonset",
  "effect": "NoSchedule"
}
```

No other fields (image, resources, selector, update strategy) were altered, and the node's taint/labels are unchanged — verified after the fix: `taints=[{"effect":"NoSchedule","key":"benchmark","value":"daemonset"}]`.

## Verified outcome

- Controller counts now match the requested count (`desiredNumberScheduled: 1`):
  `currentNumberScheduled=1, desiredNumberScheduled=1, numberReady=1, numberAvailable=1, updatedNumberScheduled=1, numberMisscheduled=0, observedGeneration=2`
- `kubectl rollout status ds/app` → "daemon set "app" successfully rolled out" (UP-TO-DATE 1/1).
- Direct per-node coverage check: the one node `...-control-plane` runs `app-bxnzf`, `Ready=True`, `phase=Running`, `restarts=0`, image `busybox:1.37.0`; it was scheduled with the new toleration present and the stale pod was replaced.
