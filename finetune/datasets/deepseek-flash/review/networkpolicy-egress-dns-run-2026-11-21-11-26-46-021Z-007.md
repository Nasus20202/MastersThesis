---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-egress-dns/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-egress-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Write commands

1. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: app-allow-egress\n namespace: egress\nspec:\n podSelector:\n matchLabels:\n app: app\n policyTypes:\n - Egress\n egress:\n # DNS resolution via cluster DNS only\n - to:\n - namespaceSelector:\n matchLabels:\n kubernetes.io/metadata.name: kube-system\n podSelector:\n matchLabels:\n k8s-app: kube-dns\n ports:\n - protocol: UDP\n port: 53\n - protocol: TCP\n port: 53\n # The application's required dependency only\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - protocol: TCP\n port: 80\n", "path": "/workspace/app-allow-egress.yaml"}`
2. `kubectl apply -f /workspace/app-allow-egress.yaml 2>`

## Final report

## Root cause

The `egress` namespace had a catch-all `default-deny-egress` NetworkPolicy (`podSelector: {}`, `policyTypes: [Egress]`) that selected **every** pod while providing no egress rules. That single persistent object was blocking both required paths from the `app` pod:

- DNS: `nslookup dep` returned `;; connection timed out; no servers could be reached`
- Dependency: `wget http://dep` returned `wget: bad address 'dep'`, and even by raw Pod IP (`192.168.86.197`) it timed out — so this was policy denial, not just a DNS problem.

## Change made

Added one scoped policy, `app-allow-egress` (namespace `egress`), selecting only `app=app`, with exactly two egress rules:

| Allowed destination                               | Port           |
| ------------------------------------------------- | -------------- |
| `kube-system` namespace + `k8s-app=kube-dns` pods | UDP/53, TCP/53 |
| `app=dep` pods in `egress`                        | TCP/80         |

`default-deny-egress` was left untouched, so everything not listed stays denied.

## Verified outcome (fresh observations from inside the pod)

Requested outcome — required outbound access restored:

- DNS: `nslookup dep` → `Name: dep.egress.svc.cluster.local  Address: 10.96.68.49`; FQDN resolves directly to `10.96.68.49`
- Dependency reachable by service name: `wget http://dep` → `<title>Welcome to nginx!</title>`, reproduced 3/3 consecutive attempts
- `app` pod: `1/1 Running`, `RESTARTS 0` (no restart used as the fix)

Constraints preserved — egress is **not** wide open:

- `app` → `other` service: `wget: download timed out` (denied)
- `app` → `1.1.1.1:443`: `wget: download timed out` (denied)
- `app` → `8.8.8.8:53`: `;; connection timed out; no servers could be reached` (only cluster DNS permitted)
- `other` pod → `dep:80`: blocked (`OTHER_BLOCKED_FROM_DEP`)
- `dep` pod → `app:80`: blocked (`DEP_BLOCKED_FROM_APP`)
