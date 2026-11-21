---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/liveness-restart-loop/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: unsure: kubectl patch --patch-file patch.json; patch.json is never written by a listed command, so its content is unverifiable | second pass against the raw record: ok: kubectl patch livenessProbe to httpGet / port 80 (msg 11); msgs 9, 10, 17, 18 back the 8080 cause, rev-1 probe and stable pods."
---

# liveness-restart-loop · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

The application Pods keep restarting and never stay Ready. Restore them to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define a liveness HTTP request

## Write commands

1. `kubectl patch deployment app -n default --type=strategic --patch-file patch.json 2>`

## Final report

**Root cause found and fixed.**

The `app` Deployment's pod template in `default` had its liveness probe pointed at a port/path nginx doesn't serve:

- `livenessProbe: httpGet /healthz:8080`, `initialDelaySeconds: 1`, `periodSeconds: 1`, `failureThreshold: 1`
- Events confirmed: `Liveness probe failed: Get "http://10.244.0.7:8080/healthz": dial tcp ...:8080: connect: connection refused` → `Container app failed liveness probe, will be restarted`, repeated until `CrashLoopBackOff` (5 restarts). The pods never got past the aggressive 1-second liveness kill cycle, so they never became Ready.
- The previous ReplicaSet revision (`app-6fd7fb94bd`, revision 1) had the correct probe (`httpGet /:80`, delay 5s, period 10s), confirming the bad probe was an unintended change in revision 2.

**Change made (single, minimal):** strategically patched only the liveness probe on container `app` to `httpGet / on port 80`, `initialDelaySeconds=5`, `periodSeconds=10`, `timeoutSeconds=1`, `failureThreshold=3`. Image (`nginx:1.31.5`), readiness probe, and `replicas: 2` were left untouched.

Command: `kubectl patch deployment app --type=strategic` with only that container's `livenessProbe`, then `kubectl rollout status` reported `deployment "app" successfully rolled out`.

**Verified outcome (fresh observations):**

- `kubectl get deploy app`: `READY 2/2`, `UP-TO-DATE 2`, `AVAILABLE 2`; status jsonpath `replicas=2 updated=2 ready=2 available=2 unavailable=` (empty) — requested count matches ready/available.
- `kubectl get pods -l app=app`: both `app-6fd7fb94bd-8smdj` and `app-6fd7fb94bd-95wlr` are `1/1 Running` with `restarts=0 ready=true` after 97s/95s of uptime — well past several 10s liveness intervals, so no more kill loop.
- Direct in-pod check: `curl http://127.0.0.1:80/` returned `HTTP 200` on both pods, while `http://127.0.0.1:8080/healthz` returned connection refused (`HTTP 000`), proving the old probe endpoint never existed and the new one targets the working endpoint.
- Deployment conditions now `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`; old crashing ReplicaSet `app-5f7667d7f8` scaled to 0/0.

The pods are now stably Running and Ready with zero restarts, and the deployment's desired/updated/ready/available counts all equal 2.
