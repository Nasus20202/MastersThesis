---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/networkpolicy-blocked/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 2

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl run -n policy -it --rm --restart=Never curl-test --image=curlimages/curl:latest -- sleep 3600`
2. `kubectl run -n policy --rm --restart=Never curl-test-curl --image=curlimages/curl:latest -- sleep 3600`
3. `kubectl run -n policy --rm --restart=Never curl-test-curl --image=curlimages/curl:latest -- sleep 3600`
4. `kubectl apply -f /tmp/networkpolicy.yaml -n policy`
5. `kubectl run -n policy curl-test --image=curlimages/curl:latest -- sleep 300`
6. `kubectl run -n default curl-test-external --rm --restart=Never --image=curlimages/curl:latest`
7. `kubectl run -n default curl-test-external --restart=Never --image=curlimages/curl:latest`
8. `kubectl run -n default curl-test-external --restart=Never --image=busybox:1.37 -- wget 2>/dev/null`
9. `kubectl run -n default curl-test-external --restart=Never --image=nginx:latest`

## Final report

## Summary

**Problem**: The policy namespace had only a `default-deny-ingress` NetworkPolicy blocking all ingress traffic, preventing the required client from reaching the application server.

**Solution**: Added a new NetworkPolicy `allow-allowed-client` that permits ingress from pods labeled `role: allowed` to server pods (labeled `app: server`) on port 80/TCP.

**Verification Results**:

| Client           | Access Status | Evidence                                    |
| ---------------- | ------------- | ------------------------------------------- |
| `allowed-client` | **SUCCESS**   | wget returns nginx HTML response (HTTP 200) |
| `blocked-client` | **BLOCKED**   | Connection times out after TCP handshake    |

**Current NetworkPolicies**:

1. `allow-allowed-client` - Allows ingress from `role: allowed` pods to `app: server` pods
2. `default-deny-ingress` - Blocks all other ingress (preserves blocking for unauthorized clients)

The fix restores access for the required client while maintaining the intended security posture where all other traffic remains blocked.
