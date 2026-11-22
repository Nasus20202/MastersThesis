---
attempt: finetune/datasets/continuation/qwen35-9b/runs/run-2026-11-21-22-42-05-380Z/readiness-probe-breaks-endpoints/prompt/003.json
verdict: accepted
reason: Probe path changed to /; report matches the ready pods and populated endpoints.
---
# readiness-probe-breaks-endpoints · Qwen3.5-9B-Q4_K_M · run-2026-11-21-22-42-05-380Z · attempt 3

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

1. `kubectl patch deployment app -n default --patch '{"spec": {"template": {"spec": {"containers": [{"name": "app", "readinessProbe": null}]}}}}' --type=merge`
2. `kubectl patch deployment app -n default --patch '{"spec": {"template": {"spec": {"containers": [{"name": "app", "readinessProbe": {"httpGet": {"path": "/", "port": 80, "scheme": "HTTP"}, "failureThreshold": 3}}]}}}}' --type=merge`
3. `kubectl apply -f /tmp/app-deploy.yaml -n default`

## Final report

The issue has been identified and resolved.

**Diagnosis:**
- The application Pods were running but not ready, causing the Service to return traffic failures.
- `kubectl describe pod` showed `Ready: False` with a failing readiness probe: `httpGet: path: /readyz`.
- Logs confirmed the Nginx container returned HTTP 404 and "No such file or directory" for the `/readyz` endpoint.

**Fix Applied:**
- Modified the Deployment manifest to change the readiness probe path from `/readyz` to `/` (the default Nginx index), which is valid and served by the `nginx:1.31.5` image.

**Verified Outcome:**
- Pods are now ready (`READY 1/1` for all three replicas).
- Service Endpoints reflect the healthy pods, confirming service availability has been restored.
