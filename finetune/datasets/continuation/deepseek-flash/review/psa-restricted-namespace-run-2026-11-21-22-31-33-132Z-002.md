---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/psa-restricted-namespace/prompt/002.json
verdict: accepted
reason: 'Replaces the Deployment with a restricted-compliant security context, verified on the running pod with no further admission errors.'
---
# psa-restricted-namespace · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 2

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

1. `kubectl replace -f deployment_v2.yaml`

## Final report

The workload `app` in the `restricted-app` namespace is now running successfully.

### Root Cause
The namespace enforces the `restricted` Pod Security Admission (PSA) profile. The Deployment's Pod template was missing the required security context, causing the API server to reject Pod creation with `FailedCreate` (`violates PodSecurity "restricted:latest"`).

### Interventtion
I updated the Deployment's container `securityContext` to satisfy the restricted policy:
- `allowPrivilegeEscalation: false`
- `capabilities.drop: ["ALL"]`
- `runAsNonRoot: true`
- `runAsUser: 1000` (added because the `busybox` image defaults to root, which would otherwise cause a `CreateContainerConfigError`)
- `seccompProfile.type: RuntimeDefault`

The namespace's security settings were left completely untouched.

### Verified Outcome
- **Deployment status:** `1` desired, `1` ready, `1` up-to-date, and `1` available.
- **Pod status:** `app-7cb4d5f976-5ksf8` is `Running` with `1/1` containers ready.
- **Applied Security Context:** The running pod's container confirms `allowPrivilegeEscalation=false`, `capabilities.drop=["ALL"]`, `runAsNonRoot=true`, `runAsUser=1000`, and `seccompProfile.type=RuntimeDefault`, matching the required `restricted` profile. Admission errors are gone from the events.
