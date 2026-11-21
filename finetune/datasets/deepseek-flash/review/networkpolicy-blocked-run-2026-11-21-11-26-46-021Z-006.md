---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-blocked/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl apply -f -`

## Final report

**Root cause:** The `policy` namespace had only a catch‑all `default-deny-ingress` NetworkPolicy (`podSelector: {}`, Ingress). The companion allow rule that was supposed to permit the required client (`role=allowed`) to reach the `app=server` pods was missing, so every client — including `allowed-client` — timed out.

**Change made (one new, additive object; the deny policy untouched):**

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-allowed-client
  namespace: policy
spec:
  podSelector: { matchLabels: { app: server } }
  policyTypes: [Ingress]
  ingress:
    - from: [{ podSelector: { matchLabels: { role: allowed } } }]
      ports: [{ protocol: TCP, port: 80 }]
```

**Verified outcome (fresh observations after the change):**

| Check                                                            | Result                                                                                                                               |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `allowed-client` → `http://server.policy.svc.cluster.local`      | `200` – nginx welcome page returned (`exit 0`)                                                                                       |
| `allowed-client` → pod IP `192.168.8.133:80` (bypassing Service) | nginx page returned (`exit 0`) — allow applies to the server pods directly                                                           |
| `blocked-client` → Service `10.96.126.206:80`                    | `wget: download timed out` (`exit 1`) — still blocked                                                                                |
| `blocked-client` → pod IP `192.168.8.133:80`                     | `wget: download timed out` (`exit 1`) — still blocked even without Service                                                           |
| Policies now present                                             | `allow-allowed-client (app=server)` + `default-deny-ingress (<none>)` — deny still enforced, allow only for `role=allowed` on TCP 80 |
| `deployment/server`                                              | desired 1, ready 1, updated 1, available 1                                                                                           |

Only the required client was granted ingress to the server on TCP/80; all other traffic to the server remains blocked by the intact default-deny policy, and no other cluster settings were modified.
