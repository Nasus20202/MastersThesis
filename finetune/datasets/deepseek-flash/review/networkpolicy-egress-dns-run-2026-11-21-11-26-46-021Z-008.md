---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-egress-dns/prompt/008.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-egress-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 8

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Write commands

1. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: allow-app-egress\n namespace: egress\nspec:\n podSelector:\n matchLabels:\n app: app\n policyTypes:\n - Egress\n egress:\n - to:\n - namespaceSelector:\n matchLabels:\n kubernetes.io/metadata.name: kube-system\n podSelector:\n matchLabels:\n k8s-app: kube-dns\n ports:\n - protocol: UDP\n port: 53\n - protocol: TCP\n port: 53\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - protocol: TCP\n port: 80\n", "path": "/tmp/allow-app-egress.yaml"}`
2. `kubectl apply -f /tmp/allow-app-egress.yaml`

## Final report

## Root cause

The `egress` namespace had a namespace-wide egress deny (`default-deny-egress`: `podSelector: {}`, `policyTypes: [Egress]`) with **no allow rules at all**. Because a `NetworkPolicy` with an empty pod selector and `Egress` in policy types blocks _all_ outbound traffic from every pod in the namespace — including the mandatory DNS traffic to CoreDNS — the `app` pod could neither resolve names (`nslookup` → "connection timed out; no servers could be reached") nor reach its `dep` dependency.

Baseline observed before the fix (from inside `app`):

| Target                                    | Result                                              |
| ----------------------------------------- | --------------------------------------------------- |
| `nslookup dep.egress.svc.cluster.local`   | `connection timed out; no servers could be reached` |
| `wget http://10.96.102.67/` (dep Service) | timeout, rc=1                                       |
| `wget http://192.168.112.133/` (dep pod)  | timeout, rc=1                                       |

## Fix

I kept the namespace default-deny intact (deleting it would have allowed all egress) and added one narrowly scoped policy granting only the two flows the app genuinely needs — DNS plus the dependency:

`egress/allow-app-egress`:

- `podSelector: app=app` (only the application pod)
- egress #1: `namespaceSelector kubernetes.io/metadata.name=kube-system` + `podSelector k8s-app=kube-dns`, ports UDP/53 and TCP/53
- egress #2: `podSelector app=dep` in the same namespace, port TCP/80

No other objects were changed; pods were never restarted (`RESTARTS 0`).

## Verification (fresh observations)

Requested outcome — name resolution and dependency access from `app`:

- `nslookup dep.egress.svc.cluster.local` → `Address: 10.96.102.67` (resolves)
- `wget -S http://dep.egress.svc.cluster.local/` → `HTTP/1.1 200 OK`, body `<title>Welcome to nginx!</title>`
- `wget http://dep/` → rc=0 (short name works); direct pod IP `192.168.112.133:80` → rc=0
- Scope is port-limited: `nc 192.168.112.133 81` → rc=1 (denied, only TCP/80 allowed)

Constraint preserved — not "all egress" (from inside `app`):

- `http://10.96.150.208/` (`other` Service) → timeout rc=1; `http://other.egress.svc.cluster.local/` → timeout rc=1
- `http://1.1.1.1/` (internet) → timeout rc=1
- `nc 10.96.0.1 443` (Kubernetes API) → rc=1

Other pods stay fully denied, confirmed with `/dev/tcp` probes (rc=124 = timeout):

- `other` → dep `10.96.102.67:80` rc=124, `other` → kube-dns `10.96.0.10:53` rc=124
- `dep` → `other` `10.96.150.208:80` rc=124, `dep` → kube-dns rc=124

Final policies in the namespace: `default-deny-egress` (podSelector `{}`, Egress) unchanged, plus `allow-app-egress` (podSelector `app=app`) with only DNS:53 and dep:80 egress rules.
