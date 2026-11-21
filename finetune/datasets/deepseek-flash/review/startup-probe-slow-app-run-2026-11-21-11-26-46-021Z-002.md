---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/startup-probe-slow-app/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# startup-probe-slow-app · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Write commands

1. `kubectl patch deploy app -n default --type=strategic --patch '`

## Final report

The Pod is restored to a stable, Ready state.

## Diagnosis

- Pod `default/app-68bc7878f-l8xjw` was in `CrashLoopBackOff` with 6 restarts, exit code **137** (SIGKILL).
- Events showed the sequence: `Container started` → 2s later `Liveness probe failed: connect: connection refused` → `Container app failed liveness probe, will be restarted`.
- The container command is `sh -c "sleep 20; exec nginx -g 'daemon off;'"`, i.e. nginx does not listen on :80 for ~20s.
- The Deployment's probe config made that impossible to survive: liveness had `initialDelaySeconds: 1, periodSeconds: 1, failureThreshold: 1` and there was **no startupProbe**, so the kubelet killed the container ~2s into its 20s warm-up. Readiness alone can't protect against this — only a startup probe or a longer liveness delay can.

## Change (single patch to the owning Deployment `default/app`)

- Added a `startupProbe` on `GET / :80` with `periodSeconds: 2, timeoutSeconds: 2, failureThreshold: 30` (60s startup budget) — it gates liveness/readiness until the app is actually serving, and leaves the 20s sleep untouched.
- Made liveness non-self-defeating for a slow/loaded app: `initialDelaySeconds: 5, periodSeconds: 5, timeoutSeconds: 2, failureThreshold: 3` (was 1/1/1/1).
- Image, command, readiness probe, replicas, and strategy were left unchanged; the rollout used the Deployment's existing `Recreate` strategy, no Pod was deleted manually.

## Verified outcome

- Pod `app-7fd79d8bb5-2tx92`: `READY=true`, `RESTARTS=0`, image `docker.io/library/nginx:1.31.5`, running continuously since `11:59:48Z` (re-checked ~90s later, still 0 restarts).
- Events for the new Pod: only `Scheduled/Pulled/Created/Started` plus `Startup probe failed` during the expected warm-up window; **no `Killing` / liveness-failure events**, and the startup probe succeeded once nginx came up (never reached its 30-failure budget).
- Deployment `app`: `replicas=1 updated=1 ready=1 available=1 unavailable=none`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- Direct request through the Pod succeeded: `curl http://10.244.0.7:80/` → **HTTP 200** with the nginx index page (there is no Service in this namespace, so the Pod IP is the endpoint to test).
