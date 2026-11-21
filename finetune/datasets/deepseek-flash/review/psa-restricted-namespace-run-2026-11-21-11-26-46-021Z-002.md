---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/psa-restricted-namespace/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# psa-restricted-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:

- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Write commands

1. `kubectl apply -f -`

## Final report

The workload now runs while the namespace's `restricted` PodSecurity enforcement remains in place.

## Diagnosis

- Namespace `restricted-app` enforces PodSecurity `restricted` (`pod-security.kubernetes.io/enforce: restricted`).
- Deployment `app` had been rolled forward to revision 2 (`pod-template-hash=787bf6bd59`) with `template.spec.securityContext: {}` and no container security context — i.e. the relaxed/privileged defaults.
- The ReplicaSet therefore failed to create pods, and its `ReplicaFailure` condition showed the admission rejection:
  `violates PodSecurity "restricted:latest": allowPrivilegeEscalation != false, unrestricted capabilities, runAsNonRoot != true, seccompProfile ...`
- Revision 1 (`app-dbd8dc899`) still carried a compliant security context, confirming the regression came from the new pod template, not from the namespace policy or the node.

## Change made

Restored the restricted-compliant security context on the **Deployment** `app` (the persistent owner of the pod template) — no change to the namespace labels, and no pod deletion or policy bypass:

- Pod level: `runAsNonRoot: true`, `runAsUser: 1001`, `seccompProfile.type: RuntimeDefault`
- Container `app`: `allowPrivilegeEscalation: false`, `capabilities.drop: ["ALL"]`

## Verification

- `kubectl rollout status` → `deployment "app" successfully rolled out`
- Deployment replica counts match the request: desired `1`, `updatedReplicas: 1`, `readyReplicas: 1`, `availableReplicas: 1`; condition `Available=True` (`MinimumReplicasAvailable`), `Progressing=True` (`NewReplicaSetAvailable`), `observedGeneration: 3`.
- Running pod `app-dbd8dc899-pbfj6`: `Running`, `ready=true`, image `docker.io/library/busybox:1.37.0` (the intended image), with the effective security context `runAsNonRoot:true, runAsUser:1001, seccompProfile:RuntimeDefault` and `allowPrivilegeEscalation:false, capabilities.drop:[ALL]`.
- `kubectl exec ... id` → `uid=1001`, confirming it actually runs as non-root.
- Namespace labels unchanged: `pod-security.kubernetes.io/enforce=restricted`, `audit=restricted`, `warn=restricted`.
- Latest events show the new pod created/started successfully; the only remaining `FailedCreate` warnings are stale, from the now-scaled-to-0 ReplicaSet `app-787bf6bd59`.
