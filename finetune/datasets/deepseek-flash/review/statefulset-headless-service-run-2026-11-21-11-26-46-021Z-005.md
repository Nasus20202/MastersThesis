---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-headless-service/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-headless-service · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A StatefulSet's Pods are running, but other workloads cannot reach them through their stable per-Pod DNS names. Restore stable network identity for the StatefulSet's Pods.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Pod Identity > Stable Network ID

## Write commands

1. `kubectl apply -f /tmp/app-headless-svc.yaml`
2. `kubectl run netcheck --image=nginx:1.31.5 --restart=Never -n default --command -- sh -c '`
3. `kubectl delete pod netcheck -n default --wait=true 2>`

## Final report

**Root cause:** StatefulSet `default/app` has `spec.serviceName: app`, but no Service named `app` existed in the namespace. Without the governing headless Service there was no EndpointSlice and no per-Pod A records, so `app-0.app.default.svc.cluster.local` / `app-1...` could not resolve or be reached. The Pods themselves were healthy (`2/2` ready).

**Fix:** Created the missing governing headless Service (no changes to the StatefulSet or Pods):

```yaml
apiVersion: v1
kind: Service
metadata: { name: app, namespace: default }
spec:
  clusterIP: None
  selector: { app: app }
  ports: [{ name: http, port: 80, targetPort: http, protocol: TCP }]
```

**Verification (fresh observations):**

- Service state: `svc/app` is `ClusterIP: None`, port `80/TCP`.
- EndpointSlice for the Service: `app-7rwcd` has endpoints `10.244.0.5` and `10.244.0.6`, both with `ready=true`.
- Pod identity fields: `app-0 hostname=app-0 subdomain=app`, `app-1 hostname=app-1 subdomain=app`.
- From a separate workload pod in the namespace (an nginx Pod, deleted after the test):
  - `getent hosts app-0.app.default.svc.cluster.local` → `10.244.0.5`
  - `getent hosts app-1.app.default.svc.cluster.local` → `10.244.0.6`
  - `curl http://app-0.app.default.svc.cluster.local/` → HTTP `200`
  - `curl http://app-1.app.default.svc.cluster.local/` → HTTP `200`
  - `curl http://app.default.svc.cluster.local/` (headless Service name) → HTTP `200`
- Controller after the change: StatefulSet `app` requested replicas `2`, updated `2`, ready `2`, available `2`.
