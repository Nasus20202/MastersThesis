---
name: connectivity
description: Diagnosing and repairing Service reachability, NetworkPolicy, DNS, Ingress and Gateway routing.
---

# Connectivity

Work from the client's path to the server and check each hop; do not change a resource until a check shows it is the broken hop.

## Service reachability

- Confirm the Service has ready endpoints. If not, the fault is behind the Service: Pod readiness or the selector.
- When endpoints are ready, compare `spec.ports[].targetPort` with the container `containerPort` and named port. A Service whose `targetPort` does not match the container port has ready endpoints and still fails.
- Compare the Service `selector` with the Pod labels; a selector that matches no Pod leaves endpoints empty.
- Test the actual request from a client Pod (`kubectl exec <client> -- wget -qO- http://<service>` or `curl`). Ready endpoints do not prove end-to-end success.
- Check NetworkPolicy only after the Service, endpoints and policy paths are otherwise consistent.

## NetworkPolicy

- A peer selects Pods with `podSelector`, namespaces with `namespaceSelector`, or addresses with `ipBlock`. `ipBlock` matches the destination Pod or endpoint addresses after Service translation, so a Service `ClusterIP` in `ipBlock` does not reliably permit traffic to its backends.
- To allow a Service, select its backend Pods with `podSelector`, combined with `namespaceSelector` when the client is in another namespace.
- To allow cluster DNS, permit egress on port 53 for UDP and TCP to the DNS Pods in `kube-system` with `namespaceSelector` (and optionally `podSelector`), not a broad address range.
- Default deny means a direction is blocked unless a policy allows it. Add only the required peer and port; keep every other path blocked and prove it.
- Verify both directions: the allowed client reaches the server, and a disallowed client or port still fails.

## Ingress and Gateway

- An Ingress needs its `ingressClassName`, a rule host and path with `pathType`, and a backend Service name and port.
- Gateway API splits the configuration: a `GatewayClass`, a `Gateway` with listeners (name, port, protocol), and an `HTTPRoute` that attaches through `spec.parentRefs` (Gateway name, optional `sectionName`), matches `hostnames` and path rules, and forwards to `backendRefs` naming a Service and port.
- A route only affects traffic once a Gateway listener admits it; check the attachment and accepted conditions as well as the objects existing.
