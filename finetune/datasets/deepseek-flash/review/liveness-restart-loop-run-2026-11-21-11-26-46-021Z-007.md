---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/liveness-restart-loop/prompt/007.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: kubectl patch -p "$(cat patch.json)"; patch.json is never written by a listed command, so its content is unverifiable | second pass against the raw record: ok: kubectl patch livenessProbe to httpGet / port 80 (msg 14); msgs 10, 13, 26, 29 match the report''s events, rev-1 probe and 0 restarts.'
---

# liveness-restart-loop · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The application Pods keep restarting and never stay Ready. Restore them to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define a liveness HTTP request

## Write commands

1. `kubectl patch deployment app -n default --type=strategic -p "$(cat patch.json)" 2>`
2. `kubectl delete pod app-5f7667d7f8-zwczw -n default --ignore-not-found 2>`

## Final report

## Root cause

`deployment/app` (namespace `default`) pointed its **liveness probe at port 8080 path `/healthz`**, but the container is `nginx:1.31.5` which only serves on **port 80**. Combined with `initialDelaySeconds: 1`, `periodSeconds: 1`, `failureThreshold: 1`, the kubelet killed nginx roughly a second after start, over and over:

- Events: `Liveness probe failed: Get "http://10.244.0.7:8080/healthz": connection refused` → `Killing ... Container app failed liveness probe, will be restarted`
- Pod status: `Last State: Terminated, Reason: Completed` (nginx shutting down gracefully on SIGTERM) → `CrashLoopBackOff`, `Ready: False`
- The pod template history confirmed the regression: revision 1 (`app-6fd7fb94bd`) probed port 80 `/`; revision 2 (`app-5f7667d7f8`) had been changed to port 8080 `/healthz`. The readiness probe (port 80 `/`) was correct — its "connection refused" messages were just startup timing.

## Change made

Single strategic patch to the owning Deployment — restored only the liveness probe to the listening port and non-pathological timings:

```
livenessProbe: httpGet path=/ port=80 scheme=HTTP, initialDelaySeconds=5, periodSeconds=10, failureThreshold=3, timeoutSeconds=1
```

Untouched: `replicas: 2`, image `nginx:1.31.5`, containerPort 80, readiness probe, resources, strategy.

## Verified outcome

- `kubectl rollout status` → `deployment "app" successfully rolled out`
- Deployment: `replicas=2 ready=2 updated=2 available=2 unavailable=` ; conditions `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`
- Pods stable ~2 minutes after the new ReplicaSet came up: `app-6fd7fb94bd-457cg` and `app-6fd7fb94bd-n7n54` → `Running`, `READY 1/1`, `Ready=True`, **restarts=0**
- Direct request to each pod: `curl http://127.0.0.1/` → `http_code=200` with the nginx index page
- Only remaining `Unhealthy` events are one-time readiness failures logged during container start (117s old, before nginx bound port 80); no liveness failures since the patch. The stale `Completed` pod from the broken ReplicaSet was removed; its ReplicaSet is at 0 desired replicas, so it cannot reappear.
