# Kubernetes API extensions

Custom resources extend the API, while a CRD defines their schema and a
controller or operator usually reconciles their desired state. Installing a
CRD does not prove that its controller is installed, healthy or watching the
requested version.

- Confirm the API group, version, kind, plural resource name and namespace
  scope. Check discovery and the live CRD schema before assuming a field exists.
- Inspect the custom resource's spec, status, conditions, generation and
  owner/reference edges. Use controller events and logs when available to
  distinguish invalid configuration from a stalled reconciler.
- Follow generated Deployments, Pods, Services, Jobs or other resources back to
  the custom resource. Repair the authoritative custom resource or controller
  configuration, not a generated child, when ownership evidence supports it.
- Treat conversion, webhook and admission failures as API-path problems rather
  than application failures. Verify the served and storage versions before
  changing an object across versions.
- Preserve unknown fields, status ownership and unrelated custom-resource
  settings. Do not delete a CRD or custom resource to force reconciliation
  without understanding the dependent resources and data consequences.

Verify both the custom resource's reported status and the behavior of its
generated resources after a change.
