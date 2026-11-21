---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/readiness-probe-breaks-endpoints/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `kubectl -n default patch deployment app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/"}}}]}}}}'`

## Final report

Service availability restored.

**Root cause**

- The `app` Deployment's readiness probe targeted `GET /readyz` on port 80, but the `nginx:1.31.5` container has no `/readyz` path (`/usr/share/nginx/html/readyz` does not exist). Kubelet logged `Readiness probe failed: HTTP probe failed with statuscode: 404`, so all Pods stayed `Ready: False`.
- Consequently the `app` Service's Endpoints/EndpointSlice contained only `notReadyAddresses` (10.244.0.2, .3, .7), so no backend was routable through ClusterIP 10.96.25.96.

**Change made (single, minimal)**

- Patched the owning Deployment `default/app` to point the readiness probe at `/` (a path nginx actually serves), keeping the probe itself, the image, replicas, ports, and the Service untouched:
  `kubectl -n default patch deployment app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/"}}}]}}}}'`
- Rollout converged: `deployment "app" successfully rolled out`.

**Verification (fresh observations)**

- Deployment: `READY 3/3, UP-TO-DATE 3, AVAILABLE 3` — matches the requested `replicas: 3`.
- Pods: `app-6c75b77dc8-77hl4`, `-86wmg`, `-8sdx2` all `1/1 Running`.
- Endpoints: `app  10.244.0.10:80,10.244.0.8:80,10.244.0.9:80` (EndpointSlice lists the same three ready addresses).
- Direct request through the Service ClusterIP from a pod: `GET http://10.96.25.96:80/` → `HTTP/1.1 200 OK`, `Server: nginx/1.31.5`, with the nginx index page body returned.
