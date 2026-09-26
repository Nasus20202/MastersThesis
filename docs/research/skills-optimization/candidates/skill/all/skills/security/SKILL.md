---
name: security
description: Making workloads compliant with Pod Security Admission and container security contexts, including read-only root filesystems.
---

# Security

A Pod that fails admission or a container security check is never created or never starts. Keep the namespace policy and make the object compliant.

## Pod Security Admission

- A namespace enables a profile through the `pod-security.kubernetes.io/enforce` label (also `audit` and `warn`). Keep the label in place; the task is to make the Pod compliant.
- The `restricted` profile requires every container to set `allowPrivilegeEscalation: false`, drop all capabilities (`capabilities.drop: ["ALL"]`), set `runAsNonRoot: true`, and set a seccomp profile (`seccompProfile.type: RuntimeDefault` or `Localhost`).
- `runAsNonRoot: true` alone is not enough when the image's default user is root or unspecified: the kubelet refuses to start the container with a `CreateContainerConfigError` about running as root. Set a non-zero `runAsUser` (for example `65532`) or choose a non-root image.
- Pod-level `securityContext` supplies defaults; container-level values override it. Some profiles require the fields on the container.

## Read-only root filesystem

- When the root filesystem is read-only, an application that writes under `/tmp`, `/var/cache`, `/var/run` or a service directory fails. Add a writable `emptyDir` volume at the exact path the process writes and keep the root filesystem read-only. Find the path from the container log or crash reason.

## Verify

Confirm the Pod is actually admitted and Ready (a rejected Pod never appears, so register the absence of a Pod as the failure signal), the effective UID is non-zero, and the root filesystem remains read-only with the added writable mount.
