---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/networkpolicy-egress-dns/prompt/003.json
verdict: accepted
reason: 'Recreates the egress policy with DNS, dep pods, and the dep ClusterIP, accurately verified by a successful dep fetch and blocked unrelated service.'
---
# networkpolicy-egress-dns · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 3

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:
- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Inherited stop

The first 35 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 1
stdout:

stderr:
Error from server (NotFound): pods "app" not found

Its writes:

1. `kubectl edit networkpolicy default-deny-egress -n egress`
2. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n annotations:\n kubectl.kubernetes.io/last-applied-configuration: |\n {\"apiVersion\":\"networking.k8s.io/v1\",\"kind\":\"NetworkPolicy\",\"metadata\":{\"annotations\":{},\"name\":\"default-deny-egress\",\"namespace\":\"egress\"},\"spec\":{\"podSelector\":{},\"policyTypes\":[\"Egress\"]}}\n creationTimestamp: \"2026-10-10T01:07:50Z\"\n generation: 1\n name: default-deny-egress\n namespace: egress\n resourceVersion: \"1019\"\n uid: 66ebe8d6-bc29-4ebf-a7b5-d9165e24a953\nspec:\n podSelector: {}\n policyTypes:\n - Egress\n egress:\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - protocol: TCP\n port: 80\n", "path": "policy.yaml"}`
3. `kubectl apply -f policy.yaml`
4. `kubectl patch networkpolicy default-deny-egress -n egress --type='json' --patch='[{"op": "add", "path": "/spec/egress", "value": [{"to": [{"podSelector": {"matchLabels": {"app": "dep"}}}]}, "ports": [{"protocol": "TCP", "port": 80}]}]'`
5. `kubectl patch networkpolicy default-deny-egress -n egress --type='merge' --type='yaml'`
6. `kubectl patch networkpolicy default-deny-egress -n egress --patch='{"spec":{"egress":[{"to":[{"podSelector":{"matchLabels":{"app":"dep"}}]}, "ports":[{"protocol":"TCP","port":80}]}]}}'`
7. `kubectl apply --force -f policy.yaml`
8. `kubectl patch networkpolicy default-deny-egress -n egress --type='merge' --type='yaml'`

## Teacher's write commands

1. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: default-deny-egress\n namespace: egress\nspec:\n podSelector: {}\n policyTypes:\n - Egress\n egress:\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - protocol: TCP\n port: 80\n", "path": "policy.yaml"}`
2. `kubectl delete networkpolicy default-deny-egress -n egress`
3. `kubectl apply -f policy.yaml`
4. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: default-deny-egress\n namespace: egress\nspec:\n podSelector: {}\n policyTypes:\n - Egress\n egress:\n - to:\n - namespaceSelector:\n matchLabels:\n kubernetes.io/metadata.name: kube-system\n ports:\n - protocol: UDP\n port: 53\n - protocol: TCP\n port: 53\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - protocol: TCP\n port: 80\n - to:\n - ipBlock:\n cidr: 10.96.140.87/32\n ports:\n - protocol: TCP\n port: 80\n", "path": "policy.yaml"}`
5. `kubectl apply -f policy.yaml`

## Final report

The issue was caused by the `default-deny-egress` NetworkPolicy in the `egress` namespace, which blocked all outbound traffic from all pods. This prevented the application from resolving DNS and reaching its dependency.

I updated the persistent `default-deny-egress` NetworkPolicy to allow only the specific required outbound traffic without enabling all egress:
1. **DNS resolution**: Allowed egress to the `kube-system` namespace on UDP/TCP port 53.
2. **Dependency access**: Allowed egress to the `dep` pods (label `app: dep`) on TCP port 80.
3. **Service routing**: Allowed egress to the `dep` Service ClusterIP (`10.96.140.87`) on TCP port 80 (necessary for CNIs that enforce policy before DNAT).

**Verification:**
- **Dependency reachability & DNS**: Running `wget http://dep` from the `app` pod returned the expected nginx welcome page (exit code 0), confirming both name resolution and connectivity work.
- **Egress restriction maintained**: Running `wget http://other` from the `app` pod timed out (exit code 1), confirming that unrelated egress remains blocked and the policy did not open all outbound traffic.
