---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/daemonset-missing-toleration/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl -n default patch ds app --type=strategic -p '{"spec":{"template":{"spec":{"tolerations":[{"key":"benchmark","operator":"Equal","value":"daemonset","effect":"NoSchedule"}]}}}}' 2>`

## Final report

## Summary

**Root cause:** The `app` DaemonSet in the `default` namespace had **no tolerations** in its pod template, while the only node (`benchmark-daemonset-missing-toleration-729efbd9-control-plane`) carries the taint `benchmark=daemonset:NoSchedule`. As a result the controller did not consider the node a target: it reported `desiredNumberScheduled=0`, `numberReady=0`, and marked the pre-existing pod as `numberMisscheduled=1`, so the node-level function was effectively unmanaged on that node.

**Change made (minimal, to the owning persistent resource only):** Added one matching toleration to the DaemonSet's pod template:

```yaml
tolerations:
  - key: benchmark
    operator: Equal
    value: daemonset
    effect: NoSchedule
```

No node objects were modified; the node's taint is still exactly `benchmark=daemonset:NoSchedule`.

**Verified outcome (fresh observations):**

- DaemonSet `app`: `DESIRED=1, CURRENT=1, READY=1, UP-TO-DATE=1, AVAILABLE=1`, `Misscheduled=0` — requested replica count matches updated/ready/available.
- Rollout produced pod `app-pfdqm` on the target node: `1/1 Running`, `ready=true`, running the intended image `busybox:1.37.0`.
- `describe` confirms `Tolerations: benchmark=daemonset:NoSchedule` and full node coverage (Desired = Current = Available = 1, Up-to-date = 1).
- Node taints remain unchanged, honoring the "do not modify cluster nodes" constraint.

The DaemonSet is now ready on every node it should target, with matching desired/current/ready/available counts.
