# Kubernetes Security

Workload security combines API identity, authorization, process isolation, image provenance, secret handling, and network controls.

Pods run as a ServiceAccount identity for Kubernetes API access. A projected ServiceAccount token is commonly mounted into Pods by default, but token automounting can be disabled. RBAC determines what that identity may do independently of the Linux identity used by the container process.

`securityContext` controls settings such as user and group IDs, privilege escalation, Linux capabilities, read-only filesystems, and seccomp or AppArmor configuration. Pod Security Standards group common workload restrictions into `privileged`, `baseline`, and `restricted` profiles that can be enforced through Pod Security Admission.

Secrets are base64-encoded API data, not encryption by themselves. Access control, encryption at rest, and external secret-management systems determine how secret values are protected.

NetworkPolicy can restrict Pod ingress and egress where the cluster networking implementation supports it. Image digests, registry authentication, and admission controls can provide additional supply-chain and provenance guarantees.

Pod Security admission enforces a profile per namespace through the `pod-security.kubernetes.io/enforce` (and `audit` and `warn`) labels. The `restricted` profile requires every container to set `allowPrivilegeEscalation: false`, drop all capabilities (`capabilities.drop: ["ALL"]`), set `runAsNonRoot: true`, and set a seccomp profile (`seccompProfile.type: RuntimeDefault` or `Localhost`). Setting `runAsNonRoot: true` on an image whose default user is root is not enough: without a non-zero `runAsUser` the kubelet refuses to start the container with a `CreateContainerConfigError` about running as root. Choose a non-root image or set `runAsUser` to a non-zero value. A Pod rejected by admission is never created, so workload child counts diverge from expectations; make the Pod compliant rather than weakening the namespace's enforce label.
