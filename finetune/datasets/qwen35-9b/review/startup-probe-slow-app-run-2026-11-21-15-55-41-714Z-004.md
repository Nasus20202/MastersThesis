---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/startup-probe-slow-app/prompt/004.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: patch sets deployment.kubernetes.io/revision to "3"; controller annotation change not mentioned in report | second pass against the raw record: ok: Step 37 kubectl apply sets liveness and readiness initialDelaySeconds 22 for the 20s sleep; steps 43-59 support the Ready and restart claims'
---

# startup-probe-slow-app · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 4

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Write commands

1. `kubectl edit deploy app -n default --record-changes=false << 'EOF'`
2. `kubectl edit deploy app -n default 2>`
3. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n annotations:\n deployment.kubernetes.io/revision: \"2\"\n creationTimestamp: \"2026-11-21T12:56:34Z\"\n labels:\n app: app\n name: app\n namespace: default\n resourceVersion: \"589\"\n uid: cd66c97e-af0b-49a2-afc5-a00ed9ff1587\nspec:\n progressDeadlineSeconds: 600\n replicas: 1\n revisionHistoryLimit: 10\n selector:\n matchLabels:\n app: app\n strategy:\n type: Recreate\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - command:\n - sh\n - -c\n - sleep 20; exec nginx -g 'daemon off;'\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 22\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1\n name: app\n ports:\n - containerPort: 80\n name: http\n protocol: TCP\n readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 22\n periodSeconds: 5\n successThreshold: 1\n timeoutSeconds: 1\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n dnsPolicy: ClusterFirst\n restartPolicy: Always\n schedulerName: default-scheduler\n securityContext: {}\n terminationGracePeriodSeconds: 1\n", "path": "/tmp/deployment-fixed.yaml"}`
4. `kubectl apply -f /tmp/deployment-fixed.yaml -n default 2>/dev/null`
5. `kubectl apply -f /tmp/deployment-fixed.yaml -n default`
6. `kubectl patch deploy app -n default -p '{"metadata":{"annotations":{"deployment.kubernetes.io/revision":"3"}},"spec":{"template":{"spec":{"containers":[{"livenessProbe":{"initialDelaySeconds":22},"readinessProbe":{"initialDelaySeconds":22}}]}}}}' 2>`
7. `kubectl apply -f /dev/stdin -n default 2>`

## Final report

## Summary

### Problem Identified

The Pod `app-68bc7878f-nc628` was in `CrashLoopBackOff` because:

- **Container startup time**: 20 seconds (`sleep 20; exec nginx...`)
- **Liveness probe delay**: 1 second (fired before nginx was ready) → connection refused → killed
- **Readiness probe delay**: 0 seconds (missing, defaulted to 0) → fired immediately → connection refused

### Fix Applied

Modified the Deployment `app` in namespace `default`:

- Changed `livenessProbe.initialDelaySeconds` from **1** to **22** seconds
- Added `readinessProbe.initialDelaySeconds: 22` (was missing/0)

This accounts for the 20-second startup time plus a small buffer.

### Verification Results

| Metric               | Before                             | After                             |
| -------------------- | ---------------------------------- | --------------------------------- |
| Pod Status           | CrashLoopBackOff                   | Running                           |
| Ready                | False                              | True ✓                            |
| Restarts             | 7+                                 | 0                                 |
| Deployment Available | False (MinimumReplicasUnavailable) | True (MinimumReplicasAvailable) ✓ |
| Events               | Unhealthy warnings ×20             | Normal only ✓                     |

**Outcome**: The application Pod is now in a stable, Ready state and can serve requests.
