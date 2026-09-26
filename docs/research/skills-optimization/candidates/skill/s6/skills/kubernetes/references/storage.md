# Kubernetes Storage

Containers have ephemeral filesystems unless data is placed on a volume. Pod volumes live according to their volume type; PersistentVolumes and PersistentVolumeClaims represent durable storage independently of a particular Pod.

A PersistentVolumeClaim declares requested capacity, access modes, volume mode, and usually a StorageClass. A PersistentVolume represents storage made available to the cluster. Binding connects a claim to a compatible volume. With dynamic provisioning, a StorageClass provisioner can create storage for a claim.

Binding, attachment, and mounting are separate stages. A claim can be bound while later attachment or mount operations still fail. Storage topology and access modes can also constrain which nodes can use a volume.

Reclaim policy controls what happens to the underlying storage after a PersistentVolume is released. `Retain` preserves the storage for manual handling, while `Delete` allows the backing storage to be removed by the provisioner.

Stateful workloads often depend on stable claim identity. Replacing a claim or volume with a new empty resource is not equivalent to preserving the original data.
