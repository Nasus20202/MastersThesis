---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/readiness-probe-breaks-endpoints/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 1

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `edit_file {"new_string": " readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n periodSeconds: 10\n successThreshold: 1\n timeoutSeconds: 1", "old_string": " readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /readyz\n port: 80\n scheme: HTTP\n periodSeconds: 10\n successThreshold: 1\n timeoutSeconds: 1", "path": "default/app.yaml"}`
2. `kubectl patch deployment app -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/"}}}]}}}}' 2>`
3. `kubectl run -it --rm curl-test --image=curlimages/curl:latest --exec -- sh -c "sleep 3`
4. `kubectl run temp-curl --rm --image=curlimages/curl:latest -- sh -c "sleep 3`
5. `kubectl run temp-curl --image=curlimages/curl:latest -- sh -c "sleep 5`

## Final report

**Fix Complete.**

**Problem identified:** The readiness probe was checking `/readyz` which doesn't exist in nginx:1.31.5. This caused all pods to receive HTTP 404 responses on the probe, failing their readiness checks repeatedly (failureThreshold: 3), keeping them in `Ready=False` state. Since no pods were ready, the Service had no backends and was unavailable.

**Fix applied:** Changed the readiness probe path from `/readyz` to `/` using:

```bash
kubectl patch deployment app -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/"}}}]}}}}'
```

**Verification of outcome:**

| Metric                | Before Fix               | After Fix                   |
| --------------------- | ------------------------ | --------------------------- |
| Pod Ready Status      | 0/1 (all pods)           | 1/1 (all pods)              |
| Deployment Available  | 0/3                      | 3/3                         |
| Service HTTP Response | Connection refused / 404 | nginx welcome page returned |

The application is now accessible through its Kubernetes Service at `http://app:80/`.
