---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/service-targetport-mismatch/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# service-targetport-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

The application Pods are healthy and the Service has ready endpoints, but the application is unavailable through the Service. Restore service availability.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Defining a Service

## Write commands

1. `kubectl run curltest --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'curl -s -m 5 -o /dev/null -w "svc:80 -> %{http_code}\n" http://app.default.svc.cluster.local/`
2. `kubectl patch svc app --type=json -p='[{"op":"replace","path":"/spec/ports/0/targetPort","value":80}]'`
3. `kubectl run curltest --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'for i in 1 2 3`

## Final report

Service availability restored.

## Root cause

The Service `default/app` forwarded traffic to `targetPort: 8080`, but the nginx Pods listen on port 80 (`containerPort: 80`). Endpoints were ready because readiness was based on Pod status, not on the port mapping.

Evidence before the fix:

- `kubectl exec deploy/app -- curl http://localhost:80/` → `200`; `http://localhost:8080/` → `000` (rc=7, connection refused)
- Request through the Service → `svc:80 -> 000` rc=7

## Change made

One-field patch to the owning Service (Deployment/Pods untouched, port 80 preserved):

```
kubectl patch svc app --type=json \
  -p='[{"op":"replace","path":"/spec/ports/0/targetPort","value":80}]'
```

## Verification (fresh observations)

- Service spec: `{"name":"http","port":80,"protocol":"TCP","targetPort":80}`
- EndpointSlice `app-6ctzd` port updated to `80`; all 3 endpoints ready:
  `10.244.0.7 ready=true, 10.244.0.4 ready=true, 10.244.0.6 ready=true`
- Three consecutive requests through the ClusterIP Service: `svc:80 -> 200` (x3)
- Controller unchanged and fully converged: Deployment `app` replicas 3, updated 3, ready 3, available 3 (3/3), all Pods `1/1 Running` with 0 restarts.
