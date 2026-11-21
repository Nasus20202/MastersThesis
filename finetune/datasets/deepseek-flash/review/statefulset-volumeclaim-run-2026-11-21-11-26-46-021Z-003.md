---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-volumeclaim/prompt/003.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: Creates StorageClass fast-ssd; "mirrors the cluster''s existing standard class exactly" is not evidenced. | second pass against the raw record: ok: Creates fast-ssd StorageClass mirroring standard (msg 8); msgs 11, 13, 15 show both claims Bound, /data writes and HTTP 200; exactly is slightly overstated.'
---

# statefulset-volumeclaim · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Write commands

1. `kubectl apply -f -`

## Final report

**Root cause**

`statefulset/app` requests `storageClassName: fast-ssd` in its volumeClaimTemplate, but no StorageClass named `fast-ssd` existed in the cluster (only the default `standard` class). The provisioner therefore failed with `ProvisioningFailed: storageclass.storage.k8s.io "fast-ssd" not found`, leaving `data-app-0` `Pending`, `app-0` unschedulable ("pod has unbound immediate PersistentVolumeClaims"), and the StatefulSet stuck at 0/2.

**Fix (smallest justified change)**

Created the missing StorageClass rather than touching the StatefulSet (whose `volumeClaimTemplates` are immutable) or its existing claim:

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata: {name: fast-ssd}
provisioner: rancher.io/local-path
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

This mirrors the cluster's existing `standard` class exactly (same working `rancher.io/local-path` provisioner installed in `local-path-storage`), so no new components were introduced and no unrelated settings were altered.

**Verified outcome (fresh observations)**

- Claims bound: `data-app-0` → `pvc-c037ea34...` (1Gi RWO, `fast-ssd`), `data-app-1` → `pvc-cdff2b4f...` (1Gi RWO, `fast-ssd`), both `Bound`; events show `ProvisioningSucceeded` for both.
- Controller match: `readyReplicas=2`, `currentReplicas=2`, `updatedReplicas=2`, `availableReplicas=2` vs `spec.replicas=2` → `READY 2/2`.
- Pods: `app-0` and `app-1` both `Running` with `READY true`.
- Storage actually usable: wrote and read back a file on `/data` in both pods.
- Service path: `http://app-0.app.default.svc.cluster.local/` and `http://app-1...` both returned HTTP `200`.
