---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/liveness-restart-loop/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# liveness-restart-loop · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

The application Pods keep restarting and never stay Ready. Restore them to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define a liveness HTTP request

## Write commands

1. `kubectl patch deploy app -n default --type=json -p='[`

## Final report

Pods are stable and Ready. Summary of what I found and changed:

## Diagnosis

- `deployment/app` (namespace `default`) had 2 pods in `CrashLoopBackOff`, `0/2` available.
- Events showed the exact failure: `Liveness probe failed: Get "http://<pod-ip>:8080/healthz": dial tcp ...:8080: connect: connection refused` → `Container app failed liveness probe, will be restarted`.
- Container logs confirmed nginx itself was healthy: `nginx/1.31.5 ... start worker processes`, then `signal 3 (SIGQUIT) received, shutting down` ~2s later (kubelet killing it), exit code 0.
- Comparing Deployment revisions pinned the regression: revision 1 (`app-6fd7fb94bd`) had a liveness probe on **port 80, path `/`**; revision 2 (`app-5f7667d7f8`) changed it to **port 8080, path `/healthz`** with `initialDelaySeconds: 1, periodSeconds: 1, failureThreshold: 1`. The container only listens on port 80, so the probe could never succeed and killed the container within ~2 seconds — before the pod could ever stay Ready. The readiness probe (port 80, `/`) was already correct.

## Change (smallest justified fix)

Patched only the container's `livenessProbe` back to the app's real listener, matching the previously working revision:

```
httpGet: {path: /, port: 80, scheme: HTTP}, initialDelaySeconds: 5, periodSeconds: 10, timeoutSeconds: 1, failureThreshold: 3
```

Image (`nginx:1.31.5`), replicas (`2`), readiness probe, and all other settings were left untouched.

## Verification (fresh observations)

- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out`; new ReplicaSet `app-6fd7fb94bd` rolled out both pods.
- Deployment status: `READY 2/2, UP-TO-DATE 2, AVAILABLE 2`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- Pods: `app-6fd7fb94bd-k877w` and `app-6fd7fb94bd-mzmw6` both `READY=true`, `RESTARTS=0`, continuously running since 12:37:04Z / 12:37:05Z (over 2.5 minutes, past the old ~2s kill window and multiple 10s liveness periods).
- Direct request check through the pod: `GET /api/v1/namespaces/default/pods/<pod>:80/proxy/` returned the nginx welcome page (HTTP 200), confirming the app actually serves traffic on the probed port.
- No new `Unhealthy`/`BackOff`/`Killing` events exist after the patch; the only probe-failure events are old ones from the now-deleted broken pods.
