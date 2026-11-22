---
attempt: finetune/datasets/continuation/qwen35-9b/runs/run-2026-11-21-22-42-05-380Z/networkpolicy-egress-dns/prompt/002.json
verdict: accepted
reason: Egress limited to DNS and app=dep on port 80; report discloses the DNS scope and matches the fetched page.
---
# networkpolicy-egress-dns · Qwen3.5-9B-Q4_K_M · run-2026-11-21-22-42-05-380Z · attempt 2

## Task

The application in the egress namespace cannot reach its dependency, and service name resolution from the application fails. Restore the application's required outbound access without allowing all egress.

Documented in:
- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Default policies > Default deny all egress traffic

## Inherited stop

The first 15 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 0
stdout:
"Running"
stderr:

Its writes:

1. `kubectl create networkpolicy allow-app-egress --namespace egress --pod-selector app: app --pod-to-pod=false --policy-types Egress`
2. `kubectl apply -f -` with a YAML content block to define the policy clearly.`
3. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: allow-app-egress\n namespace: egress\nspec:\n podSelector:\n matchLabels:\n app: app\n policyTypes:\n - Egress\n # By defining policyType: Egress and providing no specific egress rules,\n # we allow all egress traffic to the selected pods, overriding the\n # default-deny-egress policy.\n", "path": "policy.yaml"}`
4. `kubectl apply -f policy.yaml -n egress`

## Teacher's write commands

1. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: allow-app-egress\n namespace: egress\nspec:\n podSelector:\n matchLabels:\n app: app\n policyTypes:\n - Egress\n egress:\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - port: 80\n protocol: TCP\n", "path": "policy.yaml"}`
2. `kubectl apply -f policy.yaml -n egress`
3. `kubectl delete networkpolicy allow-app-egress -n egress`
4. `kubectl apply -f policy.yaml -n egress`
5. `write_file {"content": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\nmetadata:\n name: allow-app-egress\n namespace: egress\nspec:\n podSelector:\n matchLabels:\n app: app\n policyTypes:\n - Egress\n egress:\n # Allow DNS queries to resolve service names (UDP/TCP port 53)\n - ports:\n - port: 53\n protocol: UDP\n - port: 53\n protocol: TCP\n # Allow traffic to the dependency service on port 80\n - to:\n - podSelector:\n matchLabels:\n app: dep\n ports:\n - port: 80\n protocol: TCP\n", "path": "policy.yaml"}`
6. `kubectl apply -f policy.yaml -n egress`

## Final report

The issue has been resolved. The problem was a restrictive `default-deny-egress` NetworkPolicy blocking all outbound traffic from pods in the `egress` namespace, including the application pod (`app`) and its dependency service (`dep`).

**Changes Made:**
Created a new NetworkPolicy (`allow-app-egress`) applied specifically to the `app` pod that permits only required egress traffic:
1. **DNS (port 53, UDP/TCP)**: Allows service name resolution by permitting DNS queries to any destination.
2. **Dependency Service (port 80, TCP)**: Allows traffic from the `app` pod to pods labeled `app: dep`.

**Verification of Original Symptom:**
- Connectivity test performed: `wget -qO- --timeout=5 http://dep:80` from within the application pod.
- **Result**: Successfully retrieved HTML response (`<!DOCTYPE html>...`) from the nginx service on the dependency pod. Service name resolution and network connectivity are now functional.

**Constraints Satisfied:**
The fix does not allow all egress; it restricts traffic to only the specific ports required (DNS for resolution, HTTP for the dependency), leaving other pods in the namespace subject to the default deny policy.
