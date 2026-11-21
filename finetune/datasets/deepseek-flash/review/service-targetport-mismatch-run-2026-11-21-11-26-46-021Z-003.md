---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/service-targetport-mismatch/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# service-targetport-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

The application Pods are healthy and the Service has ready endpoints, but the application is unavailable through the Service. Restore service availability.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Defining a Service

## Write commands

1. `kubectl -n default patch svc app --type merge -p '{"spec":{"ports":[{"name":"http","port":80,"protocol":"TCP","targetPort":80}]}}'`

## Final report

Service availability restored.

**Root cause (observed):** The `default/app` Service had `port: 80` → `targetPort: 8080`, but the nginx Pods actually listen on container port 80. Evidence before the change:

- `curl http://10.244.0.4:80/` → `http=200`, `curl http://10.244.0.4:8080/` → `curl: (7) Failed to connect`
- `curl http://10.96.122.41/` (Service ClusterIP) → `curl: (7) Failed to connect to 10.96.122.41 port 80`
- Endpoints/EndpointSlice were populated (all 3 Pod IPs) but advertised port 8080, so kube-proxy forwarded to a closed port.

**Fix (smallest persistent change):** Patched only the Service's port definition — no Deployment, Pod, or network changes:

```
spec.ports: [{name: http, port: 80, protocol: TCP, targetPort: 8080}] → targetPort: 80
```

**Verification (fresh observations from a Pod in the cluster):**

- EndpointSlice `app-jtn9m` now advertises `{"name":"http","port":80}` with all 3 ready addresses: 10.244.0.4, 10.244.0.2, 10.244.0.6.
- Three consecutive request rounds through the Service: `via ClusterIP http=200`, `via DNS http=200` (app.default.svc.cluster.local), and short name `http://app/` returned the nginx `<!DOCTYPE html>` body.
- Service still ClusterIP type, port 80/TCP, selector `app=app`; Deployment remains 3/3 ready and available, no Pod restarts.
