# Kubernetes Configuration

Workloads can consume configuration through environment variables, command arguments, mounted files, or projected volumes.

ConfigMaps store non-secret configuration data. Secrets store sensitive values but are not inherently confidential merely because they are base64-encoded; access control and storage encryption determine protection. Pod references to ConfigMaps and Secrets are namespaced and normally refer to objects in the same namespace as the Pod.

Environment variables sourced from a ConfigMap or Secret are resolved when a container starts and do not change inside an already-running process when the source object changes. Mounted ConfigMap and Secret volumes are updated asynchronously by the kubelet, but an application may still need to reload the file contents itself.

Immutable ConfigMaps and Secrets cannot have their data changed after creation. Workloads that need versioned configuration can reference objects with versioned names or change their Pod template when configuration changes.

A controller creates new Pods when its Pod template changes. Merely changing a referenced ConfigMap or Secret does not necessarily recreate Pods.
