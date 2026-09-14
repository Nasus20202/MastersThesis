# Kubernetes workload security

Separate identity and authorization from the security properties of the
workload. A ServiceAccount identifies a Pod to the API; a Pod or container
`securityContext` controls aspects of its runtime; admission policy may reject
or mutate the requested configuration before a Pod exists.

- Compare Pod- and container-level settings such as user and group IDs,
  filesystem groups, privilege escalation, capabilities, privileged mode,
  read-only filesystems and seccomp with the observed error and policy.
- Check namespace Pod Security Admission labels or other admission policy when
  an object is rejected before creation. Do not confuse an admission rejection
  with a container that starts and later lacks a filesystem or capability.
- Keep ServiceAccount selection and RBAC analysis in `rbac.md`; changing the
  account is not a substitute for identifying the missing permission.
- Prefer the smallest security-context or policy change that explains the
  evidence. Do not grant privileged execution, host access or broad capabilities
  merely to bypass an uncertain runtime error.
- Treat image pull credentials, mounted tokens and other secrets as sensitive;
  inspect references and errors without printing their values.

Verify object acceptance, Pod admission and the actual container behavior after
any security-related change. Preserve existing isolation for unrelated Pods.
