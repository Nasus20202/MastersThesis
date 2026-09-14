# Kubernetes configuration

Configuration commonly reaches a Pod through its template, environment
variables, mounted ConfigMaps or Secrets, projected or downward-API volumes,
and admission defaults. Trace the reference from the running Pod back to the
owning template and the source object.

- Verify names, namespaces, keys and mount paths exactly; a reference can be
  validly named but still point to the wrong object or key.
- Compare the effective Pod specification with the intended template and
  inspect events for mount or projection failures.
- Distinguish values copied into the Pod at creation from mounted data that may
  be refreshed. An environment variable is not updated merely because its
  source object changed.
- Remember that changing a ConfigMap or Secret does not necessarily recreate
  or restart existing containers. Confirm how the consuming workload observes
  updates before choosing a rollout action.
- Avoid printing secret values. Preserve unrelated keys and settings when
  updating configuration.

Verify the effective configuration and the resulting behavior with fresh,
bounded observations.
