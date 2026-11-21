---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/stuck-terminating-pod/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# stuck-terminating-pod · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A Pod named app is stuck in the Terminating state and cannot be removed. Replace the workload with a healthy Pod named app.

Documented in:

- content/en/docs/concepts/overview/working-with-objects/finalizers.md: How finalizers work

## Write commands

1. `kubectl patch pod app -n default --type=merge -p '{"metadata":{"finalizers":null}}' 2>`
2. `write_file {"content": "apiVersion: v1\nkind: Pod\nmetadata:\n name: app\n namespace: default\n labels:\n app: app\nspec:\n containers:\n - name: app\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n restartPolicy: Always\n", "path": "app-pod.yaml"}`
3. `kubectl apply -f app-pod.yaml 2>`

## Final report

Done. Both the stuck pod and the replacement are verified.

## Root cause

`default/app` was a **bare Pod** (no owner controller — `ownerReferences` empty) that had been deleted, but carried a custom finalizer:

```
deletionTimestamp: 2026-11-21T11:51:28Z
finalizers: ["example.com/hold"]     # no controller exists to remove it
deletionGracePeriodSeconds: 0
```

The API server can only finish deletion when every finalizer is cleared, and no component in this cluster implements `example.com/hold`, so the object stayed in Terminating indefinitely. Its only container had already exited (`state: terminated`, `exitCode: 137`, `phase: Failed`), so nothing was actually running to stop — the finalizer alone was the blocker.

## Change made (smallest justified)

1. Cleared the orphaned finalizer on the existing object so the API server could complete the already-requested deletion:
   `kubectl patch pod app --type=merge -p '{"metadata":{"finalizers":null}}'` → `pod/app patched`
2. Created the replacement `default/app` from the same image, keeping the `app: app` label, but without the `trap '' TERM; exec sleep 3600` command that made the old container ignore SIGTERM — so the new pod is normally manageable.

## Verification (fresh observations)

- Original symptom gone: pod `app` disappeared within 1s of the finalizer patch; `kubectl get pods -n default` reported "No resources found" before recreate, and there is no second/terminating `app` object.
- New pod healthy: `1/1 Running`, `phase=Running`, `Ready=True`, `restarts=0`, `deletionTimestamp=<none>`, `finalizers=<none>`.
- Intended image confirmed running: spec `nginx:1.31.5` and `runningImage=docker.io/library/nginx:1.31.5`, `imageID=...@sha256:05b8cb60c354a44ab824ea6e7dc69b46d50762cdbe728a347a5b656e6fb3d7c4`, container `state: running`.
- Functional check through the workload: `kubectl exec app -- curl -sS http://localhost/` returned `http_code=200`.
- Stability re-check 20s later: still `1/1 Running`, `Ready=True`, `restarts=0`, not terminating.

No other resources were altered — namespaces, node, taints, and kube-system workloads were left untouched.
