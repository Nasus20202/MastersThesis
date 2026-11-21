---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/deprecated-api-group/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# deprecated-api-group · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 5

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

1. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n name: app\n labels:\n app: app\nspec:\n replicas: 3\n selector:\n matchLabels:\n app: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - name: app\n image: nginx:1.31.5\n ports:\n - name: http\n containerPort: 80\n", "path": "legacy-app/deployment.yaml"}`
2. `kubectl apply -f legacy-app/deployment.yaml -n legacy-app 2>`

## Final report

The manifest has been successfully rewritten and applied. Here's what was done:

**Change Made:**

- Changed `apiVersion` from `apps/v1beta1` to `apps/v1` (the current served API version)

**Verification of Outcome:**

- Deployment created: ✓ `deployment.apps/app created`
- Replica count: 3 requested, 3 ready, 3 available ✓
- Namespace: legacy-app ✓
- Name: app ✓
- Labels: `app: app` (in metadata and template) ✓
- Image: nginx:1.31.5 ✓

**Status Confirmation:**
From `kubectl get deployment app -n legacy-app`:

```
NAME   READY   UP-TO-DATE   AVAILABLE   AGE   CONTAINERS   IMAGES         SELECTOR
app    3/3     3            3           22s   app          nginx:1.31.5   app=app
```

The deployment is fully operational with all replicas ready and available, matching the originally requested replica count of 3.
