---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-20-57-03-137Z/startup-probe-slow-app/prompt/001.json
verdict: accepted
reason: 'Relaxes the liveness timing on the deployment and the reported Ready pod follows from it.'
---
# startup-probe-slow-app · deepseek-flash · run-2026-11-21-20-57-03-137Z · attempt 1

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:
- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Inherited stop

The first 21 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 0
stdout:

stderr:
No resources found in default namespace.

Its writes:

1. `kubectl scale deployment app -n default --replicas=2`
2. `kubectl patch deployment app -n default --type='json' -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/resources/requests/memory", "value": "256Mi"}]'`
3. `kubectl scale deployment app -n default --replicas=1`
4. `kubectl delete deployment app -n default`

## Teacher's write commands

1. `kubectl apply -f -`
2. `kubectl patch deployment app -n default --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","livenessProbe":{"initialDelaySeconds":25,"failureThreshold":3}}]}}}}'`

## Final report

The issue was resolved. Here's a summary of the diagnosis and the fix:

**Diagnosis:**
The Pod was managed by the `app` Deployment in the `default` namespace. It was in a `CrashLoopBackOff` state and exiting with **Exit Code 137**. 

The container's startup command is `sh -c sleep 20; exec nginx -g 'daemon off;'`, meaning it intentionally waits 20 seconds before starting Nginx. However, the original liveness probe was configured aggressively:
*   `initialDelaySeconds: 1`
*   `failureThreshold: 1`

Because the liveness probe began checking port 80 just 1 second after the container started and required a single successful check, it sent a **SIGKILL** to the container long before the 20-second sleep completed and Nginx could begin listening. This was the reason it repeatedly crashed before it could serve requests.

**Repair:**
I applied a strategic merge patch to the `app` Deployment to adjust the liveness probe, giving the slow-starting application the time it needs to initialize:
*   Increased `initialDelaySeconds` from `1` to `25`
*   Increased `failureThreshold` from `1` to `3`

**Verification:**
The changes were applied and a new Pod was rolled out successfully. Direct observations confirm the symptom is fully resolved:
*   **Deployment status:** `1/1 READY`, `1 UP-TO-DATE`, `1 AVAILABLE`
*   **Pod status:** `app-74f4fc4d67-4gm29` is `1/1 Running` with **0 restarts**
*   **Pod conditions:** `Ready: True`
*   **Image:** `nginx:1.31.5` (the intended image)
*   The readiness probe is now succeeding, proving the Pod is accepting and serving requests.
