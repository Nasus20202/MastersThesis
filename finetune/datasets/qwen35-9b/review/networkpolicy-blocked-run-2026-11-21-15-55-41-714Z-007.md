---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/networkpolicy-blocked/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 7

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl run temp-test --rm -it --image=busybox:1.37.0 --namespace=policy -- netcat -z 192.168.150.133 80`
2. `kubectl apply -n policy -f - << 'EOF'`

## Final report

## Summary

**Problem**: The allowed-client could not reach the application server in the policy namespace due to an existing default-deny NetworkPolicy blocking all ingress traffic.

**Solution**: Created a new NetworkPolicy `allow-allowed-client` that:

- Applies to pods with label `app: server` (the application server)
- Allows ingress only from pods with label `role: allowed` (the required client)
- The existing `default-deny-ingress` continues to block all other traffic

**Verification**:

1. **allowed-client access restored**: Successfully connects to the server and receives nginx welcome page response
2. **blocked-client remains blocked**: Connection attempts time out (blocked by default-deny-ingress policy)

**NetworkPolicies in policy namespace**:

- `default-deny-ingress`: Blocks all ingress to all pods (empty podSelector, no ingress rules = implicit deny)
- `allow-allowed-client`: Exception allowing traffic from `role: allowed` pods to the server (`app: server`)
