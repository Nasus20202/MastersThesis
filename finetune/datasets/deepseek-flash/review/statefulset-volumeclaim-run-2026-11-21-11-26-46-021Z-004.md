---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-volumeclaim/prompt/004.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: delete --cascade=orphan and re-apply of STS with storageClassName standard (existing class); claims Bound, pods Ready, probes shown."
---

# statefulset-volumeclaim · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Write commands

1. `kubectl scale sts app --replicas=0`
2. `kubectl patch sts app --type=json -p '[{"op":"replace","path":"/spec/volumeClaimTemplates/0/spec/storageClassName","value":"standard"}]'`
3. `kubectl delete pvc data-app-0 --wait=true`
4. `kubectl delete sts app --cascade=orphan`
5. `kubectl apply -f app-sts-fixed.yaml`

## Final report

## Root cause

The StatefulSet `app` (namespace `default`) had `spec.volumeClaimTemplates[0].spec.storageClassName: fast-ssd`, but no such StorageClass existed in the cluster (only `standard`, the default `rancher.io/local-path` class). The generated claim `data-app-0` therefore sat `Pending`:

```
Warning ProvisioningFailed pvc/data-app-0  storageclass.storage.k8s.io "fast-ssd" not found
Warning FailedScheduling  pod/app-0     0/1 nodes are available: pod has unbound immediate PersistentVolumeClaims.
```

Older events for the same claim name showed it provisioning successfully through `standard`/local-path, i.e. the template had been pointed at a class that does not exist in this cluster.

## Change made (smallest repair to the owning resource)

`volumeClaimTemplates` is immutable, so the fix was applied by recreating the StatefulSet from its own manifest with only that field changed:

1. Scaled `app` to 0 to release the pod holding the claim.
2. Deleted the stale unbound claim `data-app-0` (it had no volume, no capacity, no consumer — nothing lost).
3. Deleted the STS with `--cascade=orphan` (no pods existed) and re-applied the identical spec except `storageClassName: fast-ssd` → `storageClassName: standard`, `replicas: 2`.

Untouched/verified intact: `persistentVolumeClaimRetentionPolicy {whenDeleted: Retain, whenScaled: Retain}`, `podManagementPolicy: OrderedReady`, `serviceName: app`, `revisionHistoryLimit: 10`, pod template/image `nginx:1.31.5`, 1Gi RWO request, the headless Service `app`, and the existing `standard` StorageClass (no new cluster storage config invented).

## Verified outcome (fresh observations)

- Claims bound: `data-app-0` → `pvc-8f683495…`, `data-app-1` → `pvc-3d7e4123…`, both `phase=Bound`, `sc=standard`, `cap=1Gi`; both PVs `Bound` to their claims.
- Controller counters match request: `desired=2 current=2 ready=2 updated=2 available=2 observedGeneration=1/1`, `READY 2/2`.
- Pods actually running and using the storage: `app-0` and `app-1` `Running`, `ready=true`, `restarts=0`, `image=docker.io/library/nginx:1.31.5`; a write/read probe inside each pod succeeded (`hello-from-app-0`, `hello-from-app-1`) with `/data` mounted on the provisioned local-path filesystem.
- Latest events are all `Normal` (`ProvisioningSucceeded`, `Scheduled`, `Started`); the only `Warning` events are the historical pre-repair `fast-ssd`/unbound-claim ones. State re-checked ~25s later with no restarts or new failures.
