---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/node-selector-no-match/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# node-selector-no-match · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/scheduling-eviction/assign-pod-node.md: nodeSelector

## Write commands

1. `kubectl -n default patch deploy app --type=json -p='[{"op":"remove","path":"/spec/template/spec/nodeSelector"}]' 2>`

## Final report

## Root cause

The `default/app` Deployment's rollout was blocked by an unsatisfiable `nodeSelector` in its new pod template:

- New revision `app-77bd4fcf9c` had `spec.template.spec.nodeSelector: {disktype: ssd}`
- The only node in the cluster (`benchmark-node-selector-no-match-76749e63-control-plane`) has **no** `disktype` label at all (`kubectl get nodes -l disktype=ssd` → "No resources found")
- Scheduler event on the new pod: `FailedScheduling ... 0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector`
- Result: 1 of 4 pods stuck `Pending`, `updatedReplicas: 1 / 3`, `unavailableReplicas: 1`, Progressing stuck at `ReplicaSetUpdated`

The image (`nginx:1.31.5`), ports, resources and strategy were all fine — the old ReplicaSet's pods had pulled the same image successfully.

## Change made (smallest justified, on the owning resource)

Removed only the impossible selector from the Deployment pod template — no other field touched:

```
kubectl -n default patch deploy app --type=json \
  -p='[{"op":"remove","path":"/spec/template/spec/nodeSelector"}]'
# deployment.apps/app patched
```

Replicas (3), image (`nginx:1.31.5`), container port 80, `RollingUpdate` 25%/25%, and node/taint settings were left intact. I did not relabel the control-plane node, since that would mutate cluster metadata to satisfy a bad workload spec.

## Verified outcome (fresh observations)

- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out` (exit 0)
- Replica comparison: **requested `replicas=3`, `updated=3`, `ready=3`, `available=3`, `unavailable` absent**, `observedGeneration=3 == generation=3`
- Conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)` — no longer `ReplicaSetUpdated`
- No `Pending` pods remain in the namespace; 3/3 pods `Running` with `Ready=True`, image `nginx:1.31.5`, all scheduled to the node
- Functional check in a pod: `curl http://localhost/` → `http_code=200`
- Deployment template matches the active ReplicaSet template exactly (identical spec digest), so the running pods are from the current desired template
- No new `FailedScheduling` warnings after the patch (the listed ones are historical, pre-fix); other workloads (`coredns 2/2`, `local-path-provisioner 1/1`) unaffected

Note: because the faulty revision differed from the previous one only by that selector, the rollout converged onto the existing `app-575b7496bc` ReplicaSet, whose pods already run the intended image — so no image change or pod restart was needed to reach the healthy state.
