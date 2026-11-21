---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/psa-restricted-namespace/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# psa-restricted-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:

- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Write commands

1. `kubectl patch deployment app -n restricted-app --type=strategic -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"app","securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}}]}}}}'`

## Final report

Resolved. The Deployment's Pod template didn't meet the namespace's enforced `restricted` Pod Security Standard, so every Pod creation was rejected by the PodSecurity admission controller (`FailedCreate: violates PodSecurity "restricted:latest"`).

**Change made** — patched only the `app` Deployment's Pod template security context:

- Pod level: `runAsNonRoot: true`, `runAsUser: 1000`, `seccompProfile.type: RuntimeDefault`
- Container `app`: `allowPrivilegeEscalation: false`, `capabilities.drop: ["ALL"]`

The namespace itself was left untouched; its labels still read `pod-security.kubernetes.io/enforce/audit/warn: restricted`.

**Verified outcome**

- Deployment: desired=1, updated=1, ready=1, available=1; `kubectl rollout status` reported "successfully rolled out".
- Pod `app-56cccfd9d4-nd25t`: `Running`, `1/1 Ready`; new RS scaled to 1, old violating RS scaled to 0; no new `FailedCreate` events after the patch.
- Runtime enforcement confirmed: `id` → `uid=1000` (non-root), `CapEff: 0000000000000000` (all capabilities dropped), `Seccomp: 2` (RuntimeDefault filter), and the container is executing its `sleep` command normally.
