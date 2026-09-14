# Kubernetes storage

Trace persistent storage as PVC → PV → StorageClass or provisioner → CSI
attach/mount path → filesystem visible to the container. Each link can fail
independently, and a bound claim does not prove that a Pod mounted or can use
the volume.

- For a Pending PVC, check requested capacity, access modes, `storageClassName`,
  topology constraints and provisioner events. Check namespace scope for the
  claim and cluster scope for the PV and StorageClass.
- For a bound claim whose Pod cannot start, inspect Pod events for attach,
  mount, device, projection and filesystem errors. Distinguish a controller or
  CSI problem from an in-container path or permission problem.
- Compare the Pod's volume and volumeMount names, subpaths, read-only intent and
  referenced claim exactly. Preserve data and do not recreate a volume to hide
  an uncertain diagnosis.
- Account for access modes, reclaim policy and expansion support before changing
  a claim or PV. Existing data and other consumers may constrain a repair.
- Avoid printing Secret-backed storage credentials or mounted sensitive data.

Verify both attachment and the application-visible path after a change. A
successful claim bind or mount event is only an intermediate observation.
