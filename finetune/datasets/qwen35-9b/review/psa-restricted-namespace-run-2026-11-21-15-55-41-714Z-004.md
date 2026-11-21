---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/psa-restricted-namespace/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# psa-restricted-namespace · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 4

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:

- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Write commands

1. `kubectl patch deployment app -n restricted-app --type=json -p='[{"op": "replace", "path": "/spec/template/spec/securityContext", "value": {"runAsNonRoot": true, "seccompProfile": {"type": "RuntimeDefault"}}}, {"op": "replace", "path": "/spec/template/spec/containers/0/securityContext", "value": {"allowPrivilegeEscalation": false, "capabilities": {"drop": ["ALL"]}}}]' 2>`
2. `kubectl patch deployment app -n restricted-app --type=json -p='[{"op": "replace", "path": "/spec/template/spec/securityContext", "value": {"runAsNonRoot": true, "seccompProfile": {"type": "RuntimeDefault"}, "runAsUser": 1000}}, {"op": "replace", "path": "/spec/template/spec/containers/0/securityContext", "value": {"allowPrivilegeEscalation": false, "capabilities": {"drop": ["ALL"]}}}]' 2>`

## Final report

## Summary

**Root Cause:** The Deployment `app` in namespace `restricted-app` was missing required PodSecurity settings, causing pods to be rejected by the namespace's "restricted:latest" PodSecurity admission policy.

**Fix Applied:** Updated the deployment's security context with PodSecurity-compliant settings via `kubectl patch`:

- **Container securityContext:**
  - `allowPrivilegeEscalation: false`
  - `capabilities.drop: [ALL]`
- **Pod securityContext:**
  - `runAsNonRoot: true`
  - `runAsUser: 1000`
  - `seccompProfile.type: RuntimeDefault`

**Verification:**

- Pod status: `app-56cccfd9d4-jl6zc` is **Running** with `READY 1/1`
- All pod conditions: `Initialized=True`, `Ready=True`, `ContainersReady=True`, `PodReadyToStartContainers=True`
- Deployment: `Available=1`, `Progressing=True`, `ReadyReplicas=1`

The workload now runs successfully while maintaining the namespace's security settings in place.
