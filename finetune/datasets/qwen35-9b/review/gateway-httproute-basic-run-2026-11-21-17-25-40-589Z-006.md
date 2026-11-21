---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-17-25-40-589Z/gateway-httproute-basic/prompt/006.json
verdict: rejected
reason: 'first-pass review by a Haiku model: reject: "http://[gateway-ip]:8000/app.example.com/app" misstates host routing | second pass against the raw record: defect: "External requests to http://[gateway-ip]:8000/app.example.com/app" puts the host in the path; no request was ever sent through the Gateway.'
---

# gateway-httproute-basic · Qwen3.5-9B-Q4_K_M · run-2026-11-21-17-25-40-589Z · attempt 6

## Task

The Service app in the default namespace is healthy. Expose it with the Gateway API by creating a Gateway named app-gateway (gatewayClassName traefik) with an HTTP listener named http on port 8000, and an HTTPRoute named app-route that routes host app.example.com and path /app (PathPrefix) to Service app on port 80. Requests are served through the traefik Service on port 80.

Documented in:

- content/en/docs/concepts/services-networking/gateway.md: Resource model > Gateway, Resource model > HTTPRoute

## Write commands

1. `kubectl apply -f - << 'EOF'`
2. `kubectl apply -f /tmp/gateway.yaml`
3. `kubectl apply -f /tmp/httproute.yaml`
4. `kubectl apply -f /tmp/httproute.yaml`

## Final report

All resources have been created and configured correctly. Here's the verification:

**Gateway (app-gateway):**

- Namespace: default
- gatewayClassName: traefik
- Listener: HTTP, name=http, port=8000
- Status: Accepted ✓, Programmed ✓
- Attached routes: 1

**HTTPRoute (app-route):**

- Namespace: default
- Hostnames: app.example.com ✓
- Path matching: PathPrefix /app ✓
- BackendRef: Service "app", port 80 ✓
- ParentRef: Gateway app-gateway, namespace default, port 8000 ✓
- Status: ResolvedRefs ✓, Accepted ✓

**Service (app):**

- Namespace: default
- Port: 80/TCP
- Status: Running and healthy ✓

Both the Gateway and HTTPRoute are in a healthy state with no errors. The traffic flow is configured as follows:

- External requests to `http://[gateway-ip]:8000/app.example.com/app` are routed through the Gateway listener on port 8000
- The HTTPRoute matches host `app.example.com` and path prefix `/app`
- Traffic is forwarded to Service `app` on port 80
