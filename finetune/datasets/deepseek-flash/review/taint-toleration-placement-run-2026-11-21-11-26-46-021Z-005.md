---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/taint-toleration-placement/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# taint-toleration-placement · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

The cluster node carries the taint dedicated=gpu:NoSchedule. Deploy a Deployment named app with image nginx:1.31.5 and three replicas that tolerates this taint, without removing the taint from the node.

Documented in:

- content/en/docs/concepts/scheduling-eviction/taint-and-toleration.md: Concepts, Example Use Cases

## Write commands

1. `kubectl apply -f - 2>`

## Final report

Done. The Deployment was created with a toleration rather than altering the node.

**What I found**

- One node (`benchmark-taint-toleration-placement-ed287dfc-control-plane`, Ready) carrying exactly one taint: `dedicated=gpu:NoSchedule`. No other NoSchedule taints (the control-plane taint is absent), so a toleration is sufficient for scheduling.

**Change made (smallest justified)**

- Created `deployment.apps/app` in `default`: `replicas: 3`, container image `nginx:1.31.5`, with a single toleration `dedicated=gpu:NoSchedule` (`operator: Equal`, `value: gpu`). Node spec untouched.

**Verification (fresh observations)**

- Replica counts match the request: `desired=3 updated=3 ready=3 available=3`; Deployment conditions `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- All 3 pods are `1/1 Running` on the tainted node: `app-d549ddddb-67tkq`, `-gktsc`, `-j6fdg`.
- Running image is the intended one on every pod: spec `nginx:1.31.5`, status `docker.io/library/nginx:1.31.5`, `nginx -v` reports `nginx/1.31.5`.
- Service actually works: `curl http://localhost/` inside a pod returned `http=200`.
- Constraint preserved: node taints still exactly `{"effect":"NoSchedule","key":"dedicated","value":"gpu"}`.
- Toleration persisted in the pod template: `[{"effect":"NoSchedule","key":"dedicated","operator":"Equal","value":"gpu"}]`.
