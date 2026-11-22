---
attempt: finetune/datasets/continuation/qwen35-9b/runs/run-2026-11-21-22-42-05-380Z/statefulset-volumeclaim/prompt/003.json
verdict: accepted
reason: Template switched to the standard StorageClass with a force-deleted claim; report matches both bound claims and 2/2 ready.
---
# statefulset-volumeclaim · Qwen3.5-9B-Q4_K_M · run-2026-11-21-22-42-05-380Z · attempt 3

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

1. `kubectl patch pvc data-app-0 -n default --type='json' -p='[{"op": "replace", "path": "/spec/storageClassName", "value": "standard"}]'`
2. `write_file {"content": "apiVersion: apps/v1\nkind: StatefulSet\nmetadata:\n name: app\n namespace: default\n labels:\n app: app\nspec:\n replicas: 2\n serviceName: app\n selector:\n matchLabels:\n app: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - image: nginx:1.31.5\n name: app\n ports:\n - containerPort: 80\n protocol: TCP\n volumeMounts:\n - mountPath: /data\n name: data\n updateStrategy:\n type: RollingUpdate\n volumeClaimTemplates:\n - metadata:\n labels:\n app: app\n name: data\n spec:\n accessModes:\n - ReadWriteOnce\n resources:\n requests:\n storage: 1Gi\n storageClassName: standard\n volumeMode: Filesystem", "path": "statefulset_app.yaml"}`
3. `kubectl patch statefulset app -n default -p='[{"op": "replace", "path": "/spec/template/spec/volumes/0/storageClassName", "value": "standard"}]' --type json`
4. `kubectl apply -f statefulset_app.yaml -n default --force`
5. `kubectl delete pvc data-app-0 -n default --force --grace-period=0`

## Final report

The StatefulSet has been restored to a healthy state. The fix replaced the non-existent `fast-ssd` StorageClass with the available `standard` StorageClass in the volumeClaimTemplates, and force-deleted the stuck PVCs to allow recreation with the correct storage class.

**Verified Outcome:**
- StatefulSet `app` is **READY 2/2**.
- Both PVCs (`data-app-0`, `data-app-1`) are **Bound**.
- Both PVCs now use **StorageClass: standard**.
- All associated pods (`app-0`, `app-1`) are in **Running** status.

The StatefulSet is fully operational with all volume claims successfully bound.
