---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/readiness-probe-breaks-endpoints/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `kubectl -n default patch deploy app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/","port":80,"scheme":"HTTP"}}}]}}}}'`
2. `kubectl -n default run svc-check --rm -i --restart=Never --image=curlimages/curl:8.10.1 --command -- sh -c 'for i in 1 2 3 4 5 6`

## Final report

Service availability is restored and verified.

## Diagnosis

The Pods were `Running` but `0/1` Ready, so the Service had no endpoints and all traffic was refused.

Evidence chain:

- `kubectl get pods`: 3 pods `Running` but `READY 0/1`, `RESTARTS 0` — the container itself was healthy.
- Deployment owned a readinessProbe of `httpGet /readyz:80`, and events repeated:
  `Readiness probe failed: HTTP probe failed with statuscode: 404`
- Verified from inside a pod that the image `nginx:1.31.5` serves `/` with **200** (default `index.html`) but **404** for `/readyz` — plain nginx has no such endpoint.
- Consequence: `kubectl get endpoints app` was empty and the EndpointSlice `app-zc8zw` listed all three addresses with `ready: false`.

Root cause: a readiness probe pointing at a path the application never serves. The Service, selector (`app=app`), `port 80` and `targetPort 80` were all correct and were left untouched.

## Change

One minimal patch to the owning Deployment `default/app` — corrected only the probe path (`/readyz` → `/`, the endpoint nginx actually answers with 200); port, image, replicas, and all other settings preserved:

```
kubectl -n default patch deploy app --type=strategic -p \
  '{"spec":{"template":{"spec":{"containers":[{"name":"app","readinessProbe":{"httpGet":{"path":"/","port":80,"scheme":"HTTP"}}}]}}}}'
```

No objects were hand-deleted or force-restarted — the ReplicaSet/Pods rolled over normally through the Deployment.

## Verification (fresh observations)

- Rollout: `deployment "app" successfully rolled out`; new pods `1/1 Running`, `RESTARTS 0`, image `nginx:1.31.5` (intended image confirmed).
- Replica match: `spec.replicas=3`, `ready=3`, `updated=3`, `available=3`; conditions `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- Endpoints populated: `app → 10.244.0.10:80,10.244.0.8:80,10.244.0.9:80`; EndpointSlice readiness `true,true,true`.
- Actual request through the Service from an in-cluster pod: `http://app.default.svc.cluster.local/` → **HTTP 200** on the first attempt (body `Welcome to nginx!`), and `http://10.96.25.61/` (ClusterIP) → **200**.
- No `Unhealthy` events in a 45-second window afterward; the remaining warnings in the event log are timestamped `13:52:00Z` startup/pre-rollout artifacts from the old ReplicaSet. Temporary `svc-check` pod was deleted (no leftover test objects).
