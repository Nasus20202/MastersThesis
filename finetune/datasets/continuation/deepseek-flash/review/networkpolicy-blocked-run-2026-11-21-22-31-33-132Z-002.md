---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/networkpolicy-blocked/prompt/002.json
verdict: accepted
reason: 'Restricts ingress to role=allowed on port 80, verified by the allowed client fetching the page and the blocked client timing out.'
---
# networkpolicy-blocked · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 2

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:
- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Inherited stop

The first 13 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 1
stdout:

stderr:
Unable to use a TTY - input is not a terminal or the right kind of file
error: Internal error occurred: Internal error occurred: error executing command in container: failed to exec in container: failed to start exec "5dbd2fded17741a6a456d5b90ebc564a5fdeeefce6f876366b634bd41ca1d46b": OCI runtime exec failed: exec failed: unable to start container process: exec: "curl": executable file not found in $PATH

Its writes:

1. `kubectl patch networkpolicy default-deny-ingress -n policy --type='json' -p='[{"op": "add", "path": "/spec/ingress", "value": [{"from": [{"podSelector": {"matchLabels": {"role": "allowed"}}}]}, {"ports": [{"protocol": "TCP", "port": 80}]}]}]'`

## Teacher's write commands

1. `kubectl patch networkpolicy default-deny-ingress -n policy --type='json' -p='[{"op": "replace", "path": "/spec/ingress", "value": [{"from": [{"podSelector": {"matchLabels": {"role": "allowed"}}}], "ports": [{"protocol": "TCP", "port": 80}]}]}]'`

## Final report

The issue was caused by the `default-deny-ingress` NetworkPolicy in the `policy` namespace, which selected all pods (`podSelector: {}`) and had no ingress rules, thereby blocking all incoming traffic to the application server. 

I updated the `default-deny-ingress` policy to include a single, specific ingress rule that allows traffic **only** from pods labeled `role=allowed` on TCP port 80, while leaving the default deny behavior intact for everything else.

**Changes applied:**
```yaml
spec:
  ingress:
  - from:
    - podSelector:
        matchLabels:
          role: allowed
    ports:
    - port: 80
      protocol: TCP
  podSelector: {}
  policyTypes:
  - Ingress
```

**Verification:**
- **Required client restored:** Executed `wget` from the `allowed-client` pod (labeled `role=allowed`). It successfully connected to the `server` service and retrieved the nginx welcome page.
- **Other traffic remains blocked:** Executed `wget` from the `blocked-client` pod (labeled `role=blocked`). The connection failed with a download timeout, confirming that unauthorized traffic to the server is still suppressed.
