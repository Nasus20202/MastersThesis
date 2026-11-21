---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/liveness-restart-loop/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# liveness-restart-loop · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

The application Pods keep restarting and never stay Ready. Restore them to a stable, Ready state.

Documented in:

- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Define a liveness HTTP request

## Write commands

1. `kubectl patch deployment app -n default --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","livenessProbe":{"httpGet":{"path":"/","port":80,"scheme":"HTTP"}}}]}}}}' 2>`

## Final report

## Diagnosis

The `app` Deployment (2 replicas, `nginx:1.31.5`) had a **liveness probe pointed at the wrong port**:

- Liveness: `httpGet :8080/healthz` — nginx never listens on 8080, so every probe hit `connection refused`. With `failureThreshold: 1` / `periodSeconds: 1`, the kubelet killed the container within ~1s of every start → `CrashLoopBackOff`, `restarts=4`, and the container never got far enough to become Ready.
- Readiness: `httpGet :80/` — correct, this was not the fault.

Evidence from events: `Liveness probe failed: Get "http://10.244.0.7:8080/healthz": dial tcp ...: connection refused` → `Container app failed liveness probe, will be restarted` → `BackOff restarting failed container`. Deployment status showed `Available=False (MinimumReplicasUnavailable)`, `unavailableReplicas: 2`.

## Change

One minimal patch to the owning Deployment, fixing only the liveness probe target to the port nginx actually serves (port `80`, path `/`). Image, ports, resource limits, readiness probe, and replica count were left untouched:

```
livenessProbe.httpGet: {path: /healthz, port: 8080}  →  {path: /, port: 80}
```

## Verification

| Check                    | Result                                                                                                                                                |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `kubectl rollout status` | `deployment "app" successfully rolled out`                                                                                                            |
| Deployment conditions    | `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`                                                              |
| Replica match            | desired=2, updated=2, ready=2, available=2, unavailable=none                                                                                          |
| Pods                     | `app-6b9ff5f876-hzdw2` 1/1 Running, `app-6b9ff5f876-z5qmm` 1/1 Running                                                                                |
| Restarts                 | `0` on both pods, sustained across repeated checks at 49s, 2m35s, and final observation                                                               |
| Actual serving check     | In-pod request to the probed port/path returned `HTTP/1.1 200 OK`                                                                                     |
| Probe failures since fix | None for liveness; the only post-fix `Unhealthy` events were the expected one-shot readiness failures during nginx's initial startup, gone afterwards |

The Pods now stay Running and Ready with zero restarts, and the Deployment reports full availability.
