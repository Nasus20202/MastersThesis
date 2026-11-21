---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/service-targetport-mismatch/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# service-targetport-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

The application Pods are healthy and the Service has ready endpoints, but the application is unavailable through the Service. Restore service availability.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Defining a Service

## Write commands

1. `kubectl run curltest --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- curl -sS -m 5 -o /dev/null -w '%{http_code}\n' http://app.default.svc.cluster.local/ 2>`
2. `kubectl patch svc app -n default --type=merge -p '{"spec":{"ports":[{"name":"http","port":80,"protocol":"TCP","targetPort":80}]}}'`
3. `kubectl run curltest-$i --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- curl -sS -m 8 -o /dev/null -w "svc:%{http_code}\n" http://app.default.svc.cluster.local/ 2>`
4. `kubectl run directtest --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'for ip in 10.244.0.2 10.244.0.4 10.244.0.7`
5. `kubectl run finalcheck --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'curl -sS -m 8 -w "\nHTTP:%{http_code}\n" http://app/`

## Final report

## Root cause

The `app` Service pointed traffic at a port nothing was listening on:

- Service `default/app`: `port: 80` → **`targetPort: 8080`**
- Pods (`app-575b7496bc-*`) run `nginx`, whose container port is **80** (declared as `containerPort: 80, name: http`), with no process on 8080.

The selector matched and the EndpointSlice was populated (`10.244.0.2/4/7`), which is why endpoints looked "ready", but kube-proxy forwarded to `pod:8080`, where connections were refused:

```
$ curl http://app.default.svc.cluster.local/
curl: (7) Failed to connect to app.default.svc.cluster.local port 80 after 3 ms
```

## Fix (single field, Service only)

```bash
kubectl patch svc app -n default --type=merge \
  -p '{"spec":{"ports":[{"name":"http","port":80,"protocol":"TCP","targetPort":80}]}}'
```

Nothing else changed: Deployment `generation=1/observedGeneration=1`, all 3 Pods still `RESTARTS=0`, image `nginx:1.31.5`, selector `app=app`, Service type/ClusterIP (`10.96.224.51`) and `port: 80` preserved.

## Verified outcome

- Endpoints now resolve to the real listener: `app → 10.244.0.2:80, 10.244.0.4:80, 10.244.0.7:80`.
- Requests **through the Service** succeed repeatedly (fresh Pods, 12s total):
  ```
  svc:200
  svc:200
  svc:200
  ```
- Response body served through the Service name `http://app/` is the nginx welcome page (`<title>Welcome to nginx!</title>`), HTTP 200.
- Workload health unchanged and matching: `replicas=3 ready=3 updated=3 available=3`, zero restarts.
