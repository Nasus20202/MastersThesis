---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-blocked/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl apply -f -`
2. `kubectl run other-ns-test -n default --image=busybox:1.37.0 --restart=Never --labels=role=allowed --command -- sh -c 'wget -qO- --timeout=5 http://server.policy.svc.cluster.local/ >/dev/null`
3. `kubectl delete pod -n default other-ns-test --wait=false`

## Final report

Diagnosis and fix complete.

## What was wrong

The `policy` namespace had only a catch-all deny policy:

- `default-deny-ingress`: `podSelector: {}`, policyTypes `[Ingress]` → all ingress to every pod in the namespace was dropped.

No allow policy existed for the required client, so traffic from `allowed-client` (`role=allowed`) to the server (`app=server`, port 80 via Service `server` / pod IP 192.168.179.133) timed out:

```
kubectl exec -n policy allowed-client -- wget http://server.policy...
  → wget: download timed out  (exit 1)
```

## Change made (single additive policy, deny untouched)

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
  policyTypes: [Ingress]
  ingress:
    - from:
        - podSelector:
            matchLabels:
              role: allowed
      ports:
        - protocol: TCP
          port: 80
```

The `podSelector` (not a namespace selector) keeps the allowance namespace-local, and the port restriction limits it to the HTTP port only. `default-deny-ingress` was left intact.

## Verification (fresh observations)

| Check                                                        | Result                                                                                                      |
| ------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------- |
| `allowed-client` → `http://server.policy.svc.cluster.local/` | HTTP 200, served nginx "Welcome to nginx!" page, exit 0                                                     |
| `allowed-client` → pod IP `192.168.179.133:80`               | exit 0 (direct pod path also works)                                                                         |
| `blocked-client` → Service                                   | `wget: download timed out`, exit 1                                                                          |
| `blocked-client` → pod IP                                    | `wget: download timed out`, exit 1                                                                          |
| `allowed-client` → port 8080 on server                       | blocked, exit 1 (only TCP/80 permitted)                                                                     |
| Pod in `default` ns labeled `role=allowed` → Service         | `FAIL` / timeout — cross-namespace traffic still blocked                                                    |
| Policies still present                                       | `allow-allowed-client-to-server` (app=server) + `default-deny-ingress` ({} → all ingress denied by default) |
| Server workload                                              | Deployment `server` 1/1 replicas ready/available/updated; pod Running                                       |

The transient test pod in `default` was deleted. Net effect: the intended client reaches the server on port 80, while all other ingress to the server — other pods in the namespace, other namespaces, and other ports — remains blocked as before.
