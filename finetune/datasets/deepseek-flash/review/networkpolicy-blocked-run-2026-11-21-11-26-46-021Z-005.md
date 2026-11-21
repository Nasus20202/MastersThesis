---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-blocked/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-blocked · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A required client can no longer reach the application server in the policy namespace. Restore the intended access while leaving all other traffic to the server blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: The NetworkPolicy resource, Default policies > Default deny all ingress traffic

## Write commands

1. `kubectl apply -f -`
2. `kubectl run tmp-probe-$RANDOM --image=busybox:1.37.0 --restart=Never -n default --command -- sh -c 'wget -q -T 4 -O /dev/null http://server.policy.svc.cluster.local/`
3. `kubectl logs -n default -l run --tail=1 2>/dev/null`
4. `kubectl delete pod -n default tmp-probe-15629 --wait=false 2>`

## Final report

## Root cause

The `policy` namespace contained only a default-deny ingress NetworkPolicy (`podSelector: {}`, `policyTypes: [Ingress]`) and no allow rule for the server. As a result, **all** ingress to the `server` pod — including traffic from the required client — was dropped. Baseline observations showed:

- `allowed-client` → `server`: `wget: download timed out`
- `blocked-client` → `server`: `wget: download timed out`

Both clients run `busybox:1.37.0`; the server is `nginx:1.31.5`, selected by `app=server`, exposed via `Service/server:80`. The clients are distinguished only by labels: `allowed-client` has `role=allowed`, `blocked-client` has `role=blocked`.

## Change

Added one additive ingress allow policy (left `default-deny-ingress` untouched so all other traffic stays blocked):

```yaml
kind: NetworkPolicy
metadata: { name: allow-allowed-client-to-server, namespace: policy }
spec:
  podSelector: { matchLabels: { app: server } } # only the server pod
  policyTypes: [Ingress]
  ingress:
    - from: [{ podSelector: { matchLabels: { role: allowed } } }] # only the required client
      ports: [{ protocol: TCP, port: 80 }] # only the app port
```

## Verification (fresh observations)

| Check                                                                 | Result                                                                                               |
| --------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| Required client → Service (`http://server.policy.svc.cluster.local/`) | **Success**, `wget_exit=0`, returns `<!DOCTYPE html>... <title>Welcome to nginx!</title>`            |
| Blocked client → Service                                              | Still blocked: `wget_exit=1`, `download timed out`                                                   |
| Blocked client → server pod IP `192.168.94.134`                       | Still blocked: `wget_exit=1`                                                                         |
| Pod in an unrelated namespace (`default`, unlabeled) → Service        | Still blocked: `probe_exit=1`, `download timed out` (test pod removed)                               |
| Server controller health                                              | `deploy/server` replicas desired 1 / updated 1 / ready 1 / available 1; RS `server-5c67b5b859` 1/1/1 |

The required client's access is restored, every other path to the server (other label/role, other namespace, direct pod IP) remains denied, and the server workload itself was not modified.
