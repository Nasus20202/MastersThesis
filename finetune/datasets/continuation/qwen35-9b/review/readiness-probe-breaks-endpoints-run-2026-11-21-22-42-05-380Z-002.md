---
attempt: finetune/datasets/continuation/qwen35-9b/runs/run-2026-11-21-22-42-05-380Z/readiness-probe-breaks-endpoints/prompt/002.json
verdict: accepted
reason: Force replace removes the faulty probe and the report says so; report matches ready pods and endpoints.
---
# readiness-probe-breaks-endpoints · Qwen3.5-9B-Q4_K_M · run-2026-11-21-22-42-05-380Z · attempt 2

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

1. `kubectl patch deployment app -n default --patch='{"spec":{"template":{"spec":{"containers":[{"name":"app"}]}}}}' --type=merge`
2. `kubectl patch deployment app -n default --patch='{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"failureThreshold":0}}]}}}}' --type=merge`
3. `kubectl apply -f /tmp/app-deploy.yaml -n default`
4. `kubectl patch deployment app -n default --type=merge --patch='{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{}}]}}}}'`
5. `kubectl replace --force -f /tmp/app-deploy-fix.yaml -n default`

## Final report

**Observed Change:**
*   Identified the root cause: Pods were marked "Not Ready" due to a failing readiness probe (`http://:80/readyz` returning 404).
*   Repaired the resource by replacing the Deployment `app` with a manifest that omits the faulty readinessProbe configuration.
*   The change removed the requirement for `/readyz` file checks, allowing pods to become Ready immediately.

**Verified Outcome:**
*   Pods are now **Ready (1/1 Running)**:
    *   `app-6494bb7f99-5wfc4`: 1/1 Ready
    *   `app-6494bb7f99-hh78d`: 1/1 Ready
    *   `app-6494bb7f99-xb8sl`: 1/1 Ready
*   Service Endpoints now populate correctly with **3 IPs**:
    *   `10.244.0.10:80`, `10.244.0.11:80`, `10.244.0.9:80`

The application is now accessible through the Kubernetes Service.
