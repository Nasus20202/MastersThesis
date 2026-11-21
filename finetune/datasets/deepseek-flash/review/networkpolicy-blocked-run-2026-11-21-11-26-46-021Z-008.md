---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-blocked/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl apply -f /workspace/allow-client-to-server.yaml`
2. `kubectl run other-ns-client --rm -i --restart=Never --image=busybox:1.37.0 -n default --command -- sh -c 'wget -qO- --timeout=5 http://server.policy.svc.cluster.local`

## Final report

**Cause:** The `policy` namespace has a `default-deny-ingress` NetworkPolicy selecting all pods (`podSelector: {}`), and no policy existed to permit the intended client. Both `allowed-client` (`role=allowed`) and `blocked-client` (`role=blocked`) timed out reaching `http://server`.

**Change:** Added one least-privilege ingress allow policy; the default deny was left intact.

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-allowed-client-to-server
  namespace: policy
spec:
  podSelector:
    matchLabels:
      app: server
  policyTypes: ["Ingress"]
  ingress:
    - from:
        - podSelector:
            matchLabels:
              role: allowed
      ports:
        - protocol: TCP
          port: 80
```

**Verified outcome (fresh observations):**

- `allowed-client` → `http://server.policy.svc.cluster.local`: succeeded, returned the nginx welcome page, `exit=0`.
- `blocked-client` → same URL: `wget: download timed out`, `exit=1` — still blocked.
- Transient pod in the `default` namespace → same URL: timed out, `exit=1` — cross-namespace traffic remains blocked.
- `default-deny-ingress` is still present and unchanged.
- `deployment/server`: replicas 1, ready 1, updated 1, available 1 — matches desired.
