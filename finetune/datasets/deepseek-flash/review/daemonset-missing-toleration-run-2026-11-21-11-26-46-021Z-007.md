---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/daemonset-missing-toleration/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl -n default patch ds app --type=merge -p '{"spec":{"template":{"spec":{"tolerations":[{"key":"benchmark","operator":"Equal","value":"daemonset","effect":"NoSchedule"}]}}}}'`

## Final report

The DaemonSet is restored and verified.

**Root cause:** the sole node carries taint `benchmark=daemonset:NoSchedule`, while the `default/app` DaemonSet pod template had no matching toleration. The DaemonSet controller therefore excluded that node from its desired set (`desiredNumberScheduled: 0`) and treated the existing pod as misscheduled (`numberMisscheduled: 1`), so the node-level function was not considered covered.

**Change made (smallest justified, resource-level only):** added one toleration to the DaemonSet pod template — `key=benchmark, operator=Equal, value=daemonset, effect=NoSchedule`. No node, taint, label, selector, image, or resource setting was altered.

**Verified outcome (fresh observations):**

- DaemonSet status: `desiredNumberScheduled: 1`, `currentNumberScheduled: 1`, `numberReady: 1`, `numberAvailable: 1`, `updatedNumberScheduled: 1`, `numberMisscheduled: 0`, `observedGeneration: 2` — all four counts match the requested target of 1 node.
- Pod `app-rr5tj` is `Running`, `ready=true`, scheduled on `...-control-plane`, running the intended `docker.io/library/busybox:1.37.0` image (old pre-change pod `app-xjlcc` was replaced by the rolling update; `kubectl rollout status ds/app` reported "successfully rolled out").
- Cluster nodes untouched: still exactly one node, `Ready`, with its original taint `benchmark=daemonset:NoSchedule` intact.
