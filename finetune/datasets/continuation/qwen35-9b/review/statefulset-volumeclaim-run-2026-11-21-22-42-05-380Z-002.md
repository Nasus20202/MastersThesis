---
attempt: finetune/datasets/continuation/qwen35-9b/runs/run-2026-11-21-22-42-05-380Z/statefulset-volumeclaim/prompt/002.json
verdict: accepted
reason: Force replace and PVC force-delete are disclosed; report matches both bound claims and 2/2 ready.
---
# statefulset-volumeclaim · Qwen3.5-9B-Q4_K_M · run-2026-11-21-22-42-05-380Z · attempt 2

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

1. `kubectl replace -f statefulset_app.yaml -n default --force`
2. `kubectl delete pvc data-app-0 -n default --grace-period=0 --force`

## Final report

The StatefulSet `app` has been restored to a healthy state. Here is the summary of the changes and verification:

**Diagnosis:**
The StatefulSet was failing to provision storage because it referenced a non-existent StorageClass `fast-ssd`. The only available StorageClass in the cluster was `standard`.

**Fixes Applied:**
1. **Replaced the StatefulSet definition** using `kubectl replace --force` with an updated version where `storageClassName` was changed from `fast-ssd` to `standard`. This was necessary because API patches for the `spec.volumeClaimTemplates` field were rejected as immutable.
2. **Force-deleted existing PVCs** (`data-app-0`) which were stuck in a `Pending` state bound to the non-existent `fast-ssd`. Deleting them allowed the StatefulSet to recreate them using the correct `standard` StorageClass.

**Verification:**
- **PVC Status**: Both `data-app-0` and `data-app-1` are now in `Bound` status using the `standard` StorageClass.
  - `kubectl get pvc -n default` shows:
    ```
    NAME         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    data-app-0   Bound    pvc-8e98ee76-eed2-4189-a42f-509c756af0bc   1Gi        RWO            standard       32s
    data-app-1   Bound    pvc-8e7f1f8e-ae6b-41c5-a887-c1188e018e4d   1Gi        RWO            standard       28s
    ```
- **StatefulSet Health**: All replicas are ready.
  - `kubectl get statefulset app -n default` shows: `READY 2/2`.

The originally reported symptom (StatefulSet not starting due to storage provisioning failure) has been resolved, with all volume claims bound and the cluster stable.
