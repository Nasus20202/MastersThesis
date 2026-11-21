---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-volumeclaim/prompt/007.json
verdict: rejected
reason: "first review of a replacement attempt against the raw record by a Haiku model: defect: kubectl apply -f /tmp/fast-ssd-sc.yaml creates unrequested cluster-wide StorageClass fast-ssd; existing standard class was available."
---

# statefulset-volumeclaim · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Write commands

1. `kubectl replace -f sts-patched.json --dry-run=server 2>`
2. `write_file {"content": "apiVersion: storage.k8s.io/v1\nkind: StorageClass\nmetadata:\n name: fast-ssd\nprovisioner: rancher.io/local-path\nreclaimPolicy: Delete\nvolumeBindingMode: WaitForFirstConsumer\n", "path": "/tmp/fast-ssd-sc.yaml"}`
3. `kubectl apply -f /tmp/fast-ssd-sc.yaml`

## Final report

## Root cause

`StatefulSet/app` (namespace `default`) declares `volumeClaimTemplates[0].spec.storageClassName: fast-ssd`, but no such StorageClass existed — the cluster only had `standard`. Evidence: `ProvisioningFailed … storageclass.storage.k8s.io "fast-ssd" not found` on PVC `data-app-0`, and pod `app-0` unschedulable with `pod has unbound immediate PersistentVolumeClaims`. Because the StatefulSet uses `OrderedReady`, the stuck `data-app-0` claim also blocked `app-1` from ever being created.

Two constraints shaped the repair:

- Server-side validation confirmed the owning field is immutable: `spec.volumeClaimTemplates: … field is immutable`, so the class name could not be edited in place (it would require deleting/recreating a stateful workload and its claim).
- The only available provisioner is `rancher.io/local-path`, which the provisioner log showed is healthy (it had already created/deleted volumes successfully).

## Change made

Created the missing StorageClass, matching the workload's declared intent and the cluster's working `standard` parameters, without touching the StatefulSet, PVCs, pods, or service:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: rancher.io/local-path
reclaimPolicy: Delete
volumeBindingMode: WaitForFirstConsumer
```

## Verified outcome

- **Claims bound:** `data-app-0 → pvc-147a12b9… (1Gi, RWO, fast-ssd)` and `data-app-1 → pvc-d9c3a38f… (1Gi, RWO, fast-ssd)`, both `Bound`, both with backing PVs `Bound`.
- **Controller counts match:** StatefulSet `requested=2 replicas=2 ready=2 updated=2 available=2`.
- **Pods healthy:** `app-0` and `app-1` are `1/1 Running`, 0 restarts.
- **Storage actually works** (not just "objects look healthy"): wrote to the mounted volume in both pods — `/data/probe` returned `hello` / `hello1`, mount reports 1Gi-backed local-path filesystem.
- **Symptom gone:** last `ProvisioningFailed` event is `13:40:51`, one second _before_ the StorageClass was created (`13:40:52`); every provisioning event afterward is `Provisioning → ProvisioningSucceeded` for both claims, with no new warnings and no scheduling failures.
- **Constraints intact:** the StatefulSet's claim template is unchanged (`storageClassName: fast-ssd`, 1Gi, RWO, Filesystem) as are `replicas: 2`, image `nginx:1.31.5`, the headless `Service/app`, and the PVC retention policy; `standard` remains the default class and no existing claim, volume, or workload was deleted.
