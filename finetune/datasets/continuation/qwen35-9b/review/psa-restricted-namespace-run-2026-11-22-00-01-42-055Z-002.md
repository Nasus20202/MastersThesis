---
attempt: finetune/datasets/continuation/qwen35-9b/runs/run-2026-11-22-00-01-42-055Z/psa-restricted-namespace/prompt/002.json
verdict: accepted
reason: Final force-replace left the Deployment at 1/1 with a ready pod and only Normal events, as reported.
---
# psa-restricted-namespace · Qwen3.5-9B-Q4_K_M · run-2026-11-22-00-01-42-055Z · attempt 2

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

1. `kubectl replace -f deployment.yaml -n restricted-app`
2. `kubectl apply -f deployment.yaml -n restricted-app --force-conflicts`
3. `kubectl apply -f deployment.yaml --force`
4. `kubectl apply -f deployment.yaml`
5. `kubectl patch deployment app -n restricted-app --type=merge -p '{"spec":{"template":{"spec":{"containers":[{"securityContext":{"runAsUser":1000}}}]}}}'`
6. `kubectl replace -f current-deploy.yaml --force`

## Final report

The workload in namespace `restricted-app` is now running successfully while maintaining the namespace's security settings.

**Cause identified:** Pod creation was denied by the `PodSecurity "restricted:latest"` admission controller due to missing container security context fields (`allowPrivilegeEscalation`, `capabilities.drop`, `runAsNonRoot`, `seccompProfile`) and a subsequent Kubelet error because the `busybox` image runs as root while `runAsNonRoot: true` was enforced without an explicit UID.

**Fix applied:** Updated the Deployment `app` spec to include the required security context in the container definition, specifically adding `runAsUser: 1000` alongside other restrictions:
- `allowPrivilegeEscalation: false`
- `capabilities.drop: ["ALL"]`
- `runAsNonRoot: true`
- `seccompProfile.type: RuntimeDefault`

**Verification:**
- Pod status: `app-7cb4d5f976-bgpnd` is **Running** with **READY=1/1**.
- Deployment status: `app` shows **Ready=1/1**, **Up-to-date=1/1**, **Available=1/1**.
- Admission events: No "Forbidden" or policy violation errors for the new Pod; only Normal events (Scheduled, Pulled, Created, Started).

The workload is admitted and running while respecting the namespace's Pod Security Admission constraints.
