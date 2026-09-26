---
name: workloads
description: Repairing rollouts, probes, crash loops, StatefulSets, headless Services and Jobs through the owning controller.
---

# Workloads

Repair the controller's Pod template, not the generated Pod. A replacement Pod has a random name and is recreated, so a change there is not a lasting fix.

## Crash and startup faults

- Read the container waiting or terminated reason and the previous log before changing anything.
- `CreateContainerConfigError` is a missing or wrong ConfigMap/Secret reference or an invalid security context. `ImagePullBackOff` is the image reference or pull secret. `OOMKilled` is the memory limit. A command or argument error is the container command. Repeated kills of a healthy container are the liveness probe.

## Probes

Choose the probe by its purpose; do not use one to fix another.

- **Readiness** gates Service endpoints. A wrong readiness probe leaves ready Pods at zero endpoints.
- **Liveness** restarts a hung container. Making it laxer to stop restarts hides a real fault.
- **Startup** protects slow-starting containers: while it runs, liveness and readiness are suspended. A slow app that is killed before it starts needs a `startupProbe` (with a suitable `failureThreshold`/`periodSeconds`), not a changed liveness or memory limit.

## Rollouts

- Compare `spec.replicas` with `.status.updatedReplicas`, `.status.readyReplicas` and `.status.availableReplicas`. They must all match before a rollout is complete.
- Read ReplicaSet and `ReplicaFailure` events and `kubectl rollout history` when the new revision does not start.
- **Bad release.** Restore the last good revision with `kubectl rollout undo deployment/<name>` or set the known-good image; do not change the command to disguise a bad image.
- Check for immutable fields (a Deployment `selector` cannot change) and for a new Pod that cannot schedule or start.

## StatefulSet and headless Service

- A StatefulSet manages stable Pod names and per-Pod storage through `volumeClaimTemplates`. Its governing Service is normally headless: `clusterIP: None`. Set the StatefulSet `serviceName` to that Service.
- Verify per-Pod DNS (`<pod>.<service>.<namespace>.svc`) and that each Pod is Ready.

## Jobs

A failed Job's Pod holds the error. Fix the command, image or dependency and keep `backoffLimit` unless the task changes it. Distinguish a completed Job from an exhausted one through its conditions.
