---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/psa-restricted-namespace/prompt/003.json
verdict: accepted
reason: 'Patches pod and container security contexts, verified by a Ready pod, matching security fields, and unchanged namespace labels.'
---
# psa-restricted-namespace · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 3

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

1. `kubectl patch deployment app -n restricted-app --type=strategic -p '{"spec":{"template":{"spec":{"securityContext":{"runAsUser":1000,"runAsGroup":1000,"fsGroup":1000,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"app","securityContext":{"allowPrivilegeEscalation":false,"runAsNonRoot":true,"capabilities":{"drop":["ALL"]}}}]}}}}'`

## Final report

The issue was that the Deployment `app` in namespace `restricted-app` was being rejected by the namespace's Pod Security Admission policy (`pod-security.kubernetes.io/enforce=restricted`). The container specification was missing the required security context fields.

### Diagnosis
The ReplicaSet events showed clear admission denials:
```
Error creating: pods "app-787bf6bd59-..." is forbidden: violates PodSecurity "restricted:latest": 
- allowPrivilegeEscalation != false 
- unrestricted capabilities 
- runAsNonRoot != true 
- seccompProfile
```

### Fix
I patched the Deployment to satisfy the `restricted` profile while keeping the namespace's security labels intact. The following fields were added:

**Pod-level `securityContext`:**
- `runAsUser: 1000`
- `runAsGroup: 1000`
- `fsGroup: 1000`
- `seccompProfile.type: RuntimeDefault`

**Container-level `securityContext`:**
- `allowPrivilegeEscalation: false`
- `runAsNonRoot: true`
- `capabilities.drop: ["ALL"]`

### Verification
- **Deployment status:** `1/1 READY`, `1 UP-TO-DATE`, `1 AVAILABLE`
- **Pod status:** `app-5f69bcb6b4-pdg2l` is `Running` and `1/1 Ready`
- **Image:** running the intended `busybox:1.37.0`
- **Pod security context:** confirmed to match the patched requirements (`allowPrivilegeEscalation=false`, `runAsNonRoot=true`, `capabilities.drop=["ALL"]`, `seccompProfile=RuntimeDefault`, `runAsUser=1000`)
- **Namespace labels:** `pod-security.kubernetes.io/enforce=restricted` remains in place
- **Events:** no further `FailedCreate` / admission denial events in the namespace
