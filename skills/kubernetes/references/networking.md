# Kubernetes networking

Trace traffic through the actual path rather than treating a Pod, Service or
Ingress as the whole network. A typical path includes name resolution, a
Service and its EndpointSlices, policy enforcement, the backend listener and,
for north-south traffic, an Ingress or Gateway controller.

- Identify the source and destination namespaces, names, protocol and port.
  Compare DNS names and search domains with the intended Service and namespace.
- Use `services.md` for selector, ready-endpoint and `port`/`targetPort`
  analysis. A Service with endpoints can still fail if the backend is not
  listening on the referenced port or protocol.
- Inspect NetworkPolicy selection and ingress/egress direction. When a Pod is
  isolated in a direction, the connection must be allowed by the applicable
  policies for that direction; policy presence alone does not prove a denial.
- For Ingress, Gateway or related routes, verify the class or controller,
  listener, host/path match, backend reference and reported status. A route
  object existing does not prove that a controller programmed the dataplane.
- Separate dataplane symptoms from CNI, DNS, node, firewall and application
  evidence. Do not assume a particular CNI or load-balancer implementation.

After a change, re-check the relevant object status and a bounded request from
the same network context. Do not broaden policy or expose a backend merely to
make one probe succeed.
