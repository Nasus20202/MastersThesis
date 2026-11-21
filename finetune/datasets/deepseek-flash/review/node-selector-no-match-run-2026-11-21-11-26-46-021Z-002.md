---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/node-selector-no-match/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# node-selector-no-match · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/scheduling-eviction/assign-pod-node.md: nodeSelector

## Write commands

1. `kubectl -n default patch deployment app --type=json -p '[{"op":"remove","path":"/spec/template/spec/nodeSelector"}]'`

## Final report

The rollout is now complete and healthy.

**Diagnosis (evidence):**

- Deployment `default/app` was at revision 2 with `nodeSelector: disktype=ssd` in its pod template. Revision 1's template had no node selector.
- The cluster's single node (`benchmark-node-selector-no-match-0de4870d-control-plane`) has labels only for arch/os/hostname/control-plane — no `disktype` label at all.
- The new ReplicaSet `app-77bd4fcf9c` created pod `app-77bd4fcf9c-f5qlz`, which stayed `Pending` with `FailedScheduling: 0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector`.
- The Deployment was stuck at 4 total replicas with only 1 updated (surge pod unschedulable), so the rollout could not progress.

**Change made (smallest justified fix, persistent owner):** Removed the unsatisfiable `nodeSelector` from the Deployment's pod template:

```
kubectl -n default patch deployment app --type=json -p '[{"op":"remove","path":"/spec/template/spec/nodeSelector"}]'
```

No node labels, capacity, replicas, image, or other settings were altered.

**Verified outcome (fresh observations):**

- `deployment.apps/app` patched; `rollout status` returned: `deployment "app" successfully rolled out`.
- Replica counts match requested: replicas `3`, updated `3`, ready `3`, available `3`, unavailable `0`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- Pods: 3/3 `Running` and `ready=true`, all on `nginx:1.31.5` (`docker.io/library/nginx:1.31.5`), matching the Deployment's intended image.
- Direct request via the Deployment's pod: `HTTP 200` from `http://localhost:80/`.
- Deployment pod template now has no node selector (`nodeSelector` output empty).

The originally reported symptom — a rollout that never completes — is gone.
