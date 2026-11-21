---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/startup-probe-slow-app/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# startup-probe-slow-app · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Write commands

1. `kubectl patch deploy app -n default --type=strategic -p '{`
2. `kubectl run httpcheck --rm -i --restart=Never --image=curlimages/curl:latest -n default --command -- curl -sS -m 10 -o /dev/null -w 'HTTP_CODE=%{http_code}\n' http://10.244.0.7:80/ 2>`

## Final report

The Pod is restored to a stable, Ready state.

## Root cause

`Deployment/app` runs a deliberately slow-starting container — `sh -c 'sleep 20; exec nginx -g "daemon off;"'` — so nginx does not bind port 80 until ~20s after container start. The Pod template had **no `startupProbe`**, so the `livenessProbe` was live from the very beginning with `initialDelaySeconds: 1`, `periodSeconds: 1`, `failureThreshold: 1`. That gave the app roughly **2 seconds** to answer before kubelet declared it dead.

Evidence from the event stream showed exactly this loop:

- `Startup probe failed ... connect: connection refused` (implying the app simply wasn't up yet) on the first Pod `app-9b794865c-czw8q`
- `Killing ... Container app failed liveness probe, will be restarted` → `BackOff restarting failed container` → `CrashLoopBackOff` with 5 restarts

The Pod was being killed before it could ever reach Ready, which is precisely the reported symptom.

## Change made

One minimal, additive patch to the Deployment's Pod template — a `startupProbe` that gates the liveness and readiness probes until the app finishes booting:

```json
{
  "httpGet": { "path": "/", "port": 80, "scheme": "HTTP" },
  "periodSeconds": 2,
  "timeoutSeconds": 1,
  "failureThreshold": 30,
  "successThreshold": 1
}
```

This grants a 60s startup budget against a 20s actual boot, with margin. I deliberately left the existing `livenessProbe`, `readinessProbe`, image, command, and all other settings untouched — the aggressive liveness values are harmless once a startup probe gates them, so removing or loosening them would have been an unnecessary change.

## Verification

- **Replicas match, as required for a controller:** `desired=1 updated=1 ready=1 available=1`, `unavailable` unset; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`.
- **Pod is stable, not merely briefly healthy:** `app-67d9584c9f-2vd7z` reached `READY 1/1`, `RESTARTS 0`, and held that state across three checks over ~2m15s — far beyond the ~2s at which it previously died, and through many liveness cycles. `started=true`, `ready=true`.
- **The app actually serves requests** (not just "looks healthy"): a temporary `curlimages/curl` Pod made a real HTTP request over the cluster network to the Pod IP and returned **`HTTP_CODE=200`**. The container's `/proc/net/tcp` independently confirms a listener on `0x0050` (port 80).
- **The kill loop is gone:** the fresh event log shows the startup probe tolerating the boot-time `connection refused`, then succeeding — and contains **no** `Killing ... failed liveness probe` event, which was the signature of the original failure. The old CrashLooping ReplicaSet `app-68bc7878f` was scaled to 0 by the rollout.
