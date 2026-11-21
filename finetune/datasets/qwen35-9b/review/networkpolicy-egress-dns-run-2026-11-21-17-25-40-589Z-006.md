---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-17-25-40-589Z/networkpolicy-egress-dns/prompt/006.json
verdict: rejected
reason: 'first-pass review by a Haiku model: unsure: DNS rule "UDP/TCP port 53 -> any namespace" has no kube-system restriction; wider than needed, not reviewed. | second pass against the raw record: defect: Port 53 allowed to namespaceSelector {} (msg 48 apply) though CoreDNS runs in kube-system (msg 61); report gives no reason.'
---

# networkpolicy-egress-dns · Qwen3.5-9B-Q4_K_M · run-2026-11-21-17-25-40-589Z · attempt 6

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Write commands

1. `kubectl apply -f /tmp/egress-policy.yaml -n egress`
2. `kubectl apply -f /tmp/egress-policy.yaml -n egress`

## Final report

**Summary of Investigation and Fix:**

**Problem Identified:**
The egress namespace had a `default-deny-egress` NetworkPolicy that blocked all outbound traffic, preventing the app pod from reaching its dependency (the dep service).

**Solution Applied:**
Updated the existing `default-deny-egress` NetworkPolicy to add two specific egress rules:

1. **TCP port 80** → pods labeled `app=dep` (allows access to the application's dependency)
2. **UDP/TCP port 53** → any namespace (allows DNS resolution for service names)

**Verification Results:**

| Test                                    | Expected | Observed   | Status                    |
| --------------------------------------- | -------- | ---------- | ------------------------- |
| wget to `dep:80` from app               | Success  | Success ✅ | PASS                      |
| wget to dep ClusterIP (10.96.204.100)   | Success  | Success ✅ | PASS                      |
| nc to dep pod IP (192.168.93.5):80      | Success  | Success ✅ | PASS                      |
| Access to external (google.com/8.8.8.8) | Blocked  | Blocked ✅ | Confirms no-all-egress    |
| Access to other pods (app=other)        | Blocked  | Blocked ✅ | Confirms selective egress |

**Outcome:**
The application's required outbound access has been restored while maintaining the principle of least privilege - only specific traffic paths are now allowed, not all egress.
