---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/statefulset-volumeclaim/prompt/005.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: kubectl apply of StorageClass fast-ssd (msg 26); PVCs data-app-0 and data-app-1 Bound and pods 1/1 verified (msgs 43, 53)."
---

# statefulset-volumeclaim · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 5

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Write commands

1. `kubectl patch statefulset app -n default -p '{"spec":{"volumeClaimTemplates":[{"apiVersion":"v1","kind":"PersistentVolumeClaim","metadata":{"labels":{"app":"app"},"name":"data"},"spec":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard","volumeMode":"Filesystem"}}]}}'`
2. `kubectl apply -f - << 'EOF'`

## Final report

## Summary

**Problem**: The StatefulSet `app` in namespace `default` could not start because its volumeClaimTemplate referenced a non-existent storage class `fast-ssd`. All PVCs remained in Pending state, preventing pod creation.

**Diagnosis**:

- Checked existing storage classes: only `standard` was available
- Inspected PVC `data-app-0`: status was Pending with message "storageclass.storage.k8s.io 'fast-ssd' not found"
- StatefulSet showed 2 replicas desired but 0/2 ready

**Fix Applied**: Created the missing storage class `fast-ssd` with provisioner `rancher.io/local-path` (matching the existing `standard` storage class configuration).

**Verification of Fix**:

| Check                           | Result                                             |
| ------------------------------- | -------------------------------------------------- |
| Storage class `fast-ssd` exists | ✓ Present                                          |
| PVC `data-app-0` bound          | ✓ Bound (pvc-b2e3e688-cdb9-40f6-beb1-79ed60083590) |
| PVC `data-app-1` bound          | ✓ Bound (pvc-aa49b653-24af-4450-8ceb-2d4e7723dff4) |
| StatefulSet replicas ready      | ✓ 2/2 READY                                        |
| Pod app-0 status                | ✓ Running, 1/1 containers ready                    |
| Pod app-1 status                | ✓ Running, 1/1 containers ready                    |

The StatefulSet is now healthy with all volume claims bound and both pods running successfully.
