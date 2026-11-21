---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/liveness-restart-loop/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# liveness-restart-loop · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

The application Pods keep restarting and never stay Ready. Restore them to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define a liveness HTTP request

## Write commands

1. `edit_file {"new_string": " livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 1\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1", "old_string": " livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /healthz\n port: 8080\n scheme: HTTP\n initialDelaySeconds: 1\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1", "path": "/tmp/deployment.yaml"}`
2. `kubectl replace --raw /apis/apps/v1/namespaces/default/deployments/app -f /tmp/deployment.yaml 2>/dev/null`
3. `kubectl apply -f /tmp/deployment.yaml`
4. `kubectl run --rm -it test-client --image=busybox:1.35 --restart=Never -- wget -qO- http://app-6b9ff5f876-2ptbp.default.svc.cluster.local 2>/dev/null`

## Final report

## Summary

**Root Cause Identified:**
The application pods were failing in `CrashLoopBackOff` due to probe misconfiguration:

- **Liveness probe** was configured to check `http://:8080/healthz`, but the nginx container only exposes port 80
- This caused immediate liveness probe failures (`connection refused`), triggering container restarts every ~1 second
- The readiness probe at `http://:80/` also failed initially due to startup timing, but this resolved as nginx started

**Fix Applied:**
Modified the deployment's liveness probe from:

```yaml
livenessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

to:

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 80
```

**Verification:**

- Both pods now show `1/1 Ready` with `Running` status
- No restarts occurred after the fix (0 restart count on new pods)
- Deployment is fully available: 2/2 ready, up-to-date, and available
- Pod age is stable (>90 seconds) with no recent Unhealthy events

The application pods have been restored to a stable, Ready state.
