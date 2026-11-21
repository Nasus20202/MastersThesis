---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/statefulset-headless-service/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-headless-service · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

A StatefulSet's Pods are running, but other workloads cannot reach them through their stable per-Pod DNS names. Restore stable network identity for the StatefulSet's Pods.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Pod Identity > Stable Network ID

## Write commands

1. `kubectl apply -f - <<'EOF'`
2. `kubectl run --rm -it dns-test --image=curlimages/curl:latest --restart=Never -n default -- nslookup app-0.default.svc.cluster.local 2>`
3. `kubectl run --rm dns-test --image=curlimages/curl:latest -n default -- nslookup app-0.app.default.svc.cluster.local 2>`
4. `kubectl run dns-test --image=curlimages/curl:latest -n default -- nslookup app-0.app.default.svc.cluster.local 2>`
5. `kubectl run dns-test --image=curlimages/curl:latest -n default -- nslookup app-1.app.default.svc.cluster.local 2>`
6. `kubectl delete pod dns-test -n default`
7. `kubectl run dns-test --image=curlimages/curl:latest -n default -- nslookup app-1.app.default.svc.cluster.local 2>`
8. `kubectl run --rm wget-test --image=curlimages/curl:latest -n default -- sh -c "curl -s http://app-0.app.default.svc.cluster.local`
9. `kubectl run wget-test --image=curlimages/curl:latest -n default -- curl -s http://app-0.app.default.svc.cluster.local/`
10. `kubectl run wget-test --image=curlimages/curl:latest -n default -- curl -s http://app-1.app.default.svc.cluster.local/`
11. `kubectl delete pod dns-test wget-test -n default`
12. `kubectl run --rm dns-verify --image=curlimages/curl:latest -n default -- sh -c 'echo "===DNS1==="`
13. `kubectl run dns-verify --image=curlimages/curl:latest -n default -- sleep 30`
14. `kubectl delete pod dns-verify -n default 2>/dev/null`
15. `kubectl run dns-verify --image=curlimages/curl:latest -n default -- sh -c "nslookup app-0.app.default.svc.cluster.local`
16. `kubectl run http-verify --image=curlimages/curl:latest -n default -- sh -c "curl -s http://app-0.app.default.svc.cluster.local/`
17. `kubectl delete pod dns-verify http-verify -n default 2>/dev/null`

## Final report

## Summary

**Root Cause**: The StatefulSet `app` in namespace `default` was missing its required headless Service. Without this Service, Kubernetes cannot provide stable per-Pod DNS names (e.g., `app-0.app.default.svc.cluster.local`), causing DNS resolution failures (NXDOMAIN) for other workloads trying to reach the pods.

**Fix Applied**: Created a headless Service named `app` in namespace `default`:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: app
  namespace: default
spec:
  clusterIP: None
  ports:
    - port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: app
```

**Verified Outcomes**:

| Check                        | Result      | Evidence                                             |
| ---------------------------- | ----------- | ---------------------------------------------------- |
| DNS Resolution (app-0)       | ✓ Resolved  | `app-0.app.default.svc.cluster.local` → `10.244.0.4` |
| DNS Resolution (app-1)       | ✓ Resolved  | `app-1.app.default.svc.cluster.local` → `10.244.0.6` |
| HTTP Request via DNS (app-0) | ✓ Success   | Returns nginx welcome page                           |
| HTTP Request via DNS (app-1) | ✓ Success   | Returns nginx welcome page                           |
| Service Endpoints Populated  | ✓ Confirmed | `10.244.0.4:80,10.244.0.6:80`                        |

The StatefulSet's Pods now have stable network identity and are accessible via their per-Pod DNS names.
