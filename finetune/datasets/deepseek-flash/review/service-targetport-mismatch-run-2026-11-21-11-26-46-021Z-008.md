---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/service-targetport-mismatch/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# service-targetport-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

The application Pods are healthy and the Service has ready endpoints, but the application is unavailable through the Service. Restore service availability.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Defining a Service

## Write commands

1. `kubectl run t --rm -i --restart=Never --image=busybox:1.36 -- wget -qO- --timeout=5 http://app.default.svc.cluster.local/ 2>`
2. `kubectl patch svc app -n default --type=merge -p '{"spec":{"ports":[{"name":"http","port":80,"protocol":"TCP","targetPort":80}]}}' 2>`
3. `kubectl run t$i --rm -i --restart=Never --image=busybox:1.36 --command -- wget -qO- --timeout=5 http://app.default.svc.cluster.local/ 2>/dev/null`
4. `kubectl run tt --rm -i --restart=Never --image=busybox:1.36 --command -- sh -c 'for i in $(seq 1 10)`
5. `kubectl run v --rm -i --restart=Never --image=busybox:1.36 --command -- sh -c 'ok=0`

## Final report

## Root cause

The Service `default/app` forwarded to the wrong container port. It was configured with `targetPort: 8080`, but the nginx containers only listen on **80** (Deployment declares `containerPort: 80`, named `http`). Endpoints were therefore populated with the wrong port, so every packet was sent to a closed port.

Evidence gathered before the fix:

- Through the Service: `wget: can't connect to remote host (10.96.151.182): Connection refused`
- Directly in a Pod: `http://localhost:80/` → **200**, `http://localhost:8080/` → **connection refused (exit 7)**
- Endpoints/EndpointSlice listed `10.244.0.2:8080, 10.244.0.4:8080, 10.244.0.5:8080` — a port with no listener

This matched the reported symptom exactly: Pods `Running 1/1`, Endpoints present and "ready", yet no traffic could succeed.

## Change made

Smallest justified change — patched only `targetPort` on the Service, leaving the client-facing port, selector, name, and type untouched:

```
targetPort: 8080  →  targetPort: 80
```

`kubectl patch svc app -n default --type=merge -p '{"spec":{"ports":[{"name":"http","port":80,"protocol":"TCP","targetPort":80}]}}'` → `service/app patched`

## Verification (fresh observations)

- Service spec: `[{"name":"http","port":80,"protocol":"TCP","targetPort":80}]` — clients still connect on port **80**
- Endpoints now `10.244.0.2:80,10.244.0.4:80,10.244.0.5:80`
- **30 consecutive requests through the Service** (`http://app.default.svc.cluster.local/`) from a throwaway Pod: `HTTP200=30 FAILED=0`, each returning the nginx welcome page (an earlier 10-request run showed `8 Welcome to nginx` only because the `grep` captured no count for 2 responses; the status-code run confirms 30/30)
- Deployment reconciliation intact: `replicas=3 updated=3 ready=3 available=3`

Constraints preserved: the 3 application Pods and the Deployment were never modified or restarted, the Service stayed `ClusterIP` on port 80 with selector `app: app`, and no other Service in the cluster was touched.
