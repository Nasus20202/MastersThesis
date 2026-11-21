---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/daemonset-missing-toleration/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# daemonset-missing-toleration · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A DaemonSet is not running on every node it should cover, so its node-level function is missing on part of the cluster. Restore the DaemonSet so it is ready on every node it should target, without modifying the cluster nodes.

Documented in:

- content/en/docs/concepts/workloads/controllers/daemonset.md: How Daemon Pods are scheduled > Taints and tolerations

## Write commands

1. `kubectl patch ds app -n default --type=strategic -p '{"spec":{"template":{"spec":{"tolerations":[{"key":"benchmark","operator":"Equal","value":"daemonset","effect":"NoSchedule"}]}}}}'`

## Final report

**Cause:** The DaemonSet `default/app` had no tolerations, while the only node carries taint `benchmark=daemonset:NoSchedule`. The DaemonSet controller therefore excluded the node (`desiredNumberScheduled: 0`) and the pre-taint pod was reported as `numberMisscheduled: 1` — so the node-level function was absent from the node it was supposed to cover.

**Change (single, on the DaemonSet's pod template only):**

```
spec.template.spec.tolerations:
- key: benchmark
  operator: Equal
  value: daemonset
  effect: NoSchedule
```

No node objects were modified — the taint is still present (`benchmark=daemonset:NoSchedule`), `unschedulable` still unset, and no labels/other pod-spec settings (image, command, resources, selector) were changed.

**Verified outcome (fresh observations):**

- DaemonSet status: `desiredNumberScheduled=1`, `currentNumberScheduled=1`, `numberReady=1`, `numberAvailable=1`, `updatedNumberScheduled=1`, `numberMisscheduled=0`, `observedGeneration=2`; `kubectl get ds` shows DESIRED/CURRENT/READY/UP-TO-DATE/AVAILABLE = 1/1/1/1/1.
- `kubectl rollout status ds/app` → "daemon set \"app\" successfully rolled out".
- Scheduled node count vs desired: `nodes=1 desired=1` — covers every node it should target.
- Pod `app-t2hmq` on `...control-plane`: `phase=Running`, `containerReady=true`, running the intended image `docker.io/library/busybox:1.37.0`, with the new toleration applied; the misscheduled old pod `app-gvj6j` was replaced by the controller's rolling update.
