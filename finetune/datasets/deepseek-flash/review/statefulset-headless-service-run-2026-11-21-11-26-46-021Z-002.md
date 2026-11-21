---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/statefulset-headless-service/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# statefulset-headless-service · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A StatefulSet's Pods are running, but other workloads cannot reach them through their stable per-Pod DNS names. Restore stable network identity for the StatefulSet's Pods.

Documented in:

- content/en/docs/concepts/workloads/controllers/statefulset.md: Pod Identity > Stable Network ID

## Write commands

1. `kubectl run dnstest --rm -i --restart=Never --image=busybox:1.36 --command -- sh -c 'nslookup app-0.app.default.svc.cluster.local 2>`
2. `kubectl apply -f -`
3. `kubectl run dnstest --rm -i --restart=Never --image=busybox:1.36 --command -- sh -c '`
4. `kubectl run dnstest2 --rm -i --restart=Never --image=nicolaka/netshoot:latest --command -- sh -c '`

## Final report

## Diagnosis

The StatefulSet `app` prescribes stable identity via `spec.serviceName: app`, and its Pods were created with `hostname: app-0` / `subdomain: app` — trusting a governing **headless Service named `app`** to publish the DNS records. That Service did not exist:

- `kubectl get svc -A` → only `kubernetes` and `kube-dns`; no `default/app`
- In-pod lookup from a test Pod: `** server can't find app-0.app.default.svc.cluster.local: NXDOMAIN`

No Pod restarts or spec changes were needed; the Pods were healthy (`2/2`) — only the DNS-owning object was missing.

## Change (minimal, single object)

Created the missing governing headless Service in `default`:

```yaml
apiVersion: v1
kind: Service
metadata: { name: app, namespace: default, labels: { app: app } }
spec:
  clusterIP: None # headless
  selector: { app: app } # matches StatefulSet pod labels
  ports: [{ name: http, port: 80, targetPort: 80, protocol: TCP }]
```

`clusterIP: None` and the selector matching `app: app` are what make this the StatefulSet's identity service; nothing else (StatefulSet spec, Pods, image, replicas) was touched.

## Verification (fresh observations, from a Pod inside the cluster)

DNS now resolves **per-Pod, to distinct IPs** (not load-balanced), and each name serves traffic through the Service path:

| query                                 | resolved address | HTTP via `http://<name>/`  |
| ------------------------------------- | ---------------- | -------------------------- |
| `app-0.app.default.svc.cluster.local` | 10.244.0.5       | 200 (remote_ip 10.244.0.5) |
| `app-1.app.default.svc.cluster.local` | 10.244.0.6       | 200 (remote_ip 10.244.0.6) |
| `app-0.app` (search-path short form)  | 10.244.0.5       | 200                        |
| `app-1.app` (search-path short form)  | 10.244.0.6       | 200                        |

Supporting state:

- Endpoints: `app-9j7sb: 10.244.0.5/true 10.244.0.6/true` — both Pods Ready and attached.
- Service: `clusterIP=None`, `type=ClusterIP` (unchanged constraints; no NodePort/LoadBalancer added).
- StatefulSet: requested `replicas: 2` vs `readyReplicas: 2`, `availableReplicas: 2`, `updatedReplicas: 2` — all match; both Pods `READY=true`, `RESTARTS=0`.

The originally reported symptom (NXDOMAIN on per-Pod names) is gone, and each stable name maps to its own Pod rather than a shared VIP.
