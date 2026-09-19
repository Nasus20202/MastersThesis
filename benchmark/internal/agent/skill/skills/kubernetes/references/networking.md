# Kubernetes Networking

Kubernetes defines a Pod network in which Pods receive cluster addresses and communicate through the cluster networking implementation. NetworkPolicy, CNI behavior, and external infrastructure can restrict or alter reachability.

A Service provides a stable virtual endpoint for a changing backend set. For selector-based Services, the selector matches Pod labels and the control plane represents selected backends in EndpointSlices. Service `port` is the port exposed by the Service, while `targetPort` identifies the backend port.

EndpointSlices carry endpoint addresses and conditions such as `ready`, `serving`, and `terminating`. Service proxies normally use ready endpoints for traffic; features such as `publishNotReadyAddresses` can intentionally expose endpoints before readiness.

Service types describe exposure: `ClusterIP` provides an internal virtual IP, `NodePort` allocates a node port, and `LoadBalancer` requests external load-balancer integration where supported.

Cluster DNS commonly resolves Service names such as `service.namespace.svc.cluster.local`. Ingress and Gateway resources describe higher-level routing to Services and require an installed controller or implementation.

NetworkPolicy controls allowed ingress and egress for selected Pods. When no policy selects a Pod for a direction, that direction is not isolated by NetworkPolicy. Once policies isolate a direction, allowed peers and ports are the union of applicable policy rules.