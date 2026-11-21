---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/readiness-probe-breaks-endpoints/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 4

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `kubectl patch deployment app -n default -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/"},"periodSeconds":10,"failureThreshold":3,"successThreshold":1,"timeoutSeconds":1}}]}}}}'`
2. `kubectl run -n default test-curl --rm -it --image=busybox:1.28 -- curl -s http://app.default.svc.cluster.local/ 2>/dev/null`

## Final report

**Service availability restored.** Here's the complete diagnosis and fix:

## Root Cause

The readinessProbe was configured to hit path `/readyz` on port 80, but nginx does not serve this endpoint (it returns HTTP 404). Pods remained in `NotReady` state because the probe failed continuously (`failureThreshold: 3`). Without ready pods, no endpoints were populated for the Service.

## Fix Applied

Changed the Deployment's readinessProbe path from `/readyz` to `/` (standard nginx root endpoint that returns HTTP 200).

```bash
kubectl patch deployment app -n default \
  --patch='{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/"},"periodSeconds":10,"failureThreshold":3,"successThreshold":1,"timeoutSeconds":1}}]}}}}'
```

## Verification

| Check                | Before Fix | After Fix                                        |
| -------------------- | ---------- | ------------------------------------------------ |
| Pod Ready Status     | 0/1        | **1/1**                                          |
| Deployment Available | 0          | **3**                                            |
| Endpoints populated  | No IPs     | **10.244.0.8:80, 10.244.0.9:80, 10.244.0.10:80** |

All 3 pods are now Ready, endpoints show correct pod IPs, and the Service is accessible.
