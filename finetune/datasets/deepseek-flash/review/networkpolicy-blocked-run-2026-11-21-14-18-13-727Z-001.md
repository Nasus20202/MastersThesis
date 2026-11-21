---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-14-18-13-727Z/networkpolicy-blocked/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · deepseek-flash · run-2026-11-21-14-18-13-727Z · attempt 1

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl apply -f - 2>`

## Final report

**Cause:** The `policy` namespace contained only `default-deny-ingress` (podSelector `{}`, policyTypes `Ingress`), so every ingress to the `server` pods was dropped — including from the required client `allowed-client` (`role=allowed`). There was no allow policy at all.

**Fix (single additive change):** created NetworkPolicy `allow-allowed-role-to-server` in `policy`, targeting `app=server`, Ingress only, allowing `from: podSelector role=allowed` on `TCP/80`. `default-deny-ingress` was left untouched, so all other traffic stays blocked.

**Verification (fresh observations through the Service, not just object status):**

| Check                                                                        | Result                                                                 |
| ---------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| `allowed-client` → `http://server.policy.svc.cluster.local/` (10.96.138.130) | `ALLOWED-OK` (HTTP 200, nginx index HTML returned)                     |
| `allowed-client` → server pod IP `192.168.199.197:80`                        | nginx HTML returned                                                    |
| `blocked-client` (`role=blocked`) → Service                                  | `BLOCKED-FAIL` (wget timeout) — still denied                           |
| `blocked-client` → server pod IP:80                                          | `BLOCKED-FAIL` — still denied                                          |
| `allowed-client` → Service port 80                                           | `PORT80-OPEN`                                                          |
| `allowed-client` → Service port 443                                          | `PORT443-BLOCKED` (silent drop, i.e. policy drop not just no listener) |
| Deployment `server` replicas                                                 | 1 desired / 1 current / 1 updated / 1 ready / 1 available              |

Policies now present: `default-deny-ingress` (unchanged) and `allow-allowed-role-to-server`. Only ingress from `role=allowed` on port 80 was enabled; the required client can reach the server again while all other traffic to it remains blocked.
