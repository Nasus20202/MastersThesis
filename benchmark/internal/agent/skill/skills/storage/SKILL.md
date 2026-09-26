---
name: storage
description: Repairing PersistentVolumeClaims, volume mounts and shared volumes, including writable paths for a read-only root filesystem.
---

# Storage

Stop at the layer that is broken. A claim can be Bound while the Pod still fails to mount, and a mount can be correct while the application writes to a path that is not mounted.

## Claim and volume

- **Pending claim.** Compare the claim's `storageClassName`, `accessModes`, `resources.requests.storage` and `selector` with the available PersistentVolumes. A claim binds to a volume with a matching class, enough capacity, a compatible access mode and matching labels.
- **Bound but not mounting.** Check the volume's node affinity, the StorageClass `volumeBindingMode`, and events on the Pod. Fix the volume side, not the Pod's node.

## Mounting

- A container mount needs **both** a `spec.volumes` entry and a matching `volumeMounts` entry with the same `name`; adding only one leaves the mount unresolved. Apply the pair in one change.
- Strategic merge patch merges lists of containers by `name`, but the `volumeMounts` inside a container are replaced by the patch. Send the complete `volumeMounts` list for that container, and keep `name` as the merge key so the image is not dropped.
- **Shared volume.** Two containers share data only when the same volume name is mounted in both at the paths they use. A shared `emptyDir` is empty at Pod start; the reader must tolerate the writer not having written yet, or wait for the file.
- **Read-only root filesystem.** An application that writes under `/tmp`, `/var/cache`, `/var/run` or a service directory fails when the root filesystem is read-only. Add a writable `emptyDir` at the exact path the process writes and keep `readOnlyRootFilesystem: true`. Find the path from the log or crash reason instead of guessing.

## Verify

The claim is Bound, the Pod is Ready with no mount errors, and the intended read or write succeeds (for example, the file exists in the shared volume or the served content is the new one).
