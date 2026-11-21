---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/psa-restricted-namespace/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# psa-restricted-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:

- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Write commands

1. `kubectl patch deploy app -n restricted-app --type=strategic -p '{`

## Final report

Resolved. The deployment `restricted-app/app` was being rejected by the namespace's Pod Security Admission (`pod-security.kubernetes.io/enforce: restricted`) because its pod template had an empty `securityContext`. The namespace labels were not changed.

**Change made** (single patch to Deployment `restricted-app/app`):

- Pod-level: `runAsNonRoot: true`, `runAsUser: 1000`, `seccompProfile.type: RuntimeDefault`
- Container `app`: `allowPrivilegeEscalation: false`, `capabilities.drop: ["ALL"]`

This addresses all four violations reported in the `FailedCreate` events (allowPrivilegeEscalation, unrestricted capabilities, runAsNonRoot, seccompProfile) while leaving the namespace's `restricted` enforce/audit/warn settings intact.

**Verified outcome:**

- New ReplicaSet pod admitted and scheduled: `app-56cccfd9d4-qfnmx` — phase `Running`, container `ready=true`, image `busybox:1.37.0` (intended image).
- `kubectl rollout status` → `deployment "app" successfully rolled out`.
- Replica counts match: `desired=1 updated=1 ready=1 available=1`, no unavailable replicas.
- Deployment conditions: `Available=True (MinimumReplicasAvailable)`, `Progressing=True (NewReplicaSetAvailable)`.
- Events after the patch show `SuccessfulCreate`, `Pulled`, `Created`, `Started` with no new admission failures (the last `FailedCreate` predates the patch, against the old ReplicaSet `app-787bf6bd59`).
- `kubectl exec ... id` confirms the container runs as non-root: `uid=1000`.
- Namespace labels remain `pod-security.kubernetes.io/{enforce,audit,warn}: restricted`.
