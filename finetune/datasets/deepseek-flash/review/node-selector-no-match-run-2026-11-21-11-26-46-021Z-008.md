---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/node-selector-no-match/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# node-selector-no-match · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/scheduling-eviction/assign-pod-node.md: nodeSelector

## Write commands

1. `kubectl -n default patch deploy app --type=json -p='[{"op":"remove","path":"/spec/template/spec/nodeSelector"}]'`

## Final report

The rollout is now complete and the application is healthy.

**Diagnosis**

- `deployment/app` was on revision 2 with `READY 3/3` but `UP-TO-DATE 1`: a new ReplicaSet `app-77bd4fcf9c` had been created and its pod `app-77bd4fcf9c-8mjt9` was stuck `Pending`.
- Event evidence: `FailedScheduling — 0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector`.
- The new pod template had `nodeSelector: disktype=ssd`, while the previous template (old RS `app-575b7496bc`, still serving) had **no** nodeSelector. The single node in the cluster (`benchmark-node-selector-no-match-...-control-plane`) carries no `disktype` label at all, so the selector was unsatisfiable — no amount of waiting would let the rollout finish.

**Fix (smallest justified change)**

- Removed only the unsatisfiable pod-template field, leaving image, ports, resources, strategy, and the 3-replica count untouched:
  `kubectl -n default patch deploy app --type=json -p='[{"op":"remove","path":"/spec/template/spec/nodeSelector"}]'`
- Deployment moved to revision 3, generation 3, `observedGeneration: 3`.

**Verification (fresh observations)**

- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out`.
- Replica counts match the request: `requested=3 updated=3 ready=3 available=3` with no `unavailableReplicas`; conditions `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- Stuck RS `app-77bd4fcf9c` scaled to `0/0/0`; all 3 pods of `app-575b7496bc` are `Running`, `Ready=True`, placed on the node.
- Functional check inside every pod: `curl http://localhost/` returned `http_code=200` on all three pods.
