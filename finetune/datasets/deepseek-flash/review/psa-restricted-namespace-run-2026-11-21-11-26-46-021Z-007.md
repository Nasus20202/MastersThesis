---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/psa-restricted-namespace/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# psa-restricted-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A workload in namespace restricted-app cannot start because its Pods are not being admitted. Make the workload run while keeping the namespace's security settings in place.

Documented in:

- content/en/docs/tasks/configure-pod-container/enforce-standards-namespace-labels.md: Requiring the `baseline` Pod Security Standard with namespace labels, Add labels to existing namespaces with `kubectl label` > Applying to a single namespace

## Write commands

1. `kubectl patch deploy app -n restricted-app --type=strategic --patch-file /tmp/patch.json`

## Final report

## Diagnosis

`namespace/restricted-app` enforces Pod Security Admission `restricted` (enforce/audit/warn = `restricted`), but the Deployment's pod template had `spec.template.spec.securityContext: {}` and no container securityContext. Every `ReplicaSet` pod creation was rejected:

```
FailedCreate: pods "app-787bf6bd59-..." is forbidden: violates PodSecurity "restricted:latest":
allowPrivilegeEscalation != false, unrestricted capabilities (must drop ["ALL"]),
runAsNonRoot != true, seccompProfile (must be RuntimeDefault or Localhost)
```

No ResourceQuota/LimitRange existed, and the node was Ready with no taints (the earlier `FailedScheduling` was transient and self-resolved), so the only blocker was the pod's security context.

## Change

Patched the owning `Deployment/app` pod template (namespace labels untouched, image/command/args/replicas kept as-is):

- pod: `runAsNonRoot: true`, `runAsUser/runAsGroup: 65534`, `seccompProfile.type: RuntimeDefault` (an explicit non-root UID was required because the `busybox:1.37.0` image defaults to UID 0, which `runAsNonRoot` alone would reject)
- container `app`: `allowPrivilegeEscalation: false`, `privileged: false`, `runAsNonRoot: true`, `runAsUser: 65534`, `capabilities.drop: ["ALL"]`

## Verification (fresh observations)

- Pod admitted and running: `app-6955575569-9bq7n  Running  1/1  restarts=0` (new RS `app-6955575569`; old failing RS scaled to 0).
- Deployment replica match: `replicas=1 updated=1 ready=1 available=1`, `Available=True`.
- Runtime proof the workload is actually up and non-root: `exec id` → `uid=65534(nobody) gid=65534(nobody)`; `ps` → `1 nobody sleep 3600`.
- Pod spec confirms `runAsUser=65534`, `allowPrivilegeEscalation=false`, `capabilities.drop=["ALL"]`, `seccompProfile=RuntimeDefault`; image is the intended `busybox:1.37.0`.
- Namespace security settings preserved: labels still `pod-security.kubernetes.io/{enforce,audit,warn}=restricted`. All remaining `FailedCreate` events belong to the deleted/drained ReplicaSet `app-787bf6bd59` (historical, none after the new pod was created).
