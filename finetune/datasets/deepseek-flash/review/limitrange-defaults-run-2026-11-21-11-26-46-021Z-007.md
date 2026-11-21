---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/limitrange-defaults/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# limitrange-defaults · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The default namespace has a LimitRange that supplies default container resources. Create a Deployment named app in that namespace with image nginx:1.31.5 and two replicas, without setting container resources, so the Pods inherit the LimitRange defaults.

Documented in:

- content/en/docs/concepts/policy/limit-range.md: Constraints on resource limits and requests

## Write commands

1. `kubectl create deployment app -n default --image=nginx:1.31.5 --replicas=2 --dry-run=client -o yaml`
2. `kubectl create deployment app -n default --image=nginx:1.31.5 --replicas=2`

## Final report

Done — the Deployment was created and verified.

## What I did

- Discovered the pre-existing LimitRange in `default`: `app-limits` (Container type) supplying defaults `cpu: 200m / memory: 128Mi` (limits) and `cpu: 100m / memory: 64Mi` (requests).
- Created the Deployment with `kubectl create deployment app -n default --image=nginx:1.31.5 --replicas=2`, deliberately leaving container `resources` unset (the Deployment's pod template has `resources: {}`).

## Verification (fresh observations)

- Controller counts match the requested 2 replicas:
  - `deployment.apps/app   2/2   2   2` (READY / UP-TO-DATE / AVAILABLE), `rollout status` → "successfully rolled out"
  - `replicaset.apps/app-7bd5745f6c   2 desired, 2 current, 2 ready`
- Both Pods are `1/1 Running`, 0 restarts, and are running the intended image `nginx:1.31.5`, with successful `Pulled`/`Started` events.
- LimitRange defaults were actually injected by the admission controller into the live Pods:
  - `requests={"cpu":"100m","memory":"64Mi"} limits={"cpu":"200m","memory":"128Mi"}` on both Pods
  - while the Deployment's own pod template still shows `resources: {}` — i.e. no resources were set by me.

Both the requested state (2/2 ready Pods using `nginx:1.31.5` with LimitRange-inherited resources) and the constraint (no container resources specified in the Deployment) hold.
