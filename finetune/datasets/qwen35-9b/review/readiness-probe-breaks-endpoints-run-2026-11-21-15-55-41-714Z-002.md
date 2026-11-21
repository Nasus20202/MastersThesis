---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/readiness-probe-breaks-endpoints/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `edit_file {"new_string": " readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n periodSeconds: 10\n successThreshold: 1\n timeoutSeconds: 1", "old_string": " readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /readyz\n port: 80\n scheme: HTTP\n periodSeconds: 10\n successThreshold: 1\n timeoutSeconds: 1", "path": "/tmp/deployment.yaml"}`
2. `kubectl apply -f /tmp/deployment.yaml -n default`
3. `kubectl run --rm -it curl-test --image=curlimages/curl:latest --namespace default -- wget -qO- http://app.default.svc.cluster.local/ 2>`

## Final report

## Analysis Complete

**Root Cause Identified:**
The application Pods were running but not Ready because the readinessProbe path `/readyz` was configured in the Deployment, but nginx doesn't serve this endpoint (returns 404). This prevented pods from ever becoming Ready, so no endpoints were registered for the Service.

**Fix Applied:**
Updated the Deployment's readinessProbe to check `/` instead of `/readyz`:

- Changed: `httpGet.path: /readyz` → `httpGet.path: /`

**Verification Confirmed:**

| Check                    | Before Fix         | After Fix                                           |
| ------------------------ | ------------------ | --------------------------------------------------- |
| Pod Ready Status         | 0/1 (3 pods)       | 1/1 (3 pods)                                        |
| Service Endpoints        | None               | 3 endpoints: `10.244.0.8,10.244.0.9,10.244.0.10:80` |
| Deployment Available     | False              | True                                                |
| HTTP Request via Service | N/A (no endpoints) | **Success** - Returns nginx welcome page            |

**Outcome:**

- All 3 replicas are now Ready, Up-to-Date, and Available
- Service responds successfully to requests
- Original constraint (service availability) restored
