---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/liveness-restart-loop/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# liveness-restart-loop · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 1

## Task

The application Pods keep restarting and never stay Ready. Restore them to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define a liveness HTTP request

## Write commands

1. `edit_file {"new_string": " livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 1\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1", "old_string": " livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /healthz\n port: 8080\n scheme: HTTP\n initialDelaySeconds: 1\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1", "path": "default/app.yaml"}`
2. `kubectl patch deployment app -n default --type=strategic-merge -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","livenessProbe":{"httpGet":{"path":"/","port":80},"failureThreshold":1,"initialDelaySeconds":1,"periodSeconds":1,"successThreshold":1,"timeoutSeconds":1}}]}}}}'`
3. `kubectl patch deployment app -n default --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","livenessProbe":{"httpGet":{"path":"/","port":80},"failureThreshold":1,"initialDelaySeconds":1,"periodSeconds":1,"successThreshold":1,"timeoutSeconds":1}}]}}}}'`

## Final report

## Summary

**Root Cause**: The liveness probe was configured to check `http://:8080/healthz`, but nginx listens on port 80 by default and doesn't have a `/healthz` endpoint. This caused the liveness probe to fail immediately on every startup, triggering constant restarts (CrashLoopBackOff).

**Fix Applied**: Changed the liveness probe configuration in the deployment from:

- `port: 8080`, `path: /healthz`

to:

- `port: 80`, `path: /`

This matches nginx's default port and provides a valid health check endpoint.

**Verification**:

| Check                | Before             | After               |
| -------------------- | ------------------ | ------------------- |
| Pod Status           | CrashLoopBackOff   | Running             |
| Ready Count          | 0/1                | 1/1                 |
| Restarts             | 6                  | 0                   |
| Deployment Available | 0/2                | 2/2                 |
| HTTP Response        | Connection refused | nginx HTML returned |

The pods are now stable and in a Ready state.
