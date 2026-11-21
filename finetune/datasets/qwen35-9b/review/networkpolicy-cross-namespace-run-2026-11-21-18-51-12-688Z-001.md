---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-18-51-12-688Z/networkpolicy-cross-namespace/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-cross-namespace · Qwen3.5-9B-Q4_K_M · run-2026-11-21-18-51-12-688Z · attempt 1

## Task

A client in one namespace can no longer reach the application in another namespace. Restore the intended cross-namespace access while keeping other namespaces blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Targeting multiple namespaces by label, Targeting a Namespace by its name

## Write commands

1. `edit_file {"new_string": "spec:\n ingress:\n - from:\n - namespaceSelector:\n matchLabels:\n name: client-ns\n podSelector:\n matchLabels:\n role: trusted\n ports:\n - port: 80\n protocol: TCP\n podSelector:\n matchLabels:\n app: server\n policyTypes:\n - Ingress", "old_string": "spec:\n ingress:\n - from:\n - podSelector:\n matchLabels:\n role: trusted\n ports:\n - port: 80\n protocol: TCP\n podSelector:\n matchLabels:\n app: server\n policyTypes:\n - Ingress", "path": "stdout"}`
2. `kubectl patch networkpolicy allow-clients -n server-ns --type=json -p='[{"op": "replace", "path": "/spec/ingress/0/from/0", "value": {"namespaceSelector": {"matchLabels": {"name": "client-ns"}}, "podSelector": {"matchLabels": {"role": "trusted"}}}}]' 2>/dev/null`
3. `kubectl patch networkpolicy allow-clients -n server-ns --type=json -p='[{"op": "replace", "path": "/spec/ingress/0/from/0", "value": {"namespaceSelector": {"matchLabels": {"kubernetes.io/metadata.name": "client-ns"}}, "podSelector": {"matchLabels": {"role": "trusted"}}}}]' 2>/dev/null`

## Final report

## Summary

**Issue Identified**: The NetworkPolicy `allow-clients` in `server-ns` only allowed same-namespace ingress with `role: trusted` label, blocking all cross-namespace traffic including from `client-ns`.

**Fix Applied**: Updated the NetworkPolicy to add a `namespaceSelector` that explicitly permits pods from the `client-ns` namespace.

### Verification Results

| Test                            | Result     | Evidence                             |
| ------------------------------- | ---------- | ------------------------------------ |
| client-ns → server.server-ns:80 | ✅ SUCCESS | Received nginx welcome HTML page     |
| other-ns → server.server-ns:80  | ✅ BLOCKED | Connection timed out (exit code 143) |

### Final NetworkPolicy State

The updated NetworkPolicy `allow-clients` in `server-ns` now:

- Allows ingress from pods with `role: trusted` label **only** in the `client-ns` namespace
- Blocks all other namespaces including `other-ns`

This restores the intended cross-namespace access while keeping other namespaces blocked as required.
