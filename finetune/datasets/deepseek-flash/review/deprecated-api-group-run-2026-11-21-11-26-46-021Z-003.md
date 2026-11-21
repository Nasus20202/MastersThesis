---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/deprecated-api-group/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# deprecated-api-group · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

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
2. `kubectl apply -n legacy-app -f /tmp/deployment-app.yaml 2>`
3. `kubectl apply -n legacy-app -f - 2>`

## Final report

## Diagnosis

- Cluster server version is **v1.37.0**, and `kubectl api-versions` lists `apps/v1` — **`apps/v1beta1` is not served**, which is exactly why the handed-over manifest was rejected (`no matches for kind "Deployment" in version "apps/v1beta1"`, confirmed by a direct probe apply).
- The namespace `legacy-app` already existed and was empty, with no ResourceQuota/LimitRange or other constraints in the way.
- The only migration need was the API version: in `apps/v1`, the `selector` field became required rather than defaulted from pod template labels. The given manifest already specified `matchLabels: {app: app}` matching the template labels, so the translation is one-to-one with no semantic change.

## Change applied

Rewrote the manifest with `apiVersion: apps/v1` (same kind/name/labels/replicas/selector/pod template/image/port) and applied it into `legacy-app`. Nothing else on the cluster was modified.

## Verification (observed values)

| Check                                           | Result                                                                                                                                                                                                          |
| ----------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| API accepted                                    | `deployment.apps/app created`; live object reports `apiVersion=apps/v1`                                                                                                                                         |
| Name                                            | `app`, in namespace `legacy-app`                                                                                                                                                                                |
| Replicas (requested vs updated/ready/available) | `replicas=3`, `3/3 READY`, `UP-TO-DATE 3`, `AVAILABLE 3` — match                                                                                                                                                |
| Labels                                          | Deployment `metadata.labels={"app":"app"}`, pod template `{"app":"app"}`, selector `matchLabels={"app":"app"}`; all 3 pods carry `app=app`                                                                      |
| Image                                           | Deployment `image=nginx:1.31.5`; all 3 running pods `nginx:1.31.5` (`imageID .../nginx@sha256:05b8cb60...`, single consistent digest)                                                                           |
| Port                                            | `containerPort=80, name=http, protocol=TCP`                                                                                                                                                                     |
| Controller health                               | `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`, `rollout status`: "successfully rolled out"; RS `app-575b7496bc` owned by `Deployment/app` with `availableReplicas=3` |
| Functional check                                | In-pod request to `127.0.0.1:80` returned `HTTP/1.1 200 OK` and `Server: nginx/1.31.5`                                                                                                                          |
| Old API still unserved (symptom source)         | apply of `apps/v1beta1` still rejected — the new object works because it uses a served version, not because the cluster changed                                                                                 |

Final fresh observation after convergence: `app 3/3 3 3`, all three pods `READY=true` with image `nginx:1.31.5`.
