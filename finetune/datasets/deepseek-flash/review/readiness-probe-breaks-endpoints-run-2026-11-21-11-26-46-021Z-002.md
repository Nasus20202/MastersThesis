---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/readiness-probe-breaks-endpoints/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# readiness-probe-breaks-endpoints · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

The application Pods are running, but the application is unavailable through its Kubernetes Service. Restore service availability.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define readiness probes

## Write commands

1. `kubectl patch deployment app --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/readinessProbe/httpGet/path","value":"/"}]' 2>`
2. `kubectl run svc-check-$$ --image=curlimages/curl:8.11.1 --restart=Never --rm -i --quiet --command -- sh -c 'for i in 1 2 3`
3. `kubectl run final-check-$$ --image=curlimages/curl:8.11.1 --restart=Never --rm -i --quiet --command -- curl -sS -o /dev/null -w "HTTP %{http_code}\n" http://app.default.svc.cluster.local/ 2>`

## Final report

Service availability is restored and verified.

## Root cause

The `app` Deployment's readiness probe targeted `httpGet: /readyz` on port 80, but the nginx container only serves `/` (200) — `/readyz` returned **404**. With the probe failing, all 3 Pods stayed `0/1` (Ready=False), so the `app` Service's Endpoints/EndpointSlice were empty and traffic had nowhere to go.

Evidence:

- `kubectl exec ... curl /readyz` → `404`; `curl /` → `200`
- Events: `Readiness probe failed: HTTP probe failed with statuscode: 404`
- Old state: `kubectl get endpoints app` → `ENDPOINTS` empty; Deployment `unavailable=3`, `Available=False (MinimumReplicasUnavailable)`

The Service itself was correct — selector `app=app`, `port 80`, `targetPort 80`, matching Pod labels — so the defect was confined to the probe on the owning Deployment.

## Change made (smallest justified)

Patched only the probe path on `deployment/app`; image, replicas, ports, thresholds, and Service spec untouched:

```
readinessProbe.httpGet.path: /readyz  ->  /
```

## Verification (fresh observations)

- Replica counts match exactly: `desired=3 updated=3 ready=3 available=3 unavailable=0`; conditions `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- All Pods `1/1 Running`: `app-6c75b77dc8-hj4tm`, `-kx2tn`, `-vnh9h`.
- Endpoints now populated: `app → 10.244.0.8:80, 10.244.0.9:80, 10.244.0.10:80`; EndpointSlice `app-9vlk5` lists the same three addresses.
- **Request through the Service succeeds**: 3/3 `curl http://app.default.svc.cluster.local/` → `HTTP 200`, body `<title>Welcome to nginx!</title>`; re-checked after 25s → `HTTP 200`.
- No ongoing probe failures: 0 `Unhealthy` events in the last 40s for the current ReplicaSet (remaining events are pre-fix 404s from the old RS and expected pre-listen startup refusals).
