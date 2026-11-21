---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/deprecated-api-group/prompt/003.json
verdict: accepted
reason: "first review of a replacement attempt against the raw record by a Haiku model: ok: kubectl apply -f deployment.yaml -n legacy-app; apps/v1 Deployment 3/3 Ready, name, labels and image match."
---

# deprecated-api-group · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 3

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

1. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n name: app\n labels:\n app: app\nspec:\n replicas: 3\n selector:\n matchLabels:\n app: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - name: app\n image: nginx:1.31.5\n ports:\n - name: http\n containerPort: 80\n", "path": "deployment.yaml"}`
2. `kubectl apply -f deployment.yaml -n legacy-app 2>`

## Final report

The manifest has been rewritten from `apps/v1beta1` to `apps/v1` and applied successfully.

**Change made:**

- Changed `apiVersion` from `apps/v1beta1` to `apps/v1`

**Verification:**

| Check              | Expected      | Observed                                                      |
| ------------------ | ------------- | ------------------------------------------------------------- |
| Deployment created | success       | ✓ "deployment.apps/app created"                               |
| Namespace          | legacy-app    | ✓                                                             |
| Name               | app           | ✓                                                             |
| Replicas           | 3             | ✓ (spec.replicas=3, readyReplicas=3, availableReplicas=3)     |
| Labels             | app: app      | ✓ (metadata.labels.app=app, template.metadata.labels.app=app) |
| Image              | nginx:1.31.5  | ✓                                                             |
| Pod status         | Running/Ready | ✓ (all 3 pods show STATUS=Running, READY=1/1)                 |

The cluster now accepts the manifest and all 3 replicas are running with the expected configuration.
