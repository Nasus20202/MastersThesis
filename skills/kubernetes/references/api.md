# Kubernetes API objects

Use the API object's identity and scope before interpreting its fields. The
combination of `apiVersion`, `kind`, namespace when applicable and name
identifies the requested object; `uid` distinguishes one incarnation from an
older object with the same name.

- Treat `spec` as desired state and `status` and conditions as controller- or
  component-reported observed state. A missing or stale status is evidence about
  convergence, not proof that the desired state is invalid.
- Check `metadata.generation` and, when a controller reports it,
  `observedGeneration` to relate a status update to a spec change. Do not use a
  `resourceVersion` as a durable application-level version.
- Verify namespace scope, API group/version and field names from the live API.
  A familiar kind can have different versions or be cluster-scoped.
- Follow labels and selectors, owner references and named references to other
  objects. Check the referenced object's namespace and exact key or port rather
  than assuming names are globally unique.
- Separate an API or admission rejection from an accepted object that has not
  converged. Preserve server-side defaults and fields unrelated to the change.

Use the object's status, events and related resources to determine which
component should have acted next. Avoid treating a successful API write as
evidence that the requested behavior is already present.
