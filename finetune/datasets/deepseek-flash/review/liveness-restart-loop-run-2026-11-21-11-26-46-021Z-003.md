---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/liveness-restart-loop/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# liveness-restart-loop · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

The application Pods keep restarting and never stay Ready. Restore them to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define a liveness HTTP request

## Write commands

1. `kubectl apply -f - 2>`
2. `kubectl delete pod probeinsp -n default --wait=false 2>`
3. `kubectl patch deploy app -n default --type=strategic -p '{`

## Final report

Both Pods are stable and Ready. Here's what was wrong and what I changed.

## Diagnosis

The `default/app` Deployment (2 replicas, image `nginx:1.31.5`) had a liveness probe that could never succeed:

```yaml
livenessProbe:
  httpGet: { path: /healthz, port: 8080 } # wrong port AND wrong path
  initialDelaySeconds: 1
  periodSeconds: 1
  failureThreshold: 1 # kill on a single 1s failure
```

Evidence from the cluster, not assumptions:

- Events: `Liveness probe failed: Get "http://10.244.0.7:8080/healthz": dial tcp ...:8080: connect: connection refused` → `Container app failed liveness probe, will be restarted` → `BackOff`.
- Inside a diagnostic Pod running the same image: nginx config has only `listen 80;`, nothing listens on 8080 (`000 / connection refused`), `/` returns **200**, and `/healthz` returns **404** (it doesn't exist in this image, so it would fail even on the right port).
- Because `initialDelaySeconds: 1` / `periodSeconds: 1` / `failureThreshold: 1` killed the container ~1s after start, nginx never got time to bind port 80 — which is why the _readiness_ probe also reported `connection refused` on port 80. Restart counts climbed to 3+ and `Available` was `False` with `0/2` ready.

## Change (single patch to the owning Deployment)

```yaml
livenessProbe:
  httpGet: { path: /, port: 80, scheme: HTTP }
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 2
  failureThreshold: 3
  successThreshold: 1
```

The probe now targets the app's real listener and endpoint, with startup grace and tolerance so a slow start can't trigger an immediate kill. The readiness probe (port 80, `/`) was already correct and was left untouched, as were image, replicas, ports, and strategy.

## Verification (fresh observations)

| Check                               | Result                                                                                                                                                                                 |
| ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Pods                                | `app-7cbdb899bf-4n95n` **1/1 Running, RESTARTS 0**; `app-7cbdb899bf-69ck8` **1/1 Running, RESTARTS 0**                                                                                 |
| Replica match                       | desired=**2** updated=**2** ready=**2** available=**2** unavailable=_(none)_                                                                                                           |
| Deployment conditions               | `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`                                                                                               |
| Pod `Ready` condition               | `True` for both, `restartCount=0`                                                                                                                                                      |
| Liveness probe path actually serves | In-pod request to `http://127.0.0.1/` → **200** on both Pods                                                                                                                           |
| Sustained stability                 | Re-checked at ~97s age (≈9 liveness periods): still 0 restarts; no `Unhealthy`, `Killing`, or `BackOff` events after the new Pods started (latest events are normal startup/scaledown) |

The transient `Readiness probe failed` warnings for the new Pods were one-time startup samples before nginx bound port 80 — they stopped, and no liveness failures occurred at all on the new ReplicaSet. The throwaway diagnostic Pod `probeinsp` was deleted; only the two application Pods remain.
