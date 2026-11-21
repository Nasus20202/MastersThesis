---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/deprecated-api-group/prompt/002.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: Report quotes error for name "app-old", not in the task manifest (name app); orig.yaml not shown. | second pass against the raw record: ok: Manifest rewritten only to apps/v1 (msg 12 shows same name, labels, image); msgs 10, 16, 21 show 3/3 and unchanged re-apply.'
---

# deprecated-api-group · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

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

1. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n name: app\n labels:\n app: app\nspec:\n replicas: 3\n selector:\n matchLabels:\n app: app\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - name: app\n image: nginx:1.31.5\n ports:\n - name: http\n containerPort: 80\n", "path": "/tmp/deployment.yaml"}`
2. `kubectl apply -f deployment.yaml -n legacy-app 2>`
3. `kubectl apply -f orig.yaml -n legacy-app --dry-run=server 2>`
4. `kubectl apply -f deployment.yaml -n legacy-app 2>`

## Final report

## Result

The original manifest failed because this cluster (v1.37.0) no longer serves `apps/v1beta1`. Verified the symptom directly:

```
error: resource mapping not found for name: "app-old" ...: no matches for kind "Deployment" in version "apps/v1beta1"
error: the server doesn't have a resource type "deployments" in group "apps.v1beta1"
```

`kubectl api-resources --api-group=apps` lists only `apps/v1` for Deployment, so the equivalent resource is the same manifest with `apiVersion: apps/v1`. No other field needed changing: `apps/v1` requires `spec.selector` and matching `spec.template.metadata.labels`, and the handed-over manifest already had both (`app: app`).

### Manifest applied (`/tmp/deployment.yaml`)

Only the `apiVersion` line differs from the original: `apps/v1`.

```
deployment.apps/app created
```

### Verification against the live object and pods

| Check                        | Observed                                                                                         |
| ---------------------------- | ------------------------------------------------------------------------------------------------ |
| Stored API version / kind    | `apps/v1`, `Deployment`                                                                          |
| Name / namespace             | `app` in `legacy-app`                                                                            |
| `metadata.labels`            | `{"app":"app"}`                                                                                  |
| Requested vs actual replicas | requested `3`; updated `3`, ready `3`, available `3`, unavailable none                           |
| Selector / pod labels        | `{"app":"app"}` / `{"app":"app"}`                                                                |
| Container name / image       | `app` / `nginx:1.31.5`                                                                           |
| Container port               | `http` = 80, TCP                                                                                 |
| Conditions                   | `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`         |
| Pods                         | 3/3 `Running`, `1/1` ready, 0 restarts                                                           |
| Running image identity       | all three pods: `docker.io/library/nginx@sha256:05b8cb60...b3d7c4` pulled for tag `nginx:1.31.5` |
| Functional check             | `curl` inside a pod to port 80 returned `HTTP 200`                                               |

Re-applying the file returned `deployment.apps/app unchanged` and the counts held at `READY 3/3, UP-TO-DATE 3, AVAILABLE 3` after a further wait, so the state is stable and idempotent. Name, replicas, labels and image are all preserved as originally specified.
