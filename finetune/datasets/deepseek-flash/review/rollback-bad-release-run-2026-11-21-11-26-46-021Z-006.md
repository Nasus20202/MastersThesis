---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/rollback-bad-release/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl -n default rollout undo deploy/app`

## Final report

The application is healthy again.

## What was wrong

Deployment `default/app` was intended to run 3 replicas but only had 2 available. Its newest revision (rev 2, ReplicaSet `app-76ffd9fd7f`) had replaced the working `nginx:1.31.5` container with `busybox:1.37.0` running `command: ["sh","-c","exit 1"]`:

- Pods `app-76ffd9fd7f-f8zxv` / `-r7mh9` were `0/1 Error`, exit code 1, restart-looped (BackOff) — they could never pass readiness, so the rolling update stalled with `unavailableReplicas: 2`.
- Revision 1 (`app-575b7496bc`, `nginx:1.31.5`, no override command) was the known-good template.

## Change made

Rolled the Deployment back to the last working revision — one targeted change to the owning resource:

```
kubectl -n default rollout undo deploy/app
```

## Verification (fresh observations)

| Check                       | Result                                                                    |
| --------------------------- | ------------------------------------------------------------------------- |
| Replica counts              | desired **3** / updated **3** / ready **3** / available **3** — all match |
| Running image & command     | all 3 pods: `image=nginx:1.31.5`, `cmd=` (no override)                    |
| Pod readiness               | 3 pods `1/1 Running`, 0 restarts                                          |
| Failed ReplicaSet           | `app-76ffd9fd7f` scaled to **0** replicas, no failing pods remain         |
| Deployment conditions       | `Available=True`, `Progressing=True`                                      |
| App actually serves traffic | `curl http://localhost:80/` inside a pod → **HTTP 200**                   |

Note: there is no Service for this app in the namespace (only `kubernetes`), so an in-pod HTTP request is the direct end-to-end check available.

Constraints preserved: `replicas: 3`, `maxSurge/maxUnavailable: 1/1`, `minReadySeconds: 5`, `progressDeadlineSeconds: 600` — no unrelated settings were modified, and the fix was a durable template rollback rather than deleting transient failed pods.
