---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-egress-dns/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-egress-dns · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Write commands

1. `kubectl apply -f /tmp/allow-app-egress.yaml`

## Final report

## Diagnosis

The `egress` namespace contained a namespace-wide `default-deny-egress` NetworkPolicy (`podSelector: {}`, `policyTypes: [Egress]`) with **no allow rules**. Since this cluster uses Calico for policy enforcement, every pod in `egress` — including `app` (`app=app`) — had all outbound traffic dropped.

Observed symptoms before the fix:

- DNS: `nslookup dep` → `;; connection timed out; no servers could be reached` (the query to the `kube-dns` ClusterIP `10.96.0.10` was dropped, so service-name resolution failed).
- Direct pod-IP connection: `wget http://192.168.15.5` (the `dep` endpoint) → `download timed out`, proving the failure was the egress drop and not just DNS.
- `dep` itself was healthy: Deployment `dep` 1/1 ready, Service `dep` (`app=dep`) had endpoint `192.168.15.5:80`.

## Change

Added one persistent, additive policy (the default-deny policy was left intact, so nothing became "allow all"):

```yaml
kind: NetworkPolicy
metadata: { name: allow-app-egress, namespace: egress }
spec:
  podSelector: { matchLabels: { app: app } } # only the application pod
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

## Verification (fresh observations)

| Check                         | Result                                                                                              |
| ----------------------------- | --------------------------------------------------------------------------------------------------- |
| DNS resolution from `app`     | `dep.egress.svc.cluster.local` → `10.96.48.155` (resolved, server `10.96.0.10:53`)                  |
| Service request through `dep` | `wget http://dep.egress.svc.cluster.local` and short name `http://dep` → HTTP OK, served nginx page |
| Egress to _other_ dependency  | `wget http://other.egress.svc.cluster.local` and `http://192.168.15.6` → timed out (still blocked)  |
| Egress to internet            | `wget http://1.1.1.1` → timed out (still blocked)                                                   |
| Egress to Kubernetes API      | `https://10.96.0.1:443` → timed out (still blocked)                                                 |
| Pods unaffected               | `app`, `dep-...`, `other-...` all `1/1 Running`, `RESTARTS 0` — no restarts were needed             |
| Default deny preserved        | `default-deny-egress` still present alongside `allow-app-egress`                                    |

The application's required outbound paths (DNS to `kube-dns`, TCP 80 to `app=dep` pods) are restored, while all other egress remains denied.
