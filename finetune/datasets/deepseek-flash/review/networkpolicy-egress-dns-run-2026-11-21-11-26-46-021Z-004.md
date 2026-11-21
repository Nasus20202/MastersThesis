---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-egress-dns/prompt/004.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-egress-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 4

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Write commands

1. `kubectl auth can-i create networkpolicies -n egress`
2. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: allow-app-egress-dns-and-dep\n namespace: egress\nspec:\n podSelector:\n matchLabels:\n app: app\n policyTypes:\n - Egress\n egress:\n # Cluster DNS (CoreDNS) only\n - to:\n - namespaceSelector:\n matchLabels:\n kubernetes.io/metadata.name: kube-system\n podSelector:\n matchLabels:\n k8s-app: kube-dns\n ports:\n - protocol: UDP\n port: 53\n - protocol: TCP\n port: 53\n # Required dependency: dep, HTTP only\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - protocol: TCP\n port: 80\n", "path": "app-allow-egress.yaml"}`
3. `kubectl apply -f app-allow-egress.yaml`

## Final report

## Root cause

The `egress` namespace had a blanket deny policy, `default-deny-egress` (`podSelector: {}`, `policyTypes: [Egress]`), with **no allow rules at all**. It therefore blocked every pod's outbound traffic, including the `app` pod's DNS queries to CoreDNS (`10.96.0.10:53`) and its HTTP calls to `dep`. Baseline reproduction from `app`:

- `nslookup dep` → `;; connection timed out; no servers could be reached`
- `wget http://dep` → `bad address 'dep'`; `wget http://10.96.15.25` and `http://192.168.104.69` → `download timed out`

## Change

Added one narrowly scoped companion policy (default deny left untouched):

```yaml
kind: NetworkPolicy
metadata: { name: allow-app-egress-dns-and-dep, namespace: egress }
spec:
  podSelector: { matchLabels: { app: app } }
  policyTypes: [Egress]
  egress:
    - to:
        [
          {
            namespaceSelector:
              { matchLabels: { kubernetes.io/metadata.name: kube-system } },
            podSelector: { matchLabels: { k8s-app: kube-dns } },
          },
        ]
      ports: [{ protocol: UDP, port: 53 }, { protocol: TCP, port: 53 }]
    - to: [{ podSelector: { matchLabels: { app: dep } } }]
      ports: [{ protocol: TCP, port: 80 }]
```

Only the `app` pod is selected; only cluster DNS and `dep:80` are permitted.

## Verified from the `app` pod after the change

| Check                                     | Result                                                                    |
| ----------------------------------------- | ------------------------------------------------------------------------- |
| `nslookup dep`                            | resolves `dep.egress.svc.cluster.local → 10.96.15.25` via `10.96.0.10:53` |
| `wget http://dep`                         | nginx welcome page returned, `exit=0`                                     |
| `wget http://192.168.104.69` (dep pod IP) | `exit=0`                                                                  |

Constraints still enforced:

| Negative check                                | Result                                                   |
| --------------------------------------------- | -------------------------------------------------------- |
| `wget http://other` (name)                    | `download timed out`, `exit=1`                           |
| `wget http://10.96.136.29` (other svc IP)     | `download timed out`, `exit=1`                           |
| `wget https://1.1.1.1` (internet)             | `download timed out`, `exit=1`                           |
| `nslookup example.com 8.8.8.8` (external DNS) | `connection timed out`                                   |
| `default-deny-egress` spec                    | unchanged: `{"podSelector":{},"policyTypes":["Egress"]}` |

The `app` pod is `ready=true` with `restarts=0`; no pods were deleted or restarted, and no unrelated settings were modified.
