---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/node-selector-no-match/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# node-selector-no-match · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/scheduling-eviction/assign-pod-node.md: nodeSelector

## Write commands

1. `kubectl -n default patch deploy/app --type=strategic -p '{"spec":{"template":{"spec":{"nodeSelector":null}}}}'`
2. `kubectl -n default run curl-check --image=curlimages/curl:8.10.1 --restart=Never --command -- curl -sS --max-time 10 "http://$IP/" 2>`
3. `kubectl -n default delete pod curl-check --ignore-not-found`

## Final report

## Root cause

The Deployment's revision 2 template added a required node placement rule:

```yaml
nodeSelector:
  disktype: ssd
```

The cluster has a single node, and it has no `disktype` label (its only labels are `kubernetes.io/{arch,hostname,os}` and `node-role.kubernetes.io/control-plane`). The scheduler therefore rejected the new pod permanently:

```
Warning  FailedScheduling  0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector.
```

That pinned the new ReplicaSet `app-77bd4fcf9c` at `READY 0/1` and left the rollout stuck (`updatedReplicas: 1`, `unavailableReplicas: 1`). Revision 1's template had no node selector.

## Change made (single, minimal)

Removed only the unsatisfiable `nodeSelector` from the Deployment's pod template:

```
kubectl -n default patch deploy/app --type=strategic -p '{"spec":{"template":{"spec":{"nodeSelector":null}}}}'
```

No node labels, taints, replica counts, images, or other resources were altered. Because the corrected template now matches revision 1, the Deployment reused the existing `app-575b7496bc` ReplicaSet and scaled the stalled one down.

## Verified outcome (fresh observations)

- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out`; condition `Progressing=True (NewReplicaSetAvailable)`.
- Replica counts match the requested 3: `replicas=3 updated=3 ready=3 available=3`, with `unavailableReplicas` empty; `generation=3 observedGeneration=3`.
- ReplicaSets: `app-575b7496bc` desired/current/ready = `3/3/3`; stalled `app-77bd4fcf9c` scaled to `0/0/0`.
- All 3 pods are `1/1 Running`, `Ready=True`, scheduled on the node with **no** nodeSelector, running the intended image `nginx:1.31.5`; `kubectl get pods --field-selector=status.phase=Pending` returns none.
- Functional check through the pod IP with a temporary curl pod returned nginx's `Welcome to nginx!` HTML (HTTP succeeded); nginx logs show workers running. The temporary probe pod was deleted.

The rollout is complete and the application is serving traffic with all 3 replicas ready and available.
