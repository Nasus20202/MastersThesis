---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/deprecated-api-group/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# deprecated-api-group · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

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

1. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n name: app\n labels:\n app: app\nspec:\n replicas: 3\n selector:\n matchLabels:\n app: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - name: app\n image: nginx:1.31.5\n ports:\n - name: http\n containerPort: 80\n", "path": "/tmp/deployment-app.yaml"}`
2. `kubectl -n legacy-app apply -f deployment-app.yaml --dry-run=server 2>`
3. `kubectl -n legacy-app apply -f deployment-app.yaml 2>`
4. `kubectl -n legacy-app apply -f orig.yaml --dry-run=server 2>`

## Final report

Done — the manifest was ported to `apps/v1` and applied successfully.

## Diagnosis

- Cluster is **v1.37.0**; `kubectl api-resources --api-group=apps` and `kubectl api-versions` show **only `apps/v1`** is served. `kubectl get --raw /apis/apps/v1beta1` → `NotFound`, so the original manifest is rejected: `no matches for kind "Deployment" in version "apps/v1beta1"` (reproduced above).
- Namespace `legacy-app` already existed and was empty (only `kube-root-ca.crt`), so nothing had to be migrated or removed.

## Change made

Only one field changed — `apiVersion: apps/v1beta1` → `apiVersion: apps/v1`. Everything else is byte-identical, which is valid because the v1beta1 manifest already carried the `spec.selector.matchLabels` field that `apps/v1` requires (the sole API-level incompatibility, along with the version string).

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels: { app: app }
spec:
  replicas: 3
  selector:
    matchLabels: { app: app }
  template:
    metadata:
      labels: { app: app }
    spec:
      containers:
        - name: app
          image: nginx:1.31.5
          ports: [{ name: http, containerPort: 80 }]
```

Applied with a server-side dry run first (`created (server dry run)`), then `kubectl -n legacy-app apply -f deployment-app.yaml` → `deployment.apps/app created`.

## Verification (fresh observations)

- Served version / identity: `apiVersion=apps/v1`, name `app`, labels `{"app":"app"}`, selector `{"app":"app"}`, pod template labels `{"app":"app"}`, image `nginx:1.31.5`, container port `http/80` — all preserved.
- Controller convergence requested vs. actual: `replicas=3 ready=3 updated=3 available=3`; `kubectl get deploy` shows `READY 3/3, UP-TO-DATE 3, AVAILABLE 3`; conditions `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- Pods: 3/3 `Running` and `READY=true`, each running `nginx:1.31.5`, owned via ReplicaSet `app-575b7496bc` (`owner=Deployment/app`, `ready=3`). Events confirm `Pulled` + `Started` for all three — the image reference resolves, so no image substitution was needed.

No other resources were touched and no constraints (name, replicas, labels, image) were altered.
