---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/psa-restricted-namespace/prompt/001.json
verdict: accepted
reason: 'Adds the restricted-profile security context including runAsUser 1000, verified by a 1/1 Ready pod and unchanged namespace labels.'
---
# psa-restricted-namespace · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 1

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:
- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Inherited stop

The first 21 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 0
stdout:
NAME                  READY   STATUS                       RESTARTS   AGE
app-5dcbf687b-lqbn8   0/1     CreateContainerConfigError   0          72s
stderr:

Its writes:

1. `kubectl patch deployment app -n restricted-app --patch '{"spec": {"template": {"spec": {"containers": [{"securityContext": {"allowPrivilegeEscalation": false, "capabilities": {"drop": ["ALL"]}, "runAsNonRoot": true, "seccompProfile": {"type": "RuntimeDefault"}}}]}}}}'`
2. `kubectl apply -f deployment.yaml -n restricted-app --dry-run=client`
3. `kubectl apply -f deployment.yaml -n restricted-app`

## Teacher's write commands

1. `kubectl patch deployment app -n restricted-app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]},"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}}}]}}}}'`
2. `kubectl patch deployment app -n restricted-app --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","securityContext":{"runAsUser":1000}}]}}}}'`

## Final report

The workload is now running while the namespace's strict security settings remain intact. Here is a summary of the investigation and the fix:

### Diagnosis
1. **Identified the workload:** The `restricted-app` namespace contained a Deployment named `app` with `0/1` ready replicas.
2. **Checked events:** `kubectl get events` revealed that the Deployment's ReplicaSet (`app-787bf6bd59`) was failing to create Pods with the error:
   > `Error creating: pods ... is forbidden: violates PodSecurity "restricted:latest": allowPrivilegeEscalation != false ... unrestricted capabilities ... runAsNonRoot != true ... seccompProfile ...`
   This confirmed that the Pods were not being admitted because they did not comply with the namespace's `restricted` Pod Security Admission profile.
3. **Secondary issue:** Once the initial security contexts were applied, the Pod entered `CreateContainerConfigError` because `busybox:1.37.0` defaults to running as `root`, conflicting with the required `runAsNonRoot: true` setting.

### Resolution
I patched the existing `app` Deployment to add the required security context to the container. I did **not** modify the namespace's security settings.
The container's `securityContext` was updated to:
```json
{
  "allowPrivilegeEscalation": false,
  "capabilities": {"drop": ["ALL"]},
  "runAsNonRoot": true,
  "runAsUser": 1000,
  "seccompProfile": {"type": "RuntimeDefault"}
}
```

### Verification
- **Deployment Status:** `1 desired | 1 updated | 1 total | 1 available` (rollout completed successfully).
- **Pod Status:** `app-7cb4d5f976-2d7lv` is `1/1 Running`.
- **Security Settings Preserved:** The namespace `restricted-app` still has the exact original labels:
  `pod-security.kubernetes.io/audit=restricted`, `pod-security.kubernetes.io/enforce=restricted`, `pod-security.kubernetes.io/warn=restricted`.
