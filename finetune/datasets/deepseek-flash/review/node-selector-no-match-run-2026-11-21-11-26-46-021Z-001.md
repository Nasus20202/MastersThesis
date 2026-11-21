---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/node-selector-no-match/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# node-selector-no-match · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 1

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/scheduling-eviction/assign-pod-node.md: nodeSelector

## Write commands

1. `kubectl -n default patch deploy app --type=strategic -p '{"spec":{"template":{"spec":{"nodeSelector":null}}}}' 2>`

## Final report

The rollout is fixed and verified.

## Root cause

The `app` Deployment's revision 2 pod template added `nodeSelector: disktype=ssd`. The cluster's single node (`benchmark-node-selector-no-match-...-control-plane`) has no `disktype` label, so the new pod `app-77bd4fcf9c-l2djg` stayed `Pending`:

```
Warning  FailedScheduling  0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector.
```

Revision 1 had no node selector, which is why the old pods scheduled fine while `rollout status` reported "1 out of 3 new replicas have been updated".

## Change made (single, minimal)

Removed the unsatisfiable selector from the Deployment that owns the state, leaving image, replicas, ports, and strategy untouched:

```
kubectl -n default patch deploy app --type=strategic -p '{"spec":{"template":{"spec":{"nodeSelector":null}}}}'
→ deployment.apps/app patched
```

## Verification (fresh observations)

- `rollout status`: `deployment "app" successfully rolled out`
- Replica counts all match the requested count: `desired=3 updated=3 ready=3 available=3`, `unavailable` unset; `kubectl get deploy`: `READY 3/3`
- All 3 pods `Running`, `ready=true`, no node selector present, image is the intended `nginx:1.31.5`
- Stuck pod replaced: unschedulable ReplicaSet `app-77bd4fcf9c` scaled `1 → 0`
- Application actually serves traffic: HTTP request inside a pod to port 80 returned `http_code=200`
- Deployment conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`

Note: because removing the selector restored the exact revision‑1 pod template, the Deployment returned to ReplicaSet `app-575b7496bc` and scaled the failing one to zero — expected behavior, and the rollout is now complete with all 3 replicas available.
