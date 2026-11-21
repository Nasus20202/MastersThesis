---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/readiness-probe-breaks-endpoints/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `kubectl -n default patch deployment app --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/"}]' 2>`

## Final report

Service availability is restored.

## Diagnosis (evidence)

- `kubectl get pods` showed all 3 `app` pods **Running but `0/1` ready**; deployment was `READY 0/3`, `unavailableReplicas: 3`, with condition `Available=False (MinimumReplicasUnavailable)`.
- EndpointSlice `app-d9zms` listed all 3 pod IPs with **`ready: false, serving: false`**, so the `v1 Endpoints` list for `service/app` was empty → no backends to forward traffic to.
- Pod events repeated: `Readiness probe failed: HTTP probe failed with statuscode: 404` on `http://10.244.0.4:80/readyz`.
- In-container verification showed the container is stock nginx: `/readyz` → **404**, `/` → **200** (`location /` is the only served path in `/etc/nginx/conf.d/default.conf`).
- Ruled out other causes: image pulled and running (`nginx@sha256:05b8cb60...`, `nginx/1.31.5`), Service `selector={app: app}` matched the pods, `targetPort 80` matched `containerPort 80`, and there was **no ConfigMap or Secret** that was meant to supply a `/readyz` handler — so the faulty state was the probe path itself in the Deployment, not the app, Service, or networking.

## Change (smallest justified, on the persistent owner)

Patched only `spec.template.spec.containers[0].readinessProbe.httpGet.path` on `deployment/default/app`: `/readyz` → `/`. All other probe settings (`port: 80`, `scheme: HTTP`, `periodSeconds: 10`, `timeoutSeconds: 1`, `failureThreshold: 3`), the Deployment replicas/image, and the Service (ClusterIP `10.96.244.125`, `80->80`, `selector app=app`) were left untouched.

## Verification (fresh observations)

- Rollout: `deployment "app" successfully rolled out`; `deployment.apps/app` now `3 requested | 3 ready | 3 updated | 3 available`, `generation=2 observed=2`, conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- Endpoints now populated: `app 10.244.0.10:80,10.244.0.8:80,10.244.0.9:80`; EndpointSlice entries all `ready=true serving=true` for the 3 new pods.
- Traffic through the Service actually succeeds: 3/3 requests to the ClusterIP returned `200` and 3/3 to the DNS name `http://app/` returned `200`, with the nginx index body returned ("Welcome to nginx!").
