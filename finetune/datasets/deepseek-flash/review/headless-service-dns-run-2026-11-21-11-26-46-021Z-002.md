---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/headless-service-dns/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl apply -f -`

## Final report

The headless Service was missing; I created it and verified stable per-Pod DNS end to end.

**Applied change** (only new object added, nothing else modified):

```yaml
apiVersion: v1
kind: Service
metadata:
  name: app
  namespace: default
spec:
  clusterIP: None # headless
  selector: { app: app }
  ports:
    - port: 80
      targetPort: 80
      protocol: TCP
      name: http
```

**Verification (fresh observations):**

- `kubectl get svc app -n default`: `ClusterIP None`, `PORT(S) 80/TCP`, `SELECTOR app=app`.
- Endpoints populated with both Pods: `10.244.0.4:80,10.244.0.6:80`.
- DNS from inside `app-0`:
  - `app-0.app.default.svc.cluster.local` → `10.244.0.4`
  - `app-1.app.default.svc.cluster.local` → `10.244.0.6`
  - `app.default.svc.cluster.local` → both Pod IPs (expected for a headless Service).
- Requests through the Service DNS names succeed: `http://app-1.app...` → HTTP 200, `http://app-0.app...` → HTTP 200.
- StatefulSet unchanged and healthy: `statefulset.apps/app 2/2` ready, `readyReplicas: 2`, `availableReplicas: 2`, both Pods `Running 1/1`.
