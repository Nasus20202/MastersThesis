---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/startup-probe-slow-app/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# startup-probe-slow-app · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Write commands

1. `kubectl patch deployment app -n default --type=strategic --patch-file /tmp/probes.json 2>`

## Final report

## Root cause

The `app` Deployment (`default` namespace) runs `nginx:1.31.5` with the command `sh -c "sleep 20; exec nginx -g 'daemon off;'"`, so the container does not listen on port 80 for ~20 seconds. Revision 2 of the pod template had **no `startupProbe`** and an extremely aggressive liveness probe:

```
livenessProbe: initialDelaySeconds: 1, periodSeconds: 1, failureThreshold: 1
```

Liveness began probing 1s after start and killed the container on the first failure — ~2s into a 20s startup, before nginx could ever bind port 80. Evidence: `Restart Count: 6`, `Last State: Terminated, Reason: Error, Exit Code: 137`, and events `Container app failed liveness probe, will be restarted` interleaved with `dial tcp 10.244.0.6:80: connect: connection refused` and `BackOff`. The container was killed on every attempt, so the readiness probe never had a chance to pass.

Revision 1 (`app-9b794865c`) showed the intended, previously working config: a `startupProbe` (period 2s × failureThreshold 15 = 30s budget) plus a tolerant liveness probe (period 10s, failureThreshold 3).

## Fix

One `kubectl patch` to the Deployment's pod template, restoring only the probe block to the known-good revision-1 values:

- **Added** `startupProbe`: `httpGet / :80`, `periodSeconds: 2`, `failureThreshold: 15` (30s startup budget > 20s actual startup). Liveness/readiness are now gated until it succeeds.
- **Relaxed** `livenessProbe` from `initialDelaySeconds: 1 / periodSeconds: 1 / failureThreshold: 1` to `periodSeconds: 10 / failureThreshold: 3` (default `initialDelaySeconds: 0`).

Everything else was left as-is: image `nginx:1.31.5`, the `sleep 20` command, port 80/`http`, readiness probe, `strategy: Recreate`, `terminationGracePeriodSeconds: 1`, `replicas: 1`.

## Verification (direct observation)

- **Serving traffic:** `kubectl exec` into the new pod → `HTTP 200 from localhost` on `http://127.0.0.1:80/`.
- **Startup probe now protects startup instead of killing:** events show `Startup probe failed: ... connection refused` at 5s, and **no** subsequent `Killing` / `liveness probe failed` / `BackOff` events.
- **Controller counts match:** `spec.replicas=1, status.replicas=1, updated=1, ready=1, available=1, unavailable=<none>`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`; `kubectl rollout status` reported `deployment "app" successfully rolled out`.
- **Stable and Ready:** `app-9b794865c-zltmc` is `1/1 Running` with `restarts=0`, `ready=true` at 74s of age — previously it was killed roughly every 30s.
- **No collateral damage:** all `kube-system` pods still `Running` with `restarts=0`.

The originally reported symptom — a slow-starting pod killed before it can serve requests and never Ready — is gone: the pod now survives its 20-second startup and remains Ready with zero restarts.
