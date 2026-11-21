---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/readiness-probe-breaks-endpoints/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `kubectl -n default patch deployment app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/","port":80,"scheme":"HTTP"}}}]}}}}' 2>`

## Final report

## Root cause

The `app` Deployment's Pods were never marked Ready, so the `app` Service had **zero endpoints** and ClusterIP traffic was rejected/dropped.

- Service `default/app` selector `app=app` matched all 3 Pods and targetPort was correct — the selector/ports were fine.
- But `kubectl get endpoints app` showed an empty ENDPOINTS list, and the Deployment reported `unavailableReplicas: 3`, `MinimumReplicasUnavailable`.
- Pod events: `Readiness probe failed: HTTP probe failed with statuscode: 404`.
- Direct in-Pod check: `localhost/` → `200`, `localhost/readyz` → `404`. The container is a stock `nginx:1.31.5` (root `/usr/share/nginx/html` contains only `index.html`, `50x.html`); nothing in the cluster serves `/readyz` (only ConfigMap present was `kube-root-ca.crt`).

So the Pods (and therefore the Service) were healthy-ish, but the readiness probe pointed at a path the image does not implement.

## Change made (single, smallest persistent fix)

Patched only the readiness probe path on the owning resource, the Deployment (all other probe fields, image, replicas, ports, and Service left intact):

```
kubectl -n default patch deployment app --type=strategic \
  -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/","port":80,"scheme":"HTTP"}}}]}}}}'
```

Resulting probe: `{"failureThreshold":3,"httpGet":{"path":"/","port":80,"scheme":"HTTP"},"periodSeconds":10,"successThreshold":1,"timeoutSeconds":1}` — nginx serves `/` with 200.

## Verification (fresh observations)

- Rollout: `deployment "app" successfully rolled out`.
- Replica counts match: desired/updated/ready/available = **3/3/3/3**, Deployment condition `Available=True`, `Progressing=True`.
- Pods: all `1/1 Running` (`app-6c75b77dc8-6rsfd`, `-jwrrd`, `-rwv9t`), 0 restarts.
- Endpoints now populated: `app → 10.244.0.8:80, 10.244.0.9:80, 10.244.0.10:80`; EndpointSlice entries all `ready=true`.
- Request through the Service (not just object status): from inside a Pod, `http://app/` returned `200` on three consecutive calls and served the nginx index page (`<title>Welcome to nginx!</title>`); ClusterIP `http://10.96.227.137/` and `http://app.default.svc.cluster.local/` each returned `200`.
- No new `Unhealthy` probe events after 13:13:09Z; last event timestamp checked at 13:13:35Z.
