---
name: playbook
description: Concrete symptom checks, outcome proofs, common wrong fixes and field templates for Kubernetes repair.
---

# Kubernetes repair playbook

Concrete rules for the failure classes that recur in this environment. Use them with the `troubleshooting` workflow.

## Diagnose by symptom

Choose the check that separates the candidate causes, then act.

- **Service unreachable.** Check ready endpoints, then compare the Service `targetPort` with the container `containerPort` and port name, then the Service `selector` with the Pod labels, then NetworkPolicy. Ready endpoints and healthy Pods do not prove the Service routes to the right port.
- **Pod not Ready or restarting.** Read the container waiting or terminated reason: `CreateContainerConfigError` is a missing or wrong ConfigMap or Secret reference; `ImagePullBackOff` is the image reference or pull secret; `OOMKilled` is the memory limit; repeated liveness kills are the liveness probe; argument or command errors are the container command. Fix the controller's Pod template, not the generated Pod.
- **Rollout not completing.** Compare `spec.replicas` with updated, ready and available replicas and read ReplicaSet or ReplicaFailure events. Look for immutable fields (for example a Deployment `selector`), unsatisfiable constraints, or a new Pod that cannot start.
- **Forbidden or 403.** Identify the ServiceAccount, verb, resource and namespace, and test the exact permission with `kubectl auth can-i`. Correct the subject or the scoped Role; do not broaden access.
- **Configuration value not applied.** Environment variables are read at container start; changing the source does not update a running container. Trigger a new rollout and verify the new Pods report the value.
- **Admission rejection.** Read the reason: Pod Security fields, LimitRange defaults, or ResourceQuota. Make the object compliant, not the namespace laxer.
- **PersistentVolumeClaim pending.** Compare the claim's class, access modes, capacity and selector with the available volumes.

## Prove the outcome

Make the direct observation for the requested outcome before reporting success.

- Service reachable: a request from a client Pod through the Service succeeds, not merely ready endpoints.
- Rollout converged: `replicas == updatedReplicas == readyReplicas == availableReplicas`.
- Image current: every Pod runs the intended image reference.
- Configuration applied: the running Pod reports the new value.
- Least privilege: the exact `auth can-i` check passes and the Role grants nothing unused.
- Network isolation: the allowed path succeeds and the disallowed path is still blocked.
- Namespace policy intact: the enforce label is unchanged.

## Common wrong fixes

- Deleting or restarting a generated Pod instead of fixing the controller template.
- `kubectl rollout restart` for a moved image tag: use `imagePullPolicy: Always` or the digest.
- Patching a Service `type` when the fault is `targetPort` or the selector.
- Setting `runAsNonRoot: true` without a non-root `runAsUser` or a non-root image.
- Guessing NetworkPolicy peer labels or ports instead of reading the client's namespace and labels.
- Weakening a Role or namespace policy so a rejection disappears.
- Claiming success from old ready Pods, an accepted API call, or a diagnosis without a repair.

## Field templates

Restricted container security context:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 65532
  allowPrivilegeEscalation: false
  capabilities:
    drop: ["ALL"]
  seccompProfile:
    type: RuntimeDefault
```

NetworkPolicy that allows cluster DNS and one dependency while keeping other egress denied:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-required
spec:
  podSelector:
    matchLabels:
      app: <server>
  policyTypes: [Egress]
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
      ports:
        - { protocol: UDP, port: 53 }
        - { protocol: TCP, port: 53 }
    - to:
        - podSelector:
            matchLabels:
              app: <dependency>
      ports:
        - { protocol: TCP, port: <port> }
```
