---
attempt: finetune/datasets/continuation/qwen35-9b/runs/run-2026-11-21-22-42-05-380Z/startup-probe-slow-app/prompt/001.json
verdict: accepted
reason: Probe delays raised for the 20s startup; report matches the pod's probe settings and zero restarts.
---
# startup-probe-slow-app · Qwen3.5-9B-Q4_K_M · run-2026-11-21-22-42-05-380Z · attempt 1

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

1. `kubectl create deployment app -n default --image=nginx:1.31.5 -- --sh -c sleep 20`
2. `kubectl delete deployment app -n default 2>/dev/null`
3. `kubectl apply -f /tmp/app.yaml`
4. `kubectl delete deployment app -n default`
5. `kubectl apply -f /tmp/app.yaml`

## Final report

The problematic Pod in the `default` namespace (`app-68bc7878f-87sn7`) was failing to become Ready due to a `CrashLoopBackOff` caused by premature readiness probe failures during startup (Exit Code 137). The application used a long sleep command (`sleep 20; exec nginx...`) that required more time than the default probe delays allowed.

I restored the service by recreating the Deployment with adjusted probe settings:
-   **Readiness Probe:** Increased `initialDelaySeconds` to 30s, ensuring the check only occurs after the application has finished its startup sequence.
-   **Liveness Probe:** Set an appropriate `initialDelaySeconds` of 25s to allow for initialization.

**Outcome and Verification:**
The new Pod (`app-6f894cfbcc-5sflg`) is now in a stable, Ready state:
-   **Status:** Running
-   **Ready:** True
-   **Restart Count:** 0 (confirming the previous crash loop is resolved)

No further restarts or probe failures were observed after the fix was applied.
