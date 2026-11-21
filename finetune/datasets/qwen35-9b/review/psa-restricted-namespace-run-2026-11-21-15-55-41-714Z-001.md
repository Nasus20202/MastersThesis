---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/psa-restricted-namespace/prompt/001.json
verdict: accepted
reason: 'first-pass review by a Haiku model: unsure: "kubectl apply -f /tmp/deployment.yaml" twice; file never written in sheet, security context unverifiable | second pass against the raw record: ok: Only write is a heredoc apply (msg 16, 18, 24, 26) adding restricted securityContext; msg 29-33 show 1/1 Running and the report matches.'
---

# psa-restricted-namespace · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 1

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:

- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Write commands

1. `kubectl apply -f /tmp/deployment.yaml -n restricted-app`
2. `kubectl apply -f /tmp/deployment.yaml -n restricted-app`

## Final report

**Summary of the fix and verification:**

**Root Cause**: Pods were rejected due to Pod Security Standards (PSS) "restricted" policy violations. The deployment template lacked required security contexts:

- `allowPrivilegeEscalation=false` - missing
- `capabilities.drop=["ALL"]` - missing
- `runAsNonRoot=true` - missing
- `seccompProfile.type=RuntimeDefault` - missing
- `runAsUser` - needed to specify a non-root user (image runs as root by default)

**Fix Applied**: Updated the deployment in `restricted-app` namespace with proper security contexts:

```yaml
securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
      - ALL
  runAsNonRoot: true
  runAsUser: 1000
  seccompProfile:
    type: RuntimeDefault
```

**Verification Confirmed**:

| Check                           | Result                           |
| ------------------------------- | -------------------------------- |
| Deployment ready replicas       | 1/1                              |
| Deployment unavailable replicas | 0                                |
| Condition Available             | True (MinimumReplicasAvailable)  |
| Condition Progressing           | True (NewReplicaSetAvailable)    |
| Pod status                      | Running, Ready (1/1)             |
| Container runAsUser             | 1000 (non-root)                  |
| Container seccompProfile.type   | RuntimeDefault                   |
| Namespace security policy       | Unchanged (restricted: enforced) |

The workload `app` is now running successfully while the namespace's restricted Pod Security Standards remain enforced.
