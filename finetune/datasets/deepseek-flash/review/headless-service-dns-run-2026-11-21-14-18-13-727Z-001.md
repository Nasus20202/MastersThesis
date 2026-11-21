---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-14-18-13-727Z/headless-service-dns/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · deepseek-flash · run-2026-11-21-14-18-13-727Z · attempt 1

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl apply -f - 2>`

## Final report

## Diagnosis

The StatefulSet `app` was healthy (`replicas=2 ready=2 available=2`) and pointed at `spec.serviceName: app`, but `kubectl get svc` listed only the `kubernetes` Service — the governing headless Service did not exist, so no stable per-Pod DNS records could be created.

## Change made

Created one new resource (nothing else modified, no Pods restarted):

```
apiVersion: v1
kind: Service
metadata:
  name: app            # namespace: default
  labels: {app: app}
spec:
  clusterIP: None      # headless
  selector: {app: app}
  ports: [{name: http, port: 80, protocol: TCP, targetPort: 80}]
```

## Verification (fresh observations after the change)

- Service object: `clusterIP=None selector={"app":"app"} ports=80 type=ClusterIP` ✔ matches requested spec.
- EndpointsSlice `app-jhv7b`: endpoints `10.244.0.2,10.244.0.6` — exactly `app-0` and `app-1`'s Pod IPs, port 80.
- Per-Pod DNS (queried from inside `app-0`):
  - `app-0.app.default.svc.cluster.local` → `10.244.0.2` ✔
  - `app-1.app.default.svc.cluster.local` → `10.244.0.6` ✔
  - `app.default.svc.cluster.local` → both `10.244.0.6`, `10.244.0.2` ✔ (headless returns all Pod A records)
- Traffic through the Service (from inside `app-1`): `http://app.default.svc.cluster.local/` → **200**, `app-0.app...` → **200**, `app-1.app...` → **200** ✔
- Original workload untouched and still consistent: `replicas=2 ready=2 available=2 current=2 updated=2 serviceName=app` ✔

The reported symptom is resolved: the headless Service exists with the exact requested selector, `clusterIP: None` and port 80, and both stable per-Pod DNS names now resolve to the correct Pods and serve HTTP 200.
