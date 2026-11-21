---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-egress-dns/prompt/006.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-egress-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 6

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Write commands

1. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: allow-app-egress\n namespace: egress\nspec:\n podSelector:\n matchLabels:\n app: app\n policyTypes:\n - Egress\n egress:\n # Cluster DNS (CoreDNS in kube-system), UDP and TCP 53\n - to:\n - namespaceSelector:\n matchLabels:\n kubernetes.io/metadata.name: kube-system\n podSelector:\n matchLabels:\n k8s-app: kube-dns\n ports:\n - protocol: UDP\n port: 53\n - protocol: TCP\n port: 53\n # The application's dependency only\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - protocol: TCP\n port: 80\n", "path": "allow-app-egress.yaml"}`
2. `kubectl apply -f allow-app-egress.yaml`

## Final report

## Root cause

The `egress` namespace contained a catch-all `default-deny-egress` NetworkPolicy (`podSelector: {}`, `policyTypes: [Egress]`, no egress rules). It blocked _all_ outbound traffic from every pod in the namespace — including UDP/TCP 53 to CoreDNS — so service-name lookups failed and the `dep` dependency was unreachable.

Observed before the fix, from the `app` pod:

- `nslookup dep.egress.svc.cluster.local` → `connection timed out; no servers could be reached`
- `wget http://dep.egress.svc.cluster.local` → `bad address` (resolution failure)
- `wget http://10.96.14.112` → `download timed out`

## Fix

Left `default-deny-egress` in place (preserving the "no open egress" constraint) and added one additive, least-privilege policy selecting only the application pod:

`NetworkPolicy/egress/allow-app-egress` — `podSelector: app=app`, `policyTypes: [Egress]`, with exactly two egress rules:

1. DNS: to `namespaceSelector kubernetes.io/metadata.name=kube-system` + `podSelector k8s-app=kube-dns`, UDP/53 and TCP/53
2. Dependency: to `podSelector app=dep` (same namespace), TCP/80 only

## Verification (fresh observations from inside `app`)

| Check                                          | Result                                                                 |
| ---------------------------------------------- | ---------------------------------------------------------------------- |
| DNS via service name                           | `Server: 10.96.0.10`, `dep.egress.svc.cluster.local → 10.96.14.112` ✅ |
| `wget http://dep.egress.svc.cluster.local`     | `HTTP/1.1 200 OK`, `Server: nginx/1.31.5` ✅                           |
| `wget http://192.168.109.197` (dep pod IP)     | `HTTP/1.1 200 OK` ✅                                                   |
| `wget http://other.egress.svc.cluster.local`   | `download timed out` (blocked) ✅                                      |
| `wget http://192.168.109.198` (other pod IP)   | `download timed out` (blocked) ✅                                      |
| `wget http://1.1.1.1` and `http://example.com` | `download timed out` (external egress still blocked) ✅                |
| Pods                                           | `app`, `dep-*`, `other-*` all `Running`, `ready=true`; no restarts ✅  |

The application's dependency and cluster DNS are reachable again, while egress to `other` and to the internet remains denied — egress is scoped, not opened.
