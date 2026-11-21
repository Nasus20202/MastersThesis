---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/networkpolicy-cross-namespace/prompt/005.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# networkpolicy-cross-namespace · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 5

## Task

A client in one namespace can no longer reach the application in another namespace. Restore the intended cross-namespace access while keeping other namespaces blocked.

Documented in:

- content/en/docs/concepts/services-networking/network-policies.md: Behavior of `to` and `from` selectors, Targeting multiple namespaces by label, Targeting a Namespace by its name

## Write commands

1. `kubectl patch networkpolicy allow-clients -n server-ns --type=json -p='[{"op":"replace","path":"/spec/ingress/0/from","value":[{"namespaceSelector":{"matchLabels":{"kubernetes.io/metadata.name":"client-ns"}},"podSelector":{"matchLabels":{"role":"trusted"}}}]}]'`

## Final report

Fixed. Root cause and verification:

**Diagnosis**

- `NetworkPolicy allow-clients` in `server-ns` selects `app=server` on port 80/TCP for Ingress, with a single `from` peer containing only a `podSelector: role=trusted`.
- In NetworkPolicy semantics, a peer with _only_ a `podSelector` means "pods with this label **in the policy's own namespace**". `server-ns` has no `role=trusted` pods, so no source could ever match — client traffic from every namespace was dropped.
- Both `client-ns/allowed-client` and `other-ns/blocked-client` carry `role=trusted`, so the fix had to add namespace scoping, not just a bare `namespaceSelector`.

**Change** (one field, the Ingress `from` peer; podSelector, port, and policyTypes untouched)

```yaml
ingress:
  - from:
      - namespaceSelector:
          matchLabels:
            kubernetes.io/metadata.name: client-ns
        podSelector:
          matchLabels:
            role: trusted
    ports:
      - { port: 80, protocol: TCP }
```

A `namespaceSelector` and `podSelector` in the same peer are ANDed, so only `role=trusted` pods inside `client-ns` are admitted.

**Verified by direct observation (not just patch acceptance)**

- `client-ns` → `http://server.server-ns.svc.cluster.local/` → HTTP 200, body `Welcome to nginx!` (and the same via the Service ClusterIP `10.96.46.112`). Pre-fix this timed out.
- `other-ns` → same URL → still `download timed out` on 3 consecutive attempts (connection dropped), so other namespaces remain blocked.
- Server workload unaffected: `deploy/server` 1/1 ready, up-to-date 1, available 1; pod `server-5c67b5b859-dbh9q` Running, 0 restarts, no other resources modified.
