---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-volumeclaim/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-volumeclaim · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A StatefulSet is not starting because its volume claim template cannot provision storage. Restore the StatefulSet to a healthy state with all of its claims bound.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Components > Volume Claim Templates, Pod Identity > Stable Storage

## Write commands

1. `kubectl patch sts app --type=json -p '[{"op":"replace","path":"/spec/volumeClaimTemplates/0/spec/storageClassName","value":"standard"}]' 2>`
2. `kubectl delete sts app --cascade=orphan --wait=true`
3. `kubectl delete pod app-0 --wait=false`
4. `kubectl delete pvc data-app-0 --wait=true`
5. `kubectl apply -f app-sts-fixed.json`

## Final report

The issue is resolved. Here's what I found and did.

## Diagnosis

- `statefulset/app` (namespace `default`, 2 replicas) was stuck at `0/2`; `data-app-0` was `Pending`.
- `kubectl describe pvc data-app-0` and the persistentvolume-controller events showed the real cause:
  `Warning ProvisioningFailed ... storageclass.storage.k8s.io "fast-ssd" not found`
- The cluster has exactly one StorageClass: `standard` (`rancher.io/local-path`, marked default, `WaitForFirstConsumer`). No `fast-ssd` class exists anywhere, and no PVs existed.
- The fault lived in the StatefulSet's claim template: `spec.volumeClaimTemplates[0].spec.storageClassName: fast-ssd`.
- An in-place patch was rejected by the API server: `spec.volumeClaimTemplates: ... field is immutable` (server v1.37.0). The existing Pending PVC was equally unrepairable since a PVC's `storageClassName` is immutable too.

## Change (smallest justified)

The StatefulSet is the persistent resource owning the faulty state, so I fixed its template rather than inventing a new `fast-ssd` class:

1. Captured the live StatefulSet spec and changed **only** `volumeClaimTemplates[0].spec.storageClassName` from `fast-ssd` to the existing default class `standard`.
2. Since templates are immutable, replaced the StatefulSet (`delete --cascade=orphan` → `apply` of the corrected spec) and removed the two transient objects stuck with the bad class (Pending pod `app-0` and its `fast-ssd` PVC) so the new template could create claims with the valid class.
3. Everything else was preserved: replicas=2, `serviceName: app`, selector/labels, `nginx:1.31.5` image with `/data` mount, ports, `OrderedReady`, rolling-update strategy, `persistentVolumeClaimRetentionPolicy` Retain/Retain, 1Gi RWO requests. A spec diff against the original shows only `storageClassName: fast-ssd` → `standard` (plus the harmless kubectl last-applied annotation).

## Verified outcome (fresh observations)

- StatefulSet counters all match the requested 2: `replicas=2 current=2 updated=2 ready=2 available=2`, display `2/2`.
- Both claims bound: `data-app-0 → pvc-f577a7ad…` (1Gi RWO, `standard`), `data-app-1 → pvc-77431f8e…` (1Gi RWO, `standard`); corresponding PVs are `Bound`; provisioning events on the new claims are `ProvisioningSucceeded` with no `fast-ssd` errors.
- Pods `app-0` and `app-1` are `1/1 Running`, owned by `StatefulSet/app`.
- Storage is actually usable, not just "Bound": writing to `/data/probe.txt` inside both pods succeeded (`hello-from-app-0`, `hello-from-app-1`), with `/data` mounted read-write.
