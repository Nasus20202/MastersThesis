---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/service-targetport-mismatch/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# service-targetport-mismatch · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 4

## Task

The application Pods are healthy and the Service has ready endpoints, but the application is unavailable through the Service. Restore service availability.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Defining a Service

## Write commands

1. `kubectl run -n default curl-test --rm --restart=Never --image=curlimages/curl:latest -- wget -qO- http://app/`
2. `kubectl run -n default curl-test --rm=true --restart=Never --image=curlimages/curl:latest -- wget -qO- http://app/ 2>`
3. `kubectl run -n default curl-test --restart=Never --image=curlimages/curl:latest -- wget -qO- http://app/ 2>`
4. `kubectl patch svc app -n default -p '{"spec":{"ports":[{"name":"http","port":80,"protocol":"TCP","targetPort":80}]}}'`
5. `kubectl run -n default curl-test --restart=Never --image=curlimages/curl:latest -- wget -qO- http://app/ 2>`
6. `kubectl delete pod curl-test -n default`
7. `kubectl run -n default curl-test --restart=Never --image=curlimages/curl:latest -- wget -qO- http://app/ 2>`
8. `kubectl delete pod curl-test -n default 2>`

## Final report

## Issue Resolution Summary

**Root Cause Identified:**
The Service's `targetPort` was configured as **8080**, but the application (nginx) containers were actually listening on port **80**. This port mismatch caused traffic routed through the Service to reach a non-listening port on the pods.

**Evidence Found:**

- Service configuration: `targetPort: 8080`
- Container specification: `Port: 80/TCP (http)`
- All 3 app pods were healthy and running with endpoints ready, but traffic failed to reach the application

**Fix Applied:**
Updated the Service `app` in namespace `default` to change `targetPort` from **8080** to **80**.

**Verification:**
Successfully tested HTTP connectivity through the Service by executing a curl request that returned the nginx welcome page, confirming service availability is now restored.

---

| Check                             | Before Fix | After Fix                                   |
| --------------------------------- | ---------- | ------------------------------------------- |
| Service targetPort                | 8080       | 80                                          |
| Container listening port          | 80         | 80                                          |
| HTTP connectivity through Service | Failed     | **Succeeded** (nginx welcome page received) |
