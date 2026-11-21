---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/service-targetport-mismatch/prompt/008.json
verdict: accepted
reason: "accepted by the researcher for now (forceful write only); first review of a replacement attempt against the raw record by a Haiku model: force: Fix is a plain patch (msg 28, verified msg 75); test pods were force-deleted with --grace-period=0 --force (msg 40); plain delete suffices; report accurate."
---

# service-targetport-mismatch · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

The application Pods are healthy and the Service has ready endpoints, but the application is unavailable through the Service. Restore service availability.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Defining a Service

## Write commands

1. `kubectl patch svc app -n default -p '{"spec": {"ports": [{"port": 80, "targetPort": 80}]}}'`
2. `kubectl run -n default --rm --timeout=10s --image=busybox:latest test-curl -- wget -qO- http://app.default.svc.cluster.local 2>/dev/null`
3. `kubectl run -n default --rm --timeout=30s --image=curlimages/curl test-check -- curl -sI http://app.default.svc.cluster.local`
4. `kubectl run -n default --rm --image=curlimages/curl test-check -- curl -sI http://app.default.svc.cluster.local 2>`
5. `kubectl run -n default --image=curlimages/curl curl-test-curl -- curl -sI http://app.default.svc.cluster.local 2>`
6. `kubectl run curl-test-curl -n default --rm=true --image=curlimages/curl -- bash -c 'curl -sI http://app.default.svc.cluster.local`
7. `kubectl delete pod curl-test-curl -n default --grace-period=0 --force`
8. `kubectl run curl-test-verify -n default --rm=true --image=curlimages/curl -- bash -c 'echo "Request 1:"`
9. `kubectl run curl-test-verify -n default --image=curlimages/curl -- bash -c 'echo "Request 1:"`
10. `kubectl run curl-test-final -n default --image=curlimages/curl -- bash -c "curl -sI http://app.default.svc.cluster.local"`
11. `kubectl delete pod curl-test-final curl-test-verify -n default --grace-period=0 --force 2>`
12. `kubectl run curl-test-proper -n default --image=curlimages/curl -- bash -c "sleep 5`
13. `kubectl delete pod curl-test-proper -n default --grace-period=0 --force 2>`
14. `kubectl run verify-http -n default --image=httpd:alpine -- bash -c 'sleep 3`
15. `kubectl delete pod verify-http -n default --grace-period=0 --force 2>`
16. `kubectl run testcurl -n default --image=curlimages/curl -- curl -sI http://app.default.svc.cluster.local`
17. `kubectl delete pod testcurl -n default --grace-period=0 --force 2>`

## Final report

## Diagnosis Complete

**Root Cause:** Port mismatch between Service targetPort and container port.

- **Service `app` TargetPort:** 8080
- **Container/Deployment port:** 80 (nginx default)

Traffic through the Service was going to port 8080, but nginx was only listening on port 80.

**Fix Applied:** Updated Service `app` in namespace `default`:

```bash
kubectl patch svc app -n default -p '{"spec": {"ports": [{"port": 80, "targetPort": 80}]}}'
```

**Verification:**

- **Service status:** Port 80/TCP → TargetPort: 80 ✓
- **Endpoints:** 10.244.0.5:80, 10.244.0.6:80, 10.244.0.7:80 ✓
- **Service request test:** `curl http://app.default.svc.cluster.local` returned:
  ```
  HTTP/1.1 200 OK
  Server: nginx/1.31.5
  ```

**Outcome:** Service availability restored. All three app pods are healthy and successfully responding through the Service on port 80.
