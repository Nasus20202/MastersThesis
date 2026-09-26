---
name: networking
description: Repairing Service reachability, NetworkPolicy, DNS, Ingress and Gateway routing from the client's path to the server.
---

# Networking

Work hop by hop from the client to the server and change nothing until a check shows the broken hop.

## Service reachability

- Confirm the Service has ready endpoints. If not, the fault is behind the Service: Pod readiness or the Service `selector`.
- With ready endpoints, compare `spec.ports[].targetPort` with the container `containerPort` and named port. A Service whose `targetPort` does not match the container port has ready endpoints and still fails.
- Compare the Service `selector` with the Pod labels; a selector matching no Pod leaves endpoints empty.
- Prove the fix with the actual request from a client Pod, for example `kubectl exec <client> -- wget -qO- http://<service>`.

## NetworkPolicy

- Default deny blocks a direction unless a rule allows it. Add only the required peer and port.
- A peer selects Pods with `podSelector`, namespaces with `namespaceSelector`, or addresses with `ipBlock`. `ipBlock` matches endpoint addresses after Service translation, so a Service `ClusterIP` in `ipBlock` does not reliably permit traffic to the backends. To allow a Service, select its backend Pods with `podSelector`.
- To allow cluster DNS, permit egress on port 53 for UDP and TCP to the DNS Pods in `kube-system` (`namespaceSelector`, optionally `podSelector`), not a broad range.
- Verify both directions: the allowed client reaches the server, and a disallowed client or port still fails.

## DNS

- Use the full name `<service>.<namespace>.svc.cluster.local` to remove search-path ambiguity, and remember a headless Service resolves to its Pod IPs.
- Test with a query from a Pod (`nslookup`/`getent hosts`) rather than assuming.

## Ingress and Gateway

- An Ingress needs its `ingressClassName`, a rule host and path with `pathType`, and a backend Service name and port.
- Gateway API splits configuration: a `GatewayClass`, a `Gateway` with listeners (name, port, protocol), and an `HTTPRoute` attaching through `spec.parentRefs`, matching `hostnames`/paths and forwarding to `backendRefs`.
- A route takes effect only once a Gateway listener admits it; check the attachment and accepted conditions, not just that the objects exist.

## Verify

Make a fresh request through the path that was reported broken and record whether it succeeds; where isolation is required, record that the disallowed path still fails.
