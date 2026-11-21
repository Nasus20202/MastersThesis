---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-17-25-40-589Z/networkpolicy-egress-dns/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-egress-dns · Qwen3.5-9B-Q4_K_M · run-2026-11-21-17-25-40-589Z · attempt 3

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Write commands

1. `kubectl apply -f /tmp/networkpolicy.yaml`
2. `kubectl apply -f /tmp/networkpolicy.yaml`
3. `kubectl apply -f /tmp/networkpolicy.yaml`
4. `kubectl apply -f /tmp/networkpolicy.yaml`

## Final report

I've successfully restored the application's required outbound access in the egress namespace. Here's what I found and fixed:

## Problem Identified

A `default-deny-egress` NetworkPolicy was blocking ALL egress traffic from all pods in the egress namespace, preventing:

1. DNS resolution (service name resolution) - app couldn't resolve `dep.egress.svc.cluster.local`
2. HTTP connectivity to the dependency service

## Solution Applied

Updated the `default-deny-egress` NetworkPolicy to add specific egress rules while maintaining restrictive egress controls:

```yaml
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    # Allow app pods to reach dep service on port 80
    - to:
        - podSelector:
            matchLabels:
              app: dep
      ports:
        - protocol: TCP
          port: 80
    # Allow DNS queries to kube-dns (port 53) only in kube-system namespace
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
      ports:
        - protocol: UDP
          port: 53
        - protocol: TCP
          port: 53
```

## Verification Results

1. **DNS Resolution** ✅
   - App pod can resolve `dep.egress.svc.cluster.local` → `10.96.212.50`
   - Command: `nslookup dep.egress.svc.cluster.local`

2. **HTTP Connectivity** ✅
   - App pod successfully reaches dep service on port 80
   - Command: `wget http://dep.egress.svc.cluster.local:80` (returns nginx welcome page)

3. **Policy Remains Restrictive** ✅
   - Only TCP/80 to pods with `app=dep` label is allowed
   - Only UDP/TCP/53 to kube-system namespace is allowed
   - All other egress traffic is blocked by the default deny rules
