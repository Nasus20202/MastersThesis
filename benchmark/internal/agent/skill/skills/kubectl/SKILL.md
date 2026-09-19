---
name: kubectl
description: "kubectl command syntax and usage: resource addressing, output formats, logs, waits, and declarative changes."
---

# kubectl

Use this skill for kubectl command knowledge: how resources are addressed, how output is shaped, and what common subcommands do. Use the `kubernetes` skill for resource semantics and the `troubleshooting` skill for a general diagnosis workflow.

## Context, scope, and addressing

Every namespaced command is evaluated against a context, a cluster, and a namespace. The current context and its default namespace are defaults; pass `-n "$namespace"` when an explicit namespace matters and use `--all-namespaces` only for deliberate discovery:

```bash
kubectl config current-context
kubectl config view --minify --output='jsonpath={..namespace}{"\n"}'
kubectl get namespace
```

Resources can be addressed by type, type/name, labels, or field selectors, for example `pods`, `deployment/my-app`, and `pods -l 'app=my-app'`.

Authorization checks can be made without performing the operation:

```bash
kubectl auth can-i get pods -n "$namespace"
```

If observations involve `Forbidden`/403 responses, ServiceAccounts, Roles, RoleBindings, or least-privilege constraints, use the `kubernetes` skill and its `authorization.md` reference for the authorization semantics before changing RBAC objects.

## Reading objects

`get` lists or summarizes resources; `describe` presents a human-readable view including related conditions and events. YAML and JSON output expose the API object fields directly:

```bash
kubectl get pods -n "$namespace" -o wide
kubectl get deployment/my-app -n "$namespace" -o yaml
kubectl describe deployment/my-app -n "$namespace"
```

Use `-o jsonpath=...`, `-o custom-columns=...`, or `-o go-template=...` when only specific fields are needed. Prefer structured output when exact values matter.

## Logs and exec

`logs` reads container output. `--previous` reads the previous terminated container instance when one exists; `-c` selects a container; `--tail` bounds output. `exec --` runs a command in a running container:

```bash
kubectl logs "$pod" -n "$namespace" -c "$container" --tail=200
kubectl logs "$pod" -n "$namespace" -c "$container" --previous --tail=200
kubectl exec "$pod" -n "$namespace" -c "$container" -- command args
```

For multi-container overview output, `--all-containers --prefix` keeps lines attributable to their container.

## Events and waits

Events are namespace-scoped resources and can be sorted or filtered using supported selectors:

```bash
kubectl get events -n "$namespace" --sort-by=.lastTimestamp
kubectl get events -n "$namespace" --field-selector involvedObject.name="$name"
```

Bound waits explicitly rather than using an indefinite watch:

```bash
kubectl wait --for=condition=Ready pod -l 'app=my-app' -n "$namespace" --timeout=90s
kubectl rollout status deployment/my-app -n "$namespace" --timeout=90s
```

`rollout history` lists controller revisions where the resource supports rollout history.

## Changing state

`diff` previews declarative changes and `apply` submits them for reconciliation. `patch` changes selected fields, `set` provides targeted helpers for supported resource fields, and `delete` removes an object according to Kubernetes deletion and ownership semantics:

```bash
kubectl diff -f change.yaml -n "$namespace"
kubectl apply -f change.yaml -n "$namespace"
```

An accepted command reports that the API request succeeded; asynchronous controllers may still need time to reconcile the resulting state.
