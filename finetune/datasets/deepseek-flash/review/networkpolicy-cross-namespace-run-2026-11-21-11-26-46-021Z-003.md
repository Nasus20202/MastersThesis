---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-cross-namespace/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-cross-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A client in one namespace can no longer reach the application in another namespace. Restore the intended cross-namespace access while keeping other namespaces blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Targeting multiple namespaces by label, Targeting a Namespace by its name

## Write commands

1. `kubectl patch netpol -n server-ns allow-clients --type=json -p='[{"op":"replace","path":"/spec/ingress/0/from/0","value":{"namespaceSelector":{"matchLabels":{"team":"clients"}},"podSelector":{"matchLabels":{"role":"trusted"}}}}]'`

## Final report

## Diagnosis

`server-ns` has a single ingress NetworkPolicy, `allow-clients`, selecting pods with `app=server` on TCP/80. Its only `from` rule was:

```yaml
from: [{ podSelector: { matchLabels: { role: trusted } } }]
```

With no `namespaceSelector`, a `podSelector` inside `from` is scoped to **the policy's own namespace** (`server-ns`). No client pod lives there, so every cross-namespace request was denied — including the intended `client-ns/allowed-client` (label `role=trusted`). The namespace labels already encoded the intent: `client-ns: team=clients` vs `other-ns: team=others`.

## Change

Smallest persistent fix — patched the existing NetworkPolicy's single `from` element to AND the client pod label with the client namespace label (not a new/looser rule):

```yaml
from:
  - namespaceSelector:
      matchLabels: { team: clients }
    podSelector:
      matchLabels: { role: trusted }
ports: [{ port: 80, protocol: TCP }]
```

Because `namespaceSelector` and `podSelector` are in the **same** map element, they are ANDed: only `role=trusted` pods in namespaces labeled `team=clients` are admitted. Server selector, port, and `policyTypes` were left untouched.

## Verified outcome (fresh in-pod requests through the Service)

| Client                                      | Namespace labels | Result                                                                                          |
| ------------------------------------------- | ---------------- | ----------------------------------------------------------------------------------------------- |
| `client-ns/allowed-client` (`role=trusted`) | `team=clients`   | `wget http://server.server-ns.svc.cluster.local/` → **HTTP 200, full nginx welcome page**, rc=0 |
| `other-ns/blocked-client` (`role=trusted`)  | `team=others`    | `wget` → **download timed out**, rc=1 (still blocked)                                           |

Constraint checks: `other-ns` remains blocked even though its pod carries the same `role=trusted` label, confirming the restriction is namespace-based as required. Server side is healthy and unchanged: Deployment `server` 1/1 ready, up-to-date, available; Service `server` `10.96.234.81:80` → `app=server`; pod `Running 1/1` with no restarts.
