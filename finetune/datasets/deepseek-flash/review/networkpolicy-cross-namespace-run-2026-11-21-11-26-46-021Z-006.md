---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-cross-namespace/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-cross-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

A client in one namespace can no longer reach the application in another namespace. Restore the intended cross-namespace access while keeping other namespaces blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Targeting multiple namespaces by label, Targeting a Namespace by its name

## Write commands

1. `kubectl patch networkpolicy allow-clients -n server-ns --type=json -p='[{"op":"replace","path":"/spec/ingress/0/from","value":[{"namespaceSelector":{"matchLabels":{"team":"clients"}},"podSelector":{"matchLabels":{"role":"trusted"}}}]}]' 2>`

## Final report

Cross-namespace access is restored and the block on other namespaces is intact.

## Root cause

The only policy in the cluster, `server-ns/allow-clients`, selected sources with a bare `podSelector`:

```yaml
from:
  - podSelector: { matchLabels: { role: "trusted" } } # no namespaceSelector
```

A `podSelector` in a `from` element without a `namespaceSelector` matches **only pods in the policy's own namespace** (`server-ns`). No pod there has `role=trusted`, so every cross-namespace request was dropped — even though `allowed-client` (`client-ns`) does carry `role=trusted`. Evidence that this was the gate, not a server problem: the `server` Deployment was `1/1` Ready with endpoint `192.168.144.5:80`, but `wget` from `allowed-client` to the Service _and_ `nc` to the pod IP both timed out.

Note that `other-ns/blocked-client` also has `role=trusted`, so a fix keyed only on the pod label (e.g. adding a bare `namespaceSelector: {}`) would have opened the wrong namespace too.

## Change made (one field on the existing policy, nothing else touched)

Replaced the `from` entry so both selectors are ANDed in a single element, scoping the trusted-pod rule to the client team's namespace label (`client-ns` carries `team=clients`; `other-ns` carries `team=others`):

```yaml
from:
  - namespaceSelector: { matchLabels: { team: "clients" } }
    podSelector: { matchLabels: { role: "trusted" } }
```

Ports (`TCP/80`), the `app=server` pod selector, `policyTypes: [Ingress]`, and all other resources were left unchanged. Still exactly one NetworkPolicy in the cluster.

## Verification (fresh observations)

| Check                            | Command                                                                              | Result                                                                    |
| -------------------------------- | ------------------------------------------------------------------------------------ | ------------------------------------------------------------------------- |
| Reported symptom gone            | `wget` from `client-ns/allowed-client` → `http://server.server-ns.svc.cluster.local` | `200` — full nginx "Welcome" page returned (attempt 1, after propagation) |
| Also via ClusterIP               | `wget http://10.96.231.248`                                                          | `<title>Welcome to nginx!</title>`                                        |
| Other namespace stays blocked    | `wget` from `other-ns/blocked-client` → same URL, 3 tries                            | `download timed out` ×3, exit 1                                           |
| Blocked at L3/L4, not just DNS   | `nc -z -w 3 192.168.144.5 80` from `other-ns`                                        | `PORT_BLOCKED_AS_EXPECTED`                                                |
| Server health independent of fix | `kubectl get deploy/ep server -n server-ns`                                          | `1/1 Ready`, endpoint `192.168.144.5:80`                                  |

The `other-ns` result is the key control: that pod holds the same `role=trusted` label as the allowed client, so its continued timeout proves the namespace restriction is genuinely being enforced rather than the policy having been loosened.
