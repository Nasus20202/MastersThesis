---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-headless-service/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-headless-service · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A StatefulSet's Pods are running, but other workloads cannot reach them through their stable per-Pod DNS names. Restore stable network identity for the StatefulSet's Pods.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Pod Identity > Stable Network ID

## Write commands

1. `kubectl apply -f /tmp/app-headless.yaml`
2. `kubectl run dnscheck --image=busybox:1.36 --restart=Never --command -- sh -c 'for n in app-0.app.default.svc.cluster.local app-1.app.default.svc.cluster.local`
3. `kubectl delete pod dnscheck -n default --wait=true`

## Final report

Stable per-Pod network identity is restored.

## Root cause

The StatefulSet `default/app` declares `spec.serviceName: app`, and its Pods carry `hostname: app-0`/`app-1` with `subdomain: app`. That governing headless Service `app` was missing entirely — the namespace contained no Service other than `kubernetes`. Without it, the Pods' `hostname`+`subdomain` pair has no DNS zone to live in, so `app-N.app.default.svc.cluster.local` had no records even though the Pods were healthy (`2/2` Running).

## Change made

Created the missing persistent object — the headless Service the StatefulSet already referenced (no changes to the StatefulSet, its Pods, or any unrelated setting):

```yaml
kind: Service
metadata: { name: app, namespace: default }
spec:
  clusterIP: None
  selector: { app: app } # matches StatefulSet selector/pod labels
  ports: [{ name: http, port: 80, targetPort: 80 }]
```

## Verification (fresh observations)

- Service exists as headless: `service/app  ClusterIP  None  80/TCP  app=app`.
- EndpointSlice `app-j58ts` (owned by Service `app`) lists both Pods as ready, each with its own stable hostname: `hostname: app-0 → 10.244.0.5`, `hostname: app-1 → 10.244.0.6`.
- Per-Pod DNS resolves to the matching distinct Pod IP from a **separate workload** (temporary `busybox` Pod, since removed):
  - `app-0.app.default.svc.cluster.local` → `10.244.0.5` (= app-0's IP)
  - `app-1.app.default.svc.cluster.local` → `10.244.0.6` (= app-1's IP)
- Requests through those names succeed, not just DNS: `wget http://app-0.app.default.svc.cluster.local/` and `.../app-1...` each returned the nginx index page (`<!DOCTYPE html>`).
- Cross-checked from inside the Pod too: `getent hosts` resolved both FQDNs and short forms `app-0.app` / `app-1.app` to the correct per-Pod IPs.

## Constraints preserved

- StatefulSet ownership of identity untouched: `replicas=2 requested, readyReplicas=2, availableReplicas=2`, `generation 1 == observedGeneration 1`, no Pod restarts, same Pod names and IPs as before the fix.
- No Pods or other resources deleted/recreated as a "fix"; the only deleted object was my own temporary verification Pod, and the StatefulSet, PVC retention policy, and image were left as-is.
