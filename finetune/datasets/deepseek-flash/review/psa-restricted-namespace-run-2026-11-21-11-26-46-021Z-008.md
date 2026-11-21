---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/psa-restricted-namespace/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# psa-restricted-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:

- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Write commands

1. `kubectl patch deployment app -n restricted-app --type=merge -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"seccompProfile":{"type":"RuntimeDefault"}},"containers":[{"name":"app","image":"busybox:1.37.0","command":["sh","-c"],"args":["exec sleep 3600"],"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}}]}}}}'`
2. `kubectl patch deployment app -n restricted-app --type=merge -p '{"spec":{"template":{"spec":{"securityContext":{"runAsNonRoot":true,"runAsUser":1000,"runAsGroup":1000,"seccompProfile":{"type":"RuntimeDefault"}}}}}}'`

## Final report

Fixed. The Deployment's Pod template didn't satisfy the namespace's enforced `restricted` Pod Security Admission profile, so the ReplicaSet's Pod creation requests were rejected.

**Evidence of the original cause** (from the ReplicaSet FailedCreate events):

```
violates PodSecurity "restricted:latest":
 allowPrivilegeEscalation != false
 unrestricted capabilities (must drop ["ALL"])
 runAsNonRoot != true
 seccompProfile (must be RuntimeDefault or Localhost)
```

**Change made** — patched only the `restricted-app/app` Deployment Pod template (namespace labels untouched, enforcement kept at `restricted`):

- Pod `securityContext`: `runAsNonRoot: true`, `runAsUser: 1000`, `runAsGroup: 1000`, `seccompProfile.type: RuntimeDefault`
- Container `app` `securityContext`: `allowPrivilegeEscalation: false`, `capabilities.drop: ["ALL"]`

An intermediate attempt was admitted by PSA but hit `container has runAsNonRoot and image will run as root`; adding the non-root UID resolved that.

**Verified outcome (fresh observations):**

- `kubectl rollout status` → `deployment "app" successfully rolled out`
- Deployment: `replicas=1 updated=1 ready=1 available=1` (requested count matches)
- Pod `app-6c668bb6c8-dprq6`: `Running`, `ready=true`, UID `1000`, `runAsNonRoot=true`, `allowPrivilegeEscalation=false`, `seccomp=RuntimeDefault`, `drop=["ALL"]`
- `kubectl exec ... id` → `uid=1000 gid=1000`
- Namespace still carries `pod-security.kubernetes.io/enforce=restricted` (and audit/warn=restricted)
