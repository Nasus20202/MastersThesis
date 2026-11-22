---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-20-57-03-137Z/readiness-probe-breaks-endpoints/prompt/003.json
verdict: accepted
reason: 'Repoints the readiness path to / and the reported endpoints follow from the pods becoming ready.'
---
# readiness-probe-breaks-endpoints · deepseek-flash · run-2026-11-21-20-57-03-137Z · attempt 3

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:
- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Inherited stop

The first 19 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 1
stdout:

stderr:
Error from server (BadRequest): invalid JSON patch

Its writes:

1. `kubectl edit deployment app -n default`
2. `kubectl delete pod app-648b6f877-57d5s -n default`
3. `kubectl rollout restart deployment app -n default`
4. `kubectl patch deployment app -n default --type=merge -p '{"spec": {"template": {"spec": {"containers": [{"name": "app", "readinessProbe": {"httpGet": {"path": "/readyz", "port": 80}, "failureThreshold": 0}}}}}}'`

## Teacher's write commands

1. `kubectl patch deployment app -n default --type=json -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/readinessProbe/httpGet/path", "value": "/"}]'`

## Final report

Based on the investigation, the application pods were running but were marked as **not ready** because the Kubernetes Readiness Probe was failing. The root cause was that the readiness probe was configured to check the `/readyz` path, which did not exist in the `nginx:1.31.5` container image, causing the probe to return a `404` error and keep the pods in a `0/1 Running` state.

As a result, the Service had no endpoints and was routing no traffic.

### Resolution

I applied the smallest justified change to the owning resource (the `app` Deployment) by updating the readiness probe to check a valid path (`/`) that the application actually serves. The original probe configuration was:

```yaml
readinessProbe:
  httpGet:
    path: /readyz
    port: 80
```

I patched it using:

```bash
kubectl patch deployment app -n default --type=json -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/readinessProbe/httpGet/path", "value": "/"}]'
```

### Verification

1. **Pod Readiness:** The deployment successfully rolled out, and all pods are now healthy.
   * **Requested replicas:** `3`
   * **Updated replicas:** `3`
   * **Ready replicas:** `3`
   * **Available replicas:** `3`

2. **Service Endpoints:** The Service `app` now correctly resolves to the running pods:
   ```
   kubectl get endpoints app -n default
   NAME   ENDPOINTS
   app    10.244.0.10:80,10.244.0.11:80,10.244.0.9:80
   ```
   The presence of all three pod IPs in the Endpoints resource proves the Service is actively routing traffic to the application pods. The application is now available through its Service.
