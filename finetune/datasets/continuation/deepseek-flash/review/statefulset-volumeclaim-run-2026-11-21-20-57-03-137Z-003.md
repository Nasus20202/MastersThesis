---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-20-57-03-137Z/statefulset-volumeclaim/prompt/003.json
verdict: accepted
reason: 'Deletes and recreates with standard storage and reports the pending PVC removal.'
---
# statefulset-volumeclaim · deepseek-flash · run-2026-11-21-20-57-03-137Z · attempt 3

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:
- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Inherited stop

The first 23 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 1
stdout:

stderr:
The StatefulSet "app" is invalid: spec.volumeClaimTemplates: Invalid value: [{"name":"data","labels":{"app":"app"},"Spec":{"AccessModes":["ReadWriteOnce"],"Selector":null,"Resources":{"Limits":null,"Requests":{"storage":"1Gi"}},"VolumeName":"","StorageClassName":"standard","VolumeMode":"Filesystem","DataSource":null,"DataSourceRef":null,"VolumeAttributesClassName":null},"Status":{"Phase":"Pending","AccessModes":null,"Capacity":null,"Conditions":null,"AllocatedResources":null,"AllocatedResourceStatuses":null,"CurrentVolumeAttributesClassName":null,"ModifyVolumeStatus":null,"HealthStatus":null}}]: field is immutable

Its writes:

1. `kubectl edit statefulset app -n default`
2. `kubectl patch statefulset app -n default --type='json' -p='[{"op": "replace", "path": "/spec/volumeClaimTemplates/0/spec/storageClassName", "value": "standard"}]'`
3. `write_file {"content": "apiVersion: apps/v1\nkind: StatefulSet\nmetadata:\n creationTimestamp: \"2026-10-10T00:23:47Z\"\n generation: 1\n labels:\n app: app\n name: app\n namespace: default\n resourceVersion: \"648\"\n uid: 2d91b1a2-5e0e-485b-a90b-5b8c037b3365\nspec:\n persistentVolumeClaimRetentionPolicy:\n whenDeleted: Retain\n whenScaled: Retain\n podManagementPolicy: OrderedReady\n replicas: 2\n revisionHistoryLimit: 10\n selector:\n matchLabels:\n app: app\n serviceName: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n name: app\n ports:\n - containerPort: 80\n name: http\n protocol: TCP\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n volumeMounts:\n - mountPath: /data\n name: data\n dnsPolicy: ClusterFirst\n restartPolicy: Always\n schedulerName: default-scheduler\n securityContext: {}\n terminationGracePeriodSeconds: 30\n updateStrategy:\n rollingUpdate:\n maxUnavailable: 1\n partition: 0\n type: RollingUpdate\n volumeClaimTemplates:\n - apiVersion: v1\n kind: PersistentVolumeClaim\n metadata:\n labels:\n app: app\n name: data\n spec:\n accessModes:\n - ReadWriteOnce\n resources:\n requests:\n storage: 1Gi\n storageClassName: standard\n volumeMode: Filesystem\n status:\n phase: Pending\nstatus:\n availableReplicas: 0\n collisionCount: 0\n currentReplicas: 1\n currentRevision: app-5f7c7cd965\n observedGeneration: 1\n replicas: 1\n updateRevision: app-5f7c7cd965\n updatedReplicas: 1", "path": "statefulset_app.yaml"}`
4. `kubectl apply -f statefulset_app.yaml -n default`
5. `kubectl patch statefulset app -n default --type='json' -p='[{"op": "replace", "path": "/spec/volumeClaimTemplates/0/spec/storageClassName", "value": "standard"}]'`

## Teacher's write commands

1. `kubectl delete statefulset app -n default`
2. `kubectl delete pvc data-app-0 -n default`
3. `kubectl apply -f statefulset_app.yaml -n default`

## Final report

The StatefulSet is now healthy with all its claims bound.

**Diagnosis:**
The PersistentVolumeClaim `data-app-0` was stuck in a `Pending` state with the following event:
```
storageclass.storage.k8s.io "fast-ssd" not found
```
The StatefulSet's `volumeClaimTemplates` requested the non-existent `fast-ssd` StorageClass. I confirmed that the only available StorageClass in the cluster is `standard`.

**Repair:**
Because the `spec.volumeClaimTemplates` field of a StatefulSet is immutable and rejected in-place patches, I corrected the resource by:
1.  Deleting the StatefulSet `app` (which had `persistentVolumeClaimRetentionPolicy: whenDeleted: Retain`, preserving the PVC).
2.  Deleting the unprovisionable `data-app-0` PVC.
3.  Re-creating the StatefulSet with the `volumeClaimTemplates` updated to use `storageClassName: standard`.

**Verified Outcome:**
- **StatefulSet:** `kubectl get statefulset app` shows `READY 2/2`.
- **PVCs:** Both `data-app-0` and `data-app-1` are `Bound` to `1Gi RWO` volumes using the `standard` StorageClass.
- **Pods:** `app-0` and `app-1` are both `1/1 Running`.
