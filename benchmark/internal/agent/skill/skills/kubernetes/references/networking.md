# Kubernetes Networking

Kubernetes defines a Pod network in which Pods receive cluster addresses and communicate through the cluster networking implementation. NetworkPolicy, CNI behavior, and external infrastructure can restrict or alter reachability.

A Service provides a stable virtual endpoint for a changing backend set. For selector-based Services, the selector matches Pod labels and the control plane represents selected backends in EndpointSlices. Service `port` is the port exposed by the Service, while `targetPort` identifies the backend port.

EndpointSlices carry endpoint addresses and conditions such as `ready`, `serving`, and `terminating`. Service proxies normally use ready endpoints for traffic; features such as `publishNotReadyAddresses` can intentionally expose endpoints before readiness.

Service types describe exposure: `ClusterIP` provides an internal virtual IP, `NodePort` allocates a node port, and `LoadBalancer` requests external load-balancer integration where supported.

Cluster DNS commonly resolves Service names such as `service.namespace.svc.cluster.local`. Ingress and Gateway resources describe higher-level routing to Services and require an installed controller or implementation.

NetworkPolicy controls allowed ingress and egress for selected Pods. When no policy selects a Pod for a direction, that direction is not isolated by NetworkPolicy. Once policies isolate a direction, allowed peers and ports are the union of applicable policy rules.

A policy peer selects by `podSelector`, `namespaceSelector`, or `ipBlock`. `ipBlock` matches the destination Pod or endpoint addresses that traffic is delivered to, after Service address translation, so listing a Service's virtual `ClusterIP` in `ipBlock` does not reliably permit traffic to that Service. To allow traffic to a Service, select its backend Pods with `podSelector` together with their namespace with `namespaceSelector`. To allow cluster DNS, permit egress on port 53 for UDP and TCP to the DNS Pods in `kube-system`, selected with `namespaceSelector` and `podSelector` rather than a broad address range. An empty `podSelector` selects every Pod in the policy's namespace.

An Ingress routes by `spec.ingressClassName`, `spec.rules[].host`, and `http.paths[].path` with `pathType`, forwarding to a Service backend by name and port. Gateway API splits this across resources: a `GatewayClass` names an implementation, a `Gateway` declares listeners with a name, port, and protocol, and route resources such as `HTTPRoute` attach to it. An `HTTPRoute` attaches through `spec.parentRefs` (the Gateway name and an optional `sectionName` naming a listener), matches `hostnames` and `rules` with path matches, and forwards to `backendRefs` naming a Service and port. A route affects traffic only once a Gateway listener admits it, so check attachment and acceptance conditions in addition to the objects existing.
