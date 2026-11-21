---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/rollback-bad-release/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl -n default rollout undo deploy/app 2>`

## Final report

The application is restored to a healthy state.

## What I found (evidence)

- `default/deployment app`: desired `replicas: 3`, but status showed `availableReplicas: 2`, `unavailableReplicas: 2`, and `rollout status` hung at _"2 out of 3 new replicas have been updated"_.
- Two ReplicaSets during the stuck rollout:
  - `app-575b7496bc` (revision 1) — `nginx:1.31.5`, 2/2 ready (last known-good).
  - `app-76ffd9fd7f` (revision 2) — `busybox:1.37.0` with container command `sh -c exit 1`.
- Pod `app-76ffd9fd7f-c9j5c` was `Error` with `BackOff restarting failed container app` (4 restarts); its sibling had 4 restarts and only flickered Ready because it had no readiness probe — not real availability.
- So the release itself was faulty (a container that exits 1), not a scheduling, image-pull, or capacity problem: `rollout history --revision=2` confirmed the bad command was the only delta vs. revision 1.

## Change made (smallest justified fix)

```
kubectl -n default rollout undo deploy/app      # → deployment.apps/app rolled back
```

This reverted the Deployment's pod template to the last known-good release (`nginx:1.31.5`, default entrypoint, port 80/http). No unrelated settings were touched — `replicas: 3`, `maxSurge: 1`, `maxUnavailable: 1`, `minReadySeconds: 5`, `progressDeadlineSeconds: 600` are unchanged.

## Verification (fresh observations)

- `rollout status`: `deployment "app" successfully rolled out` (repeated after a further 25s wait — still successful, i.e. stable).
- Replica counts match the request exactly: `3 spec / 3 total / 3 updated / 3 ready / 3 available`, `0 unavailable`; conditions `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)` for `app-575b7496bc`.
- Running image is the intended one: all three pods report `nginx:1.31.5` (imageID `docker.io/library/nginx@sha256:05b8cb60…`), with no `command` override in the template (previously `sh -c exit 1`).
- Functional check, not just object health: `curl` to `http://localhost:80/` inside a pod and to pod IP `10.244.0.10:80` from a peer pod both returned `HTTP 200`.
- Cluster-wide sweep shows no non-`Running`/`Completed` pods; the bad ReplicaSet `app-76ffd9fd7f` is scaled to `0/0/0` and retained only as revision history.
