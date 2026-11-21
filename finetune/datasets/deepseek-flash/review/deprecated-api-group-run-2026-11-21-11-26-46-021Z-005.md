---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/deprecated-api-group/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# deprecated-api-group · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

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

1. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n name: app\n labels:\n app: app\nspec:\n replicas: 3\n selector:\n matchLabels:\n app: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - name: app\n image: nginx:1.31.5\n ports:\n - name: http\n containerPort: 80\n", "path": "/tmp/app-deployment.yaml"}`
2. `kubectl apply -f app-deployment.yaml -n legacy-app 2>`

## Final report

Done.

## What the cluster serves

`kubectl api-versions` shows the only serving `apps` group version is `apps/v1` — `apps/v1beta1` is not served (server v1.37.0), which is exactly why the handed-over manifest was rejected.

## Change made

The only change was the API version and the fields required by the newer schema:

| Field           | Before (`apps/v1beta1`)     | After (`apps/v1`)                                                      |
| --------------- | --------------------------- | ---------------------------------------------------------------------- |
| `apiVersion`    | `apps/v1beta1` (not served) | `apps/v1` (served)                                                     |
| `spec.selector` | optional                    | `matchLabels: app: app` — **required**, carried over from the original |

Everything else was kept byte-for-byte: name `app`, replicas `3`, labels `app: app` on the Deployment and the pod template, container name `app`, image `nginx:1.31.5`, and port `http`/80. Applied with `kubectl apply -f app-deployment.yaml -n legacy-app` → `deployment.apps/app created` (manifest also written to `/tmp/app-deployment.yaml`).

## Verified outcome (fresh observations)

- **Requested vs. actual replicas match:** `replicas=3 ready=3 updated=3 available=3` — the Deployment reports `3/3 READY`, `3 UP-TO-DATE`, `3 AVAILABLE`.
- **Conditions:** `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)` — no `ReplicaFailure`.
- **Image is the intended one:** all 3 pods report `image=docker.io/library/nginx:1.31.5`, `ready=true`, `imageID=docker.io/library/nginx@sha256:05b8cb60...d7c4`, via a single ReplicaSet `app-575b7496bc` (3/3 ready).
- **Preserved identity:** stored object is `apps/v1`, `metadata.name=app`, `metadata.labels={"app":"app"}`, `spec.selector.matchLabels={"app":"app"}`, pod labels `{"app":"app"}`, container image `nginx:1.31.5`, port `[{"containerPort":80,"name":"http","protocol":"TCP"}]`.
- **Functionally live:** in-pod `curl http://127.0.0.1:80/` returned `http_code=200`, so the declared container port really serves.
- **Symptom gone:** no `apps/v1beta1` object exists; the only Deployment in `legacy-app` is served and stored as `apps/v1`. No taints, restarts, or image-pull errors in the events (all three pods show `Pulled`/`Started`).

I made no changes beyond the API version and the now-required selector — no unrelated settings were touched, and nothing was deleted or restarted to fake health.
