---
name: authorization
description: "Repairing RBAC and ServiceAccount faults: forbidden errors, least privilege and service account token automount."
---

# Authorization

Faults here are about identity and scope. Grant the smallest permission the workload actually uses, and never widen access to make a rejection disappear.

## Diagnose a rejection

- Read the `Forbidden` message: it names the verb, the resource and the namespace or the cluster scope.
- Identify the acting identity: `kubectl get pod <pod> -n <ns> -o jsonpath='{.spec.serviceAccountName}'`, then the Role/ClusterRole bound to it.
- Test the exact permission without performing it:

```bash
kubectl auth can-i get configmaps -n <ns> --as=system:serviceaccount:<ns>:<sa>
```

## Repair

- Correct the Role `rules` (`apiGroups`, `resources`, `verbs`, and `resourceNames` when scope must be narrow) or the binding subject/namespace. Do not add `*` or bind `cluster-admin`.
- **Overprivileged.** Remove verbs, resources or wildcards the workload does not use while keeping the access it needs. The target is the smallest set that still passes the required `auth can-i` check.
- **Token automount.** A workload that does not talk to the API does not need a token. Set `automountServiceAccountToken: false` on the ServiceAccount (or the Pod template) and confirm the projected token is gone while the workload still runs.
- Increase scope only when the task explicitly requires the workload to reach more; otherwise keep the current scope.

## Verify

Run `kubectl auth can-i` for the required action (must be allowed) and for an action the workload must not perform (must be denied), and confirm the workload is still Running/Ready. Report the exact checks and results.
