---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/headless-service-dns/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl apply -f /tmp/app-headless-svc.yaml 2>`

## Final report

Done. The missing headless Service was the only fault; the StatefulSet itself was healthy (2/2 ready) and already pointed at `serviceName: app`.

**Change made** — created Service `app` in `default` (single object, no other resources touched):

- `clusterIP: None` (headless)
- `selector: app=app` (matches the StatefulSet's pod template/selector)
- `port: 80` → `targetPort: 80` (matches the container's `http` port 80)

**Verified outcome (fresh observations):**

- Service spec: `clusterIP=None`, `selector={"app":"app"}`, `port=80/80`.
- Endpoints populated with both replicas: `10.244.0.2:80` (app-0) and `10.244.0.6:80` (app-1); matching EndpointSlice `app-drjvh` lists both IPs on port 80.
- Stable per-Pod DNS now resolves from inside a workload pod:
  - `app-0.app.default.svc.cluster.local -> 10.244.0.2`
  - `app-1.app.default.svc.cluster.local -> 10.244.0.6`
  - `app.default.svc.cluster.local -> 10.244.0.2, 10.244.0.6`
- A real request through the Service succeeded: `HTTP 200` returning the nginx welcome page.
- StatefulSet health unchanged and matched: `replicas=2, ready=2, available=2, updated=2`; pods `app-0` and `app-1` both `1/1 Running`.

All requested constraints (name `app`, `default` namespace, headless `clusterIP: None`, selector `app=app`, port 80) are satisfied and the originally missing per-Pod DNS names are now functional.
