---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/daemonset-missing-toleration/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl patch daemonset app -n default --type='merge' -p '{"spec":{"template":{"spec":{"tolerations":[{"key":"benchmark","operator":"Equal","value":"daemonset","effect":"NoSchedule"}]}}}}' 2>`

## Final report

The DaemonSet is restored and its node-level function is present on the entire cluster.

## Root cause

The single cluster node carries the taint `benchmark=daemonset:NoSchedule`, but the `default/app` DaemonSet had no matching toleration. As a result the controller computed `desiredNumberScheduled: 0` and the node received no new DaemonSet pod. The one pod present (`app-qplft`) was a leftover scheduled before the taint took effect: `NoSchedule` does not evict already-running pods, so it lingered as `numberMisscheduled: 1` and was not managed by the current template.

## Change made

Added exactly one toleration to the DaemonSet's pod template (no other settings touched):

```
tolerations:
- key: benchmark
  operator: Equal
  value: daemonset
  effect: NoSchedule
```

The node was **not** modified — its taint is still `benchmark=daemonset:NoSchedule`.

## Verified outcome

- Rollout completed: `daemon set "app" successfully rolled out`.
- Final DaemonSet status: `desired=1 current=1 ready=1 available=1 upToDate=1 misscheduled=0` — all replica counts match, and the former misscheduled pod was replaced by `app-7gjj4`, which carries the new toleration.
- Coverage check: `nodes=1`, `readyPods=1` — the DaemonSet is ready on every node it targets.
