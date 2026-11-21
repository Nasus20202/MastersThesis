---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-volumeclaim/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-volumeclaim · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Write commands

1. `kubectl patch sts app -n default --type=json -p '[{"op":"replace","path":"/spec/volumeClaimTemplates/0/spec/storageClassName","value":"standard"}]' 2>`
2. `kubectl delete sts app -n default --wait=true --timeout=60s 2>`
3. `kubectl delete pvc data-app-0 -n default --wait=true --timeout=60s 2>`
4. `kubectl create -f /tmp/sts-new.json 2>`

## Final report

## Root cause

The StatefulSet `default/app` had a `volumeClaimTemplate` requesting `storageClassName: fast-ssd`, but no such StorageClass existed in the cluster — the only class was `standard` (`rancher.io/local-path`, default, `WaitForFirstConsumer`). The provisioner therefore failed:

```
Warning  ProvisioningFailed  persistentvolumeclaim/data-app-0
  storageclass.storage.k8s.io "fast-ssd" not found
pod/app-0  0/1 nodes are available: pod has unbound immediate PersistentVolumeClaims
```

## Diagnosis path

1. `kubectl get sts -A` → `default/app 0/2`; `kubectl get pvc` → `data-app-0 Pending` with `STORAGECLASS=fast-ssd`.
2. `kubectl get sc` → only `standard` (rancher.io/local-path) exists; no CSI driver that could back a `fast-ssd` class is installed.
3. Attempted a minimal in-place patch of the template's class — rejected by the API server: `spec.volumeClaimTemplates: Invalid value ... field is immutable`. So the owning object had to be recreated with the corrected template.

## Change made

Repaired the owning resource (StatefulSet) with the smallest justified change: the volume claim template's `storageClassName` was pointed at the existing default class `standard`. Nothing else was altered — `replicas: 2`, `serviceName: app`, `image: nginx:1.31.5`, container port/mounts, `podManagementPolicy`, `updateStrategy`, `selector`, labels and the `persistentVolumeClaimRetentionPolicy` (`whenDeleted/whenScaled: Retain`) were carried over verbatim from the live object. The unbound `Pending` PVC `data-app-0` (no PV, no data, class field immutable, created only seconds earlier by the same broken template) was deleted so the recreated controller could issue it correctly.

## Verification (fresh observations)

- Claims all bound, correct driver:
  - `data-app-0  Bound  pvc-68fc828e…  1Gi  RWO  standard`
  - `data-app-1  Bound  pvc-578caf07…  1Gi  RWO  standard`
- Controller counts match the requested replica count: `spec.replicas=2`, `current=2`, `updated=2`, `ready=2`, `available=2`, `observedGeneration=1 == generation=1`, `kubectl get sts` → `app 2/2`.
- Pods actually running and stable after a second check ~30s later: `app-0` and `app-1` both `Running`, `ready=true`, `RESTARTS=0`.
- Storage really works, not just "bound": `echo ok > /data/probe && cat /data/probe` returned `ok` inside both `app-0` and `app-1`.
- No new warnings: the only `Warning` events are pre-fix ones (`fast-ssd not found` last seen at the time of the deleted PVC); post-fix events are all `Normal` (`ProvisioningSucceeded`, `Scheduled`, `Started`).
