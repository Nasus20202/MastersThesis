---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/statefulset-volumeclaim/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-volumeclaim · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Write commands

1. `kubectl create storageclass fast-ssd --provisioner=rancher.io/local-path --reclaimpolicy=Delete --volumebindingmode=WaitForFirstConsumer 2>`
2. `kubectl create sc fast-ssd --provisioner=rancher.io/local-path --reclaim-policy=Delete --volume-binding-mode=WaitForFirstConsumer 2>`
3. `kubectl create sc --help 2>`
4. `kubectl apply -f -`

## Final report

## Summary

**Problem Identified:** The StatefulSet `app` in namespace `default` was failing to start (0/2 ready replicas) because its volumeClaimTemplate referenced a non-existent storage class `fast-ssd`.

**Root Cause:** The StatefulSpec's volumeClaimTemplates specified `storageClassName: fast-ssd`, but this StorageClass did not exist in the cluster. Only `standard` existed.

**Fix Applied:** Created the missing StorageClass `fast-ssd` with matching parameters to the existing `standard` storage class:

- Provisioner: `rancher.io/local-path`
- ReclaimPolicy: `Delete`
- VolumeBindingMode: `WaitForFirstConsumer`

**Verification - Observed Changes:**

- StatefulSet status changed from `0/2` ready to `2/2` ready replicas
- Both pods (`app-0`, `app-1`) are now in `Running` state with 1/1 containers ready
- PVCs `data-app-0` and `data-app-1` are now `Bound` using storage class `fast-ssd`
- Final replica counts confirm all replicas are healthy: available=2, current=2, total=2, ready=2

The StatefulSet is now healthy with all volume claims bound.
