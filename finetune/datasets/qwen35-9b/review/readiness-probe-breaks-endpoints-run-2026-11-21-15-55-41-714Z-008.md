---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/readiness-probe-breaks-endpoints/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `kubectl patch deployment app -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/"}}}]}}}}' 2>`
2. `kubectl run -q --rm -t --restart=Never curl-test --image=curlimages/curl:7.85.0 -- command -c "curl -s --connect-timeout 3 http://app.default.svc.cluster.local/`
3. `kubectl run -q --rm --restart=Never curl-test --image=curlimages/curl:7.85.0 -- wget -qO- http://app.default.svc.cluster.local/ 2>`
4. `kubectl run -q --rm --restart=Never --generator=run-pod/v1 test-curl --image=curlimages/curl:7.85.0 -- wget -qO- http://app.default.svc.cluster.local/ 2>`
5. `kubectl run -q --rm -t --restart=Never --image=curlimages/curl:7.85.0 curl-test -- wget -qO- http://app.default.svc.cluster.local/ 2>`
6. `kubectl run -q --rm --restart=Never --image=curlimages/curl:7.85.0 curl-test -- wget -qO- http://app.default.svc.cluster.local/ 2>`
7. `kubectl run -q --restart=Never --image=curlimages/curl:7.85.0 curl-test -- wget -qO- http://app.default.svc.cluster.local/ 2>`

## Final report

## Summary

**Root Cause:** The application pods had a readiness probe configured to hit `http://:80/readyz`, but nginx doesn't serve this endpoint, causing continuous 404 errors. This kept all pods from becoming Ready (0/1), making the Service unavailable.

**Fix Applied:** Changed the readiness probe path in the Deployment from `/readyz` to `/` (the default nginx endpoint that always responds).

**Verified Outcome:**

- All 3 pods are now Ready: `1/1 Running` with no failures
- Deployment status: **desired: 3, available: 3, ready: 3, unavailable: N/A**
- Service request test: Successfully connected to `http://app.default.svc.cluster.local/` (curl-test pod completed with exit code 0)
