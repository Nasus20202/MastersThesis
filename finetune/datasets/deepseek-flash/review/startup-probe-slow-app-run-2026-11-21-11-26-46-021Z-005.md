---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/startup-probe-slow-app/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# startup-probe-slow-app · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Write commands

1. `kubectl patch deployment app -n default --type=strategic -p "$(cat patch.json)" 2>`

## Final report

## Root cause

The `app` Deployment's live pod template (revision 2) ran `sh -c "sleep 20; exec nginx -g 'daemon off;'"` but had **no `startupProbe`** and an extremely aggressive liveness probe:

- `livenessProbe`: `initialDelaySeconds=1, periodSeconds=1, failureThreshold=1`
- Killed at ~1s with `Exit Code: 137` → `CrashLoopBackOff` (5 restarts); readiness probe never had a chance to succeed.

The pod's event history confirmed: `Container app failed liveness probe, will be restarted` and `Readiness probe failed: ... connection refused`. Revision 1 (retained in the old ReplicaSet `app-9b794865c`) had a correct `startupProbe` with a >20s budget; the change to revision 2 is what broke startup.

## Fix applied

Strategic merge patch on `deployment/app` (the owner of the faulty pod template) — probes only:

- Added `startupProbe`: HTTP GET `/` :80, `periodSeconds=2`, `failureThreshold=15` → 30s startup budget, covering the 20s `sleep`.
- Relaxed liveness back to the working revision-1 values: `periodSeconds=10`, `failureThreshold=3` (was 1s/1).

Image (`nginx:1.31.5`), command, ports, replicas, `Recreate` strategy, and resources were left untouched.

## Verification (fresh observations)

- After the rollout the pod stayed `Running 1/1`, `RESTARTS=0` — checked at 24s, 68s, and 112s of pod age (well past the 20s startup window and multiple liveness cycles).
- Startup probe behaved as intended: 10 `connection refused` failures during the `sleep`, then succeeded — no kill.
- Controller counts match the requested replica count: `replicas=1 updated=1 ready=1 available=1`, `generation=3/3`, Deployment conditions `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`; `kubectl rollout status` → "successfully rolled out".
- Application actually serves traffic: `curl` through a port-forward to the pod returned **HTTP 200** with the nginx welcome page.

The original symptom (pod killed before serving, never Ready) is gone: the pod reached Ready, served a real HTTP request, and has not restarted.
