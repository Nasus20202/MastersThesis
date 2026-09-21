# Kubernetes Extensions

CustomResourceDefinitions extend the API with new object types: a declared group, version, names, and schema, after which custom objects behave like native ones for storage, watch, and authorization. The definition only declares the type; a controller (often called an operator) implements its behavior by watching those objects and reconciling the world toward their spec.

Operators are ordinary controllers for custom types: they read the custom object's spec, create or update owned resources (Deployments, Services, ConfigMaps), and report back through the custom object's status and events. Ownership and generation semantics apply unchanged: the operator's template owns the generated children.

Admission webhooks extend policy: mutating webhooks alter admitted objects and validating webhooks accept or reject them, scoped by namespace selectors, object selectors, operations, and failure policy. A failing webhook with `Fail` policy rejects every matching write, including unrelated ones. Aggregated API servers extend the API surface itself with additional endpoints behind the same authentication and authorization.

Extensions move behavior out of static manifests into running code, so a custom resource whose controller is absent, outdated, or denied permission keeps its spec forever unreconciled.
