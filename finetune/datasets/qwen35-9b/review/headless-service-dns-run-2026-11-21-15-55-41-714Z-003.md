---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/headless-service-dns/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# headless-service-dns · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 3

## Task

A StatefulSet named app runs two replicas, but the headless Service that provides their stable per-Pod DNS names is missing. Create a headless Service named app in the default namespace with clusterIP None, selector app=app and port 80.

Documented in:

- content/en/docs/concepts/services-networking/service.md: Headless Services

## Write commands

1. `kubectl create service headless app -n default --selector="app=app" --port=80`
2. `kubectl create service headless app -n default --type ClusterIP=None`
3. `kubectl create service --help`
4. `kubectl create service headless --help`
5. `kubectl apply -f - <<'EOF'`

## Final report

The headless Service has been successfully created with all required specifications:

**Verification:**

- **Name**: `app` ✓
- **Namespace**: `default` ✓
- **Type**: Headless (clusterIP: `None`) ✓
- **Selector**: `app=app` ✓
- **Port**: `80/TCP` ✓

The Service configuration is confirmed via `kubectl get service app -n default`:

```
NAME   TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE   SELECTOR
app    ClusterIP   None         <none>        80/TCP    7s    app=app
```

The StatefulSet `app` with 2 replicas now has a proper headless Service providing stable per-Pod DNS names (e.g., `app-0.app.default.svc.cluster.local`, `app-1.app.default.svc.cluster.local`).
