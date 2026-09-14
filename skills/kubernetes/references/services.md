# Kubernetes Services and EndpointSlices

Trace service delivery as a chain: Service selector → matching Pod labels →
EndpointSlices → Service port and target port → ready backend. A failure at
any link can look like an application failure.

- Compare the Service selector with the exact labels on candidate Pods in the
  same namespace. A selector mismatch yields no selected backends.
- Compare `spec.ports[].port`, `targetPort` and protocol with the container's
  listening port and protocol. A Service port is not automatically the
  container port.
- Inspect EndpointSlices for addresses, conditions such as ready/serving and
  the referenced port. EndpointSlices are the observed routing data, not just
  a restatement of the selector.
- Account for readiness and terminating behavior when interpreting endpoints.
  Do not infer availability from Pod phase alone.

Preserve unrelated selectors and ports. After a change, re-check selected
endpoints and a bounded request or equivalent observable behavior.
