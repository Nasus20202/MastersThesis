---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/node-selector-no-match/prompt/006.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: kubectl patch removed nodeSelector disktype ssd (msg 28); node labels lack it (msg 17); rollout and curl verified (msgs 71, 67)."
---

# node-selector-no-match · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 6

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/scheduling-eviction/assign-pod-node.md: nodeSelector

## Write commands

1. `kubectl delete pod app-77bd4fcf9c-c8dwl 2>`
2. `kubectl patch deployment app --type=json -p='[{"op": "remove", "path": "/spec/template/spec/nodeSelector"}]' 2>`

## Final report

The Kubernetes application rollout issue has been resolved. Here's what was found and fixed:

## Root Cause

The deployment had `nodeSelector: disktype: ssd` in its pod template, but no nodes in the cluster had this label. This caused one pod (app-77bd4fcf9c-c8dwl) to remain in **Pending** state with the error: "0/1 nodes are available: 1 node(s) didn't match Pod's node affinity/selector".

## Fix Applied

Removed the `nodeSelector` from the deployment template using:

```bash
kubectl patch deployment app --type=json -p='[{"op": "remove", "path": "/spec/template/spec/nodeSelector"}]'
```

Also deleted the stuck pending pod to allow the deployment to recover.

## Verification - Observed Changes and Results

**Deployment Status:**

- Replicas: 3 desired | 3 updated | 3 total | **3 available | 0 unavailable**
- READY: 3/3, UP-TO-DATE: 3, AVAILABLE: 3
- Rollout status: "deployment 'app' successfully rolled out"

**Pod Status:**
All 3 pods are now Running and Ready:

| Pod Name             | READY | STATUS  | RESTARTS |
| -------------------- | ----- | ------- | -------- |
| app-575b7496bc-g5kk8 | 1/1   | Running | 0        |
| app-575b7496bc-pn8x6 | 1/1   | Running | 0        |
| app-575b7496bc-xh9ml | 1/1   | Running | 0        |

**Application Health Confirmed:**
Executed `curl http://localhost` inside the container pod and received nginx welcome page response, confirming pods are serving traffic.

The application rollout is now complete and healthy.
