---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-15-19-27-938Z/taint-toleration-placement/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# taint-toleration-placement · deepseek-flash · run-2026-11-21-15-19-27-938Z · attempt 6

## Task

The cluster node carries the taint dedicated=gpu:NoSchedule. Deploy a Deployment named app with image nginx:1.31.5 and three replicas that tolerates this taint, without removing the taint from the node.

Documented in:

- content/en/docs/concepts/scheduling-eviction/taint-and-toleration.md: Concepts, Example Use Cases

## Write commands

1. `kubectl apply -f -`
2. `kubectl run check --rm -i --restart=Never --image=curlimages/curl:latest --overrides='{"spec":{"tolerations":[{"key":"dedicated","operator":"Equal","value":"gpu","effect":"NoSchedule"}]}}' -- curl -s -o /dev/null -w '%{http_code}\n' http://10.244.0.2 2>`

## Final report

Deployment created and verified.

**What I did**

- Confirmed the node's taint: `kubectl get node -o jsonpath='{.items[*].spec.taints}'` → `[{"effect":"NoSchedule","key":"dedicated","value":"gpu"}]` (single node `benchmark-taint-toleration-placement-82e141dd-control-plane`).
- Created `Deployment/app` with `replicas: 3`, container image `nginx:1.31.5`, and a matching toleration:
  ```yaml
  tolerations:
    - key: dedicated
      operator: Equal
      value: gpu
      effect: NoSchedule
  ```

**Verification (fresh observations)**

- Replica counts match exactly: `3 desired, 3 updated, 3 ready, 3 available`; `kubectl get deploy app` → `READY 3/3`, `UP-TO-DATE 3`, `AVAILABLE 3`.
- All pods scheduled onto the tainted node and `Running`/`Ready=True`:
  - `app-d549ddddb-2r99s`, `app-d549ddddb-cddql`, `app-d549ddddb-rpzkx` — all on `...-control-plane`.
- Running image is the intended one: each pod reports `spec.containers[0].image = nginx:1.31.5` with imageID `docker.io/library/nginx@sha256:05b8cb60c354a44ab824ea6e7dc69b46d50762cdbe728a347a5b656e6fb3d7c4`.
- App actually serves traffic: `kubectl exec deploy/app -- curl http://localhost:80/` → HTTP `200`.
- Constraint preserved: node taint still present and unchanged → `[{"effect":"NoSchedule","key":"dedicated","value":"gpu"}]`; no node modifications were made.
