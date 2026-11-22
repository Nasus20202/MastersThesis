---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-20-57-03-137Z/psa-restricted-namespace/prompt/003.json
verdict: accepted
reason: 'Applies a restricted-compliant security context and leaves the namespace labels alone, as reported.'
---
# psa-restricted-namespace · deepseek-flash · run-2026-11-21-20-57-03-137Z · attempt 3

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:
- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Inherited stop

The first 23 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 1
stdout:

stderr:
Error from server (BadRequest): invalid JSON patch

Its writes:

1. `edit_file {"new_string": " spec:\n containers:\n - args:\n - exec sleep 3600\n command:\n - sh\n - -c\n image: busybox:1.37.0\n imagePullPolicy: IfNotPresent\n name: app\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n dnsPolicy: ClusterFirst\n restartPolicy: Always\n schedulerName: default-scheduler\n securityContext:\n runAsNonRoot: true\n allowPrivilegeEscalation: false\n capabilities:\n drop:\n - ALL\n seccompProfile:\n type: RuntimeDefault\n terminationGracePeriodSeconds: 30\n status:\n conditions:\n - lastTransitionTime: \"2026-10-10T00:25:54Z\"\n lastUpdateTime: \"2026-10-10T00:25:54Z\"\n message: Deployment does not have minimum availability.\n reason: MinimumReplicasUnavailable\n status: \"False\"\n type: Available\n - lastTransitionTime: \"2026-10-10T00:25:41Z\"\n lastUpdateTime: \"2026-10-10T00:26:26Z\"\n message: Created new replica set \"app-787bf6bd59\"\n reason: NewReplicaSetCreated\n status: \"True\"\n type: Progressing\n - lastTransitionTime: \"2026-10-10T00:26:26Z\"\n lastUpdateTime: \"2026-10-10T00:26:26Z\"\n message: 'pods \"app-787bf6bd59-s8zzv\" is NOT forbidden: all constraints passed'\n reason: FailedCreate\n status: \"True\"\n type: ReplicaFailure\n observedGeneration: 2\n terminatingReplicas: 0\n unavailableReplicas: 1\n reason: FailedCreate", "old_string": " securityContext: {}", "path": "app-deployment.yaml"}`
2. `kubectl patch deployment app -n restricted-app --type='merge' -p '{"spec": {"template": {"spec": {"containers": [{"name": "app", "securityContext": {"runAsNonRoot": true, "allowPrivilegeEscalation": false, "capabilities": {"drop": ["ALL"]}, "seccompProfile": {"type": "RuntimeDefault"}}}}}}}}'`
3. `edit_file {"content": " securityContext:\n runAsNonRoot: true\n allowPrivilegeEscalation: false\n capabilities:\n drop:\n - ALL\n seccompProfile:\n type: RuntimeDefault", "old_string": " securityContext: {}", "path": "deployment.yaml"}`
4. `edit_file {"content": " securityContext:\n runAsNonRoot: true\n allowPrivilegeEscalation: false\n capabilities:\n drop:\n - ALL\n seccompProfile:\n type: RuntimeDefault", "old_string": " securityContext: {}", "path": "deployment.yaml"}`
5. `kubectl patch deployment app -n restricted-app --type='merge' -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","securityContext":{"runAsNonRoot":true,"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"seccompProfile":{"type":"RuntimeDefault"}}}}}}}}'`
6. `kubectl patch deployment app -n restricted-app --type='merge' -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","securityContext":{"runAsNonRoot":true,"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"seccompProfile":{"type":"RuntimeDefault"}}}}}}}}'`

## Teacher's write commands

1. `kubectl apply -f deployment_fixed.yaml`
2. `kubectl apply -f deployment_v2.yaml`

## Final report

The workload is now running and the namespace security settings are fully intact.

## Root cause

The `restricted-app` namespace enforces **Pod Security Admission** at the `restricted` level (labels: `pod-security.kubernetes.io/enforce=restricted`, `audit=restricted`, `warn=restricted`). The `app` Deployment's Pod template had an empty `securityContext: {}`, so every Pod it created was rejected by the admission controller with:

```
pods "..." is forbidden: violates PodSecurity "restricted:latest":
  allowPrivilegeEscalation != false,
  unrestricted capabilities,
  runAsNonRoot != true,
  seccompProfile not set to RuntimeDefault/Localhost
```

## Fix applied (smallest justified change)

I kept the namespace's PSA settings untouched and instead made the Pod template satisfy them, by adding a **container-level** `securityContext` to the Deployment:

```yaml
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  runAsUser: 1000
  runAsGroup: 1000
  runAsNonRoot: true
  seccompProfile:
    type: RuntimeDefault
```

A second iteration was needed: after admission passed, the kubelet reported `container has runAsNonRoot and image will run as root` because busybox defaults to UID 0. Adding `runAsUser: 1000` / `runAsGroup: 1000` resolved that while keeping the non-root requirement satisfied.

## Verified outcome

- **Namespace security settings preserved:** labels still show `enforce=restricted`, `audit=restricted`, `warn=restricted`.
- **Workload running:** Deployment `app` reports `1 desired, 1 updated, 1 ready, 1 available` — all counts match, with `Available=True (MinimumReplicasAvailable)` and `Progressing=True (NewReplicaSetAvailable)`, and no `ReplicaFailure` condition.
- **Pod admitted and healthy:** Pod `app-77d85df4f8-8qxw5` is `1/1 Running`; events show normal `Scheduled → Pulled → Created → Started` with no PSA rejection.
- **Actually non-root and executing:** `id` inside the container returns `uid=1000 gid=1000`, and the process table shows `1 1000 sleep 3600`, confirming the workload's command is live.
