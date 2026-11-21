---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-cross-namespace/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-cross-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A client in one namespace can no longer reach the application in another namespace. Restore the intended cross-namespace access while keeping other namespaces blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Targeting multiple namespaces by label, Targeting a Namespace by its name

## Write commands

1. `kubectl patch networkpolicy allow-clients -n server-ns --type merge -p '{"spec":{"ingress":[{"from":[{"namespaceSelector":{"matchLabels":{"team":"clients"}},"podSelector":{"matchLabels":{"role":"trusted"}}}],"ports":[{"port":80,"protocol":"TCP"}]}]}}' 2>`

## Final report

## Root cause

The NetworkPolicy `allow-clients` in `server-ns` selected sources with a bare `podSelector`:

```yaml
from:
  - podSelector:
      matchLabels: { role: trusted }
```

In a NetworkPolicy, a `podSelector` in `from` is evaluated **only against pods in the policy's own namespace** (`server-ns`). With no `namespaceSelector`, no pod outside `server-ns` could ever match, so the `client-ns` client was denied even though its pod is labeled `role=trusted`. A second tell: both the allowed and the intended-blocked client carry `role=trusted`, so pod labels alone cannot distinguish them — the namespace labels (`team=clients` vs `team=others`) are the intended discriminator.

## Change made

Single merge patch to that one NetworkPolicy, adding a `namespaceSelector` so source pods are matched by namespace **and** pod label (AND semantics):

```yaml
from:
  - namespaceSelector:
      matchLabels: { team: clients }
    podSelector:
      matchLabels: { role: trusted }
ports:
  - port: 80
    protocol: TCP
```

The `podSelector: app=server`, `policyTypes: [Ingress]`, port 80/TCP restriction, and all other settings were left untouched.

## Verification (fresh observations)

| Check                                                                    | Result                                                                                               |
| ------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------- |
| `client-ns/allowed-client` → `http://server.server-ns.svc.cluster.local` | **200 OK** — response body contains `<title>Welcome to nginx!</title>`                               |
| `client-ns/allowed-client` → server pod IP `192.168.15.69:80` direct     | **200 OK**, nginx page returned (so this is policy, not DNS/Service plumbing)                        |
| `other-ns/blocked-client` → Service name                                 | **Blocked** — `wget: download timed out`, exit 1                                                     |
| `other-ns/blocked-client` → server pod IP direct                         | **Blocked** — timed out, exit 1                                                                      |
| `other-ns` DNS resolution                                                | Resolves correctly to `10.96.207.202`, proving the block is policy enforcement, not a lookup failure |
| Namespaces matching `team=clients`                                       | only `client-ns` — no unintended namespace gained access                                             |
| Namespaces matching `team=others`                                        | `other-ns` — remains excluded                                                                        |
| `server` deployment health                                               | `1/1` ready, up-to-date, available; pod Running                                                      |

The originally reported symptom (client timing out) is gone, and the constraint that other namespaces stay blocked is confirmed by direct request tests from `other-ns`.
