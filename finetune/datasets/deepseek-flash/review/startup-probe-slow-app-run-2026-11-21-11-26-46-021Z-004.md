---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/startup-probe-slow-app/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# startup-probe-slow-app · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Write commands

1. `kubectl patch deploy app -n default --type=strategic -p '{`

## Final report

The Pod is now stable and Ready.

**Diagnosis (evidence):**

- Container command is `sh -c "sleep 20; exec nginx -g 'daemon off;'"` — it needs ~20s before nginx listens on :80.
- The Deployment had **no startup probe**, and the liveness probe was `initialDelaySeconds: 1, periodSeconds: 1, failureThreshold: 1`, so kubelet killed the container ~2s after start. Events: `Liveness probe failed: ... connection refused` → `Container app failed liveness probe, will be restarted` → `BackOff`; last state `Exit Code: 137`, 6 restarts, pod stuck `0/1 CrashLoopBackOff`.

**Fix (smallest justified change):** patched only the `app` container of Deployment `default/app` to add a startup probe that gates liveness/readiness until the app is up:

```yaml
startupProbe:
  httpGet: { path: /, port: 80, scheme: HTTP }
  periodSeconds: 5
  timeoutSeconds: 2
  successThreshold: 1
  failureThreshold: 12 # 60s budget >> 20s startup
```

Everything else was left intact: `sleep 20; exec nginx...` command, liveness probe, readiness probe, `strategy: Recreate`, `terminationGracePeriodSeconds: 1`, image `nginx:1.31.5`, replicas 1.

**Verification (fresh observations):**

- `rollout status`: "deployment \"app\" successfully rolled out".
- Replica comparison matches requested: `spec.replicas=1, updated=1, ready=1, available=1, unavailable` gone; Deployment `Available` is now satisfied.
- Pod `app-77f9d9978f-frjk8`: `1/1 Running`, `Ready=True`, `restarts=0`, `startupProbe=12x5s` (the single "Startup probe failed" event at t+20s is the expected probe during the 20s sleep, within the 60s budget — no liveness kill followed).
- Re-checked 40s+ later: still `1/1 Running`, restarts still `0`, counts still `1 1 1 1` — stable across multiple liveness intervals.
- Traffic actually served: in-pod request to `127.0.0.1:80` returned `HTTP/1.1 200 OK`, `Server: nginx/1.31.5`.
