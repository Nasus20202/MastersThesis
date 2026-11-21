---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-headless-service/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-headless-service · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 1

## Task

A StatefulSet's Pods are running, but other workloads cannot reach them through their stable per-Pod DNS names. Restore stable network identity for the StatefulSet's Pods.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Pod Identity > Stable Network ID

## Write commands

1. `kubectl run dnstest --image=busybox:1.36 --restart=Never --command -- sleep 300 2>`
2. `kubectl apply -f /tmp/app-headless-svc.yaml 2>`
3. `kubectl delete pod dnstest -n default --wait=true --timeout=60s 2>`

## Final report

## Outcome

Stable network identity is restored for StatefulSet `app`'s Pods.

**Root cause:** StatefulSet `default/app` declares `spec.serviceName: app`, and its Pods are created with `hostname: app-0`/`app-1` and `subdomain: app`. That subdomain only produces per-Pod DNS records if a **headless Service named `app`** exists to govern it — and no such Service existed in `default` (only `kubernetes` in `default`, `kube-dns` in `kube-system`). CoreDNS therefore had no records to serve.

**Evidence of the fault (before):**

- `kubectl get svc -A` → no `app` Service; `kubectl get all -n default` showed only `kubernetes`
- From a separate workload: `nslookup app-0.app.default.svc.cluster.local` → **NXDOMAIN** (`rc=1`)

**Change made (single, minimal):** created the missing governing headless Service `app` in `default` — `clusterIP: None`, `selector: app=app` (matching the StatefulSet selector/labels), port `http:80 → targetPort http`, `publishNotReadyAddresses: true`. The StatefulSet was left untouched (its `serviceName`, replicas, and template were already correct); no Pods were deleted or restarted.

**Verification (after, fresh observations):**

- Endpoints populated with exactly the two Pod IPs: `app 10.244.0.3:80,10.244.0.6:80`; EndpointSlice `app-np562` → `10.244.0.3,10.244.0.6`
- Per-Pod DNS from a separate Pod now resolves: `app-0.app.default.svc.cluster.local` → `10.244.0.3`, `app-1.app.default.svc.cluster.local` → `10.244.0.6` (both `rc=0`, previously NXDOMAIN)
- Actual traffic through those names succeeds: HTTP GET on both returned the nginx welcome page (`wget_rc=0`)
- StatefulSet healthy and undisturbed: requested/updated/ready/available = **2/2/2/2**, `serviceName=app`; `app-0` and `app-1` both `1/1 Running` with **0 restarts** and original ages (3m54s / 3m31s), proving no rollout was triggered
- Temporary diagnostic Pod removed; only `app-0` and `app-1` remain in `default`
