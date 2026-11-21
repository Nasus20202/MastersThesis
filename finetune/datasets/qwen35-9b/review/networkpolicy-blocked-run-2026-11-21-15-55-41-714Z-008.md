---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/networkpolicy-blocked/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 8

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl run -n policy networkpolicy-allow-server --rm --image=busybox:1.37.0 -c allow-from-allowed -- dry-run=Client -o yaml`
2. `kubectl apply -f /tmp/allow-server-networkpolicy.yaml -n policy`

## Final report

The issue has been resolved. Here's the summary:

## Root Cause

A `default-deny-ingress` NetworkPolicy was blocking all incoming traffic to pods in the policy namespace, including from the required `allowed-client`.

## Fix Applied

Created a new NetworkPolicy (`allow-from-allowed`) that allows ingress traffic to server pods (labeled `app: server`) only from pods with label `role: allowed`.

### Network Policies Now In Place:

1. **`default-deny-ingress`**: Blocks all ingress by default
2. **`allow-from-allowed`**: Allows ingress to `app=server` pods from pods matching `role=allowed`

## Verification Results:

| Client                      | Status       | Result                                           |
| --------------------------- | ------------ | ------------------------------------------------ |
| **allowed-client** → server | ✗ Before fix | ✓ Now reachable (nginx welcome page returned)    |
| **blocked-client** → server | - After fix  | ✗ Still blocked (connection timeout as expected) |

The `allowed-client` now successfully receives responses from the application server while all other traffic (including `blocked-client`) remains blocked.
