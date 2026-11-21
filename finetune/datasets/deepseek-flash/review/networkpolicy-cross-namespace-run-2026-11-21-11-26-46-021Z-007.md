---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-cross-namespace/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-cross-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A client in one namespace can no longer reach the application in another namespace. Restore the intended cross-namespace access while keeping other namespaces blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Targeting multiple namespaces by label, Targeting a Namespace by its name

## Write commands

1. `kubectl apply -f - 2>`

## Final report

## Root cause

The NetworkPolicy `server-ns/allow-clients` had an ingress peer with a **`podSelector` but no `namespaceSelector`**:

```yaml
from:
  - podSelector:
      matchLabels:
        role: trusted
```

In a NetworkPolicy, a `podSelector` on its own selects pods **only in the policy's own namespace** (`server-ns`). No pod in `server-ns` carries `role=trusted` — the only pod there is `app=server`. So the allow rule matched nothing, and the default-deny created by `policyTypes: [Ingress]` blocked every client, cross-namespace or not. Both `client-ns/allowed-client` and `other-ns/blocked-client` timed out before the fix.

A naive fix of adding a separate `namespaceSelector` list entry would have opened the port to _any_ namespace, since both client pods share the label `role=trusted`. The namespace and pod selectors had to be combined in a single `from` element (AND semantics), so I scoped the namespace by its existing `team=clients` label.

## Change

Applied one field to the same NetworkPolicy — adding `namespaceSelector: {team: clients}` alongside the existing `podSelector` in the same peer (generation 2 → 3). The `podSelector: app=server`, `policyTypes: [Ingress]`, and the port-80-only restriction were left untouched.

## Verification

| Check                                      | Result                                                                      |
| ------------------------------------------ | --------------------------------------------------------------------------- |
| `allowed-client` → `server.server-ns:80`   | **HTTP 200** with nginx page, `EXIT:0` on 6/6 consecutive probes            |
| `blocked-client` (other-ns) → same address | **timed out**, `EXIT:1` on 6/6 probes                                       |
| `allowed-client` → port 8080               | **timed out** — port restriction still enforced                             |
| `server-ns` deployment                     | **1/1 ready, 1 up-to-date, 1 available**; Service `10.96.174.247:80` intact |

Intended cross-namespace access from `client-ns` is restored, `other-ns` remains blocked, and the original port limitation and server workload were left unchanged.
