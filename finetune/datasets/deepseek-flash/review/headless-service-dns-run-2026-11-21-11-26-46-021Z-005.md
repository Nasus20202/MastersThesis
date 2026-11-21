---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/headless-service-dns/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl apply -f -`
2. `kubectl run dnstest --image=busybox:1.36 --restart=Never --namespace default --command -- sleep 300 2>`
3. `kubectl delete pod dnstest -n default --wait=true 2>`

## Final report

The missing headless Service has been created and verified end-to-end.

**What was wrong:** StatefulSet `app` (default ns) declares `spec.serviceName: app`, and both Pods already carried `subdomain: app`, but no Service named `app` existed — only `kubernetes`. So no EndpointSlice existed and no per-Pod DNS names could be published.

**Change made** (single persistent object, nothing else touched — StatefulSet, Pods and their labels left as-is):

```yaml
apiVersion: v1
kind: Service
metadata: { name: app, namespace: default }
spec:
  clusterIP: None # headless
  selector: { app: app }
  ports: [{ name: http, port: 80, protocol: TCP, targetPort: 80 }]
```

**Verified (fresh observations):**

- Service: `CLUSTERIP=None`, `PORT=80`, `TARGETPORT=80`, `SELECTOR=map[app:app]`.
- EndpointSlice `app-4vr76` now exists with both Pod IPs `10.244.0.4`, `10.244.0.6`, `ready=true true` for port 80.
- DNS from an in-cluster Pod (`busybox` test Pod, since `nginx` has no DNS client): `app.default.svc.cluster.local` → `10.244.0.4` and `10.244.0.6` (A records for the headless set); per-Pod names resolve — `app-0.app...` → `10.244.0.4`, `app-1.app...` → `10.244.0.6`.
- SRV records published: `_http._tcp.app.default.svc.cluster.local → 0 50 80 app-0.app...` and `... app-1.app...`, confirming the named port is exposed per Pod.
- Traffic through the Service succeeds: 4/4 `wget http://app.default.svc.cluster.local/` returned the nginx page; direct per-Pod `app-0`/`app-1` requests also returned the nginx page.
- Original workload state intact and matching: StatefulSet requested=`2`, ready=`2`, updated=`2`, available=`2`, current=`2`; both Pods `READY=true`, `SUBDOMAIN=app`.

The temporary `dnstest` Pod used only for DNS/HTTP probing was deleted; no other resources were modified.
