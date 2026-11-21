---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/statefulset-headless-service/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-headless-service · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 5

## Task

A StatefulSet's Pods are running, but other workloads cannot reach them through their stable per-Pod DNS names. Restore stable network identity for the StatefulSet's Pods.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Pod Identity > Stable Network ID

## Write commands

1. `kubectl create svc ClusterIP app -n default --selector app=app --dry-run=client -o yaml`
2. `kubectl run app --dry-run=client -o yaml`
3. `kubectl apply -f - <<'EOF'`
4. `kubectl run -n default -i --rm dns-test --image=curlimages/curl:latest --restart=Never -- cmd nslookup app-0.app.default.svc.cluster.local 2>/dev/null`
5. `kubectl run -n default -i --rm dns-test2 --image=curlimages/curl:latest --restart=Never -- cmd nslookup app-1.app.default.svc.cluster.local 2>/dev/null`
6. `kubectl run -n default -i --rm dns-test3 --image=curlimages/curl:latest --restart=Never -- cmd curl -s --connect-timeout 2 http://app-0.app.default.svc.cluster.local:80`
7. `kubectl run -n default -i --rm dns-test4 --image=curlimages/curl:latest --restart=Never -- cmd curl -s --connect-timeout 2 http://app-1.app.default.svc.cluster.local:80`
8. `kubectl run -n default -i --rm dns-test5 --image=curlimages/curl:latest --restart=Never -- cmd curl -s --connect-timeout 2 http://app.default.svc.cluster.local`

## Final report

## Summary

**Problem:** The StatefulSet `app` in namespace `default` had 2 running pods, but other workloads couldn't reach them through stable per-Pod DNS names because the required headless Service was missing.

**Root Cause:** A StatefulSet requires a headless Service (with `clusterIP: None`) to create stable per-Pod DNS records like `app-0.app.default.svc.cluster.local` and `app-1.app.default.svc.cluster.local`. The StatefulSet referenced `serviceName: app`, but no Service with that name existed.

**Fix Applied:** Created a headless Service matching the StatefulSet's selector:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: app
  namespace: default
spec:
  clusterIP: None # Headless service
  selector:
    app: app
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
      name: http
```

**Verification:**

1. **Service created successfully:** `service/app created` with type ClusterIP, clusterIP None (headless)
2. **Endpoints registered:** Both pods `app-0` (10.244.0.2) and `app-1` (10.244.0.6) are registered as endpoints
3. **DNS resolution verified:**
   - `curl http://app-0.app.default.svc.cluster.local:80` → nginx welcome page ✓
   - `curl http://app-1.app.default.svc.cluster.local:80` → nginx welcome page ✓
4. **Pods healthy:** Both pods running with 1/1 ready

The StatefulSet's Pods now have stable network identity and are accessible through their DNS names `app-0.app.default.svc.cluster.local` and `app-1.app.default.svc.cluster.local`.
