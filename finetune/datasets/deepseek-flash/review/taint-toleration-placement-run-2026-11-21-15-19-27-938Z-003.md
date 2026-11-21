---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-15-19-27-938Z/taint-toleration-placement/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# taint-toleration-placement · deepseek-flash · run-2026-11-21-15-19-27-938Z · attempt 3

## Task

The cluster node carries the taint dedicated=gpu:NoSchedule. Deploy a Deployment named app with image nginx:1.31.5 and three replicas that tolerates this taint, without removing the taint from the node.

Documented in:

- content/en/docs/concepts/scheduling-eviction/taint-and-toleration.md: Concepts, Example Use Cases

## Write commands

1. `kubectl apply -f /tmp/work/app.yaml`

## Final report

Done. The Deployment was created with a toleration matching the node taint, and all checks pass with fresh observations.

**What was deployed** (`/tmp/work/app.yaml`):

- Deployment `app`, 3 replicas, image `nginx:1.31.5`
- Pod template toleration: `key=dedicated, operator=Equal, value=gpu, effect=NoSchedule` — an exact match for the node's taint

**Verified evidence:**

| Check                                                         | Observed value                                                                                                                 |
| ------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Deployment replicas (requested / updated / ready / available) | `3 / 3 / 3 / 3` — all match                                                                                                    |
| Pods status                                                   | 3 pods `1/1 Running`, 0 restarts, all scheduled to `benchmark-taint-toleration-placement-beca24c4-control-plane`               |
| Running image                                                 | `docker.io/library/nginx:1.31.5` on all 3 pods (intended image)                                                                |
| Toleration persisted                                          | `[{"key":"dedicated","operator":"Equal","value":"gpu","effect":"NoSchedule"}]` present on the Deployment spec and on every pod |
| Serving traffic                                               | `HTTP 200` from `curl http://localhost/` inside pod `app-d549ddddb-67zdp`                                                      |
| Node taint intact                                             | `[{"key":"dedicated","value":"gpu","effect":"NoSchedule"}]` — unchanged, not removed                                           |

No changes were made to the node or any other resource; the only change was creating the `app` Deployment.
