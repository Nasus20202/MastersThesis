---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/deprecated-api-group/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# deprecated-api-group · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A team handed over the manifest below, but the cluster rejects it because it targets an API version that is no longer served. Rewrite it as an equivalent resource for the API version this cluster serves and apply it into the namespace legacy-app, keeping the same name, replicas, labels and image.

apiVersion: apps/v1beta1
kind: Deployment
metadata:
name: app
labels:
app: app
spec:
replicas: 3
selector:
matchLabels:
app: app
template:
metadata:
labels:
app: app
spec:
containers: - name: app
image: nginx:1.31.5
ports: - name: http
containerPort: 80

Documented in:

- content/en/docs/reference/using-api/deprecation-guide.md: Removed APIs by release > v1.16 > Deployment, What to do > Migrate to non-deprecated APIs

## Write commands

1. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n name: app\n namespace: legacy-app\n labels:\n app: app\nspec:\n replicas: 3\n selector:\n matchLabels:\n app: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - name: app\n image: nginx:1.31.5\n ports:\n - name: http\n containerPort: 80\n", "path": "deployment-app.yaml"}`
2. `kubectl apply -f deployment-app.yaml`

## Final report

The manifest has been rewritten and applied successfully. Here's what was done and verified:

**Change made:** Updated `apiVersion` from `apps/v1beta1` to `apps/v1`, which is the current supported version in this cluster.

**Verification results:**

| Check             | Expected     | Observed                                         |
| ----------------- | ------------ | ------------------------------------------------ |
| API version       | apps/v1      | apps/v1 (confirmed via `kubectl get deployment`) |
| Namespace         | legacy-app   | legacy-app                                       |
| Name              | app          | app                                              |
| Replicas          | 3            | 3/3 READY, 3 UP-TO-DATE, 3 AVAILABLE             |
| Labels (metadata) | app: app     | app                                              |
| Labels (template) | app: app     | app                                              |
| Image             | nginx:1.31.5 | nginx:1.31.5                                     |
| Pod status        | Running      | All 3 pods Running                               |

The deployment `app` is now running with 3 ready/available replicas using the correct image in the `legacy-app` namespace.
