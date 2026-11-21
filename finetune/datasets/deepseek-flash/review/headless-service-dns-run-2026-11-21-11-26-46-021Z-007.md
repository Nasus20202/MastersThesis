---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/headless-service-dns/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl apply -f - <<'EOF'`

## Final report

## Observed state and fix

**Before:** `kubectl get svc -n default` listed only `kubernetes`; `service/app not found`. The StatefulSet `app` (replicas 2, selector `app=app`, `serviceName: app`, container port 80/http) had no governing headless Service, so the Pods `app-0`/`app-1` (both `Running`, `1/1`, labels `app=app`) had no stable per-Pod DNS names.

**Change (single, minimal):** created Service `app` in `default` with `clusterIP: None`, selector `app=app`, port 80 → targetPort 80/TCP. Nothing else was modified; the StatefulSet, its Pods and unrelated objects were left untouched.

**Verification (fresh observations after the change):**

- `service/app`: `CLUSTER-IP None`, `SELECTOR app=app`, `PORT(S) 80/TCP` — headless as requested.
- EndpointSlice `app-jjxkr` (label `kubernetes.io/service-name=app`): endpoints `10.244.0.5,10.244.0.6`, ports `80` — exactly the two Pod IPs of `app-0`/`app-1`, no more, no fewer.
- Per-Pod DNS resolve from inside `app-0` (`getent hosts`):
  - `app-0.app.default.svc.cluster.local` → `10.244.0.5`
  - `app-1.app.default.svc.cluster.local` → `10.244.0.6`
  - `app.default.svc.cluster.local` → both `10.244.0.5` and `10.244.0.6`
- Traffic through the Service DNS name (from `app-1`): `http://app.default.svc.cluster.local/` → **HTTP 200** from `10.244.0.6`.
- Per-Pod DNS traffic: `app-1` → `http://app-0.app.default.svc.cluster.local/` → **HTTP 200** via `10.244.0.5`; `app-0` → `http://app-1.app.default.svc.cluster.local/` → **HTTP 200** via `10.244.0.6`.
- Controller health unchanged and matching: `statefulset/app` `READY 2/2`, both Pods `1/1 Running`, 0 restarts.

The originally reported symptom (missing stable per-Pod DNS for the StatefulSet) is gone, and the Service matches the requested name, namespace, `clusterIP None`, selector `app=app`, and port 80 without altering any other setting.
