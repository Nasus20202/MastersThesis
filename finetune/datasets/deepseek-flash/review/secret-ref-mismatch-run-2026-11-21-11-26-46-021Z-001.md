---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/secret-ref-mismatch/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# secret-ref-mismatch · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 1

## Task

A Kubernetes application rollout is not completing. Restore the application to a healthy state.

Documented in:

- content/en/docs/tasks/inject-data-application/distribute-credentials-secure.md: Define container environment variables using Secret data > Define a container environment variable with data from a single Secret

## Write commands

1. `kubectl -n default patch deployment app --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key","value":"password"}]'`

## Final report

The rollout is fixed. Here's what I found and did:

**Diagnosis**

- Deployment `default/app` was at generation 2 with `1 updated / 4 total / 1 unavailable`: the new ReplicaSet `app-c6d597897`'s pod `app-c6d597897-h68ht` was stuck in `CreateContainerConfigError`.
- Pod events showed: `Error: couldn't find key passwd in Secret default/app-secret`.
- The current (revision 1) ReplicaSet `app-849cc7cbf4` — whose 3 pods were running fine — referenced `key: password`, while the new revision's pod template referenced `key: passwd`. Secret `app-secret` only contains the key `password` (`czNjcjN0`).
- So the bad change was in the Deployment's pod template (wrong secret key), not the Secret. I left the Secret untouched.

**Change**

- Patched only the offending field on the Deployment (smallest justified change, no other settings altered):
  `kubectl -n default patch deployment app --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/env/0/valueFrom/secretKeyRef/key","value":"password"}]'`
- `kubectl rollout status deploy/app` → `deployment "app" successfully rolled out`. Because the corrected template matches the previously working revision, the controller reused ReplicaSet `app-849cc7cbf4` and scaled the broken `app-c6d597897` to 0.

**Verified outcome (fresh observations)**

- Deployment: `generation=3 observedGeneration=3`, `desired=3 updated=3 ready=3 available=3 unavailable=<none>` — requested, updated, ready and available counts all match.
- Pods: all three `app-849cc7cbf4-*` pods are `1/1 Running`, `RESTARTS=0`; no pods in `CreateContainerConfigError`.
- Live template now resolves: `secretKeyRef {name: app-secret, key: password}`; Secret still only has key `password`; exec inside a running pod shows `APP_PASSWORD=s3cr3t`, confirming the reference actually resolves at runtime.

Note: unrelated platform components (`coredns`, `kube-proxy` in `kube-system`) show their own readiness issues in this kind cluster; they were not part of the application rollout, are not referenced by the app (no Service/DNS dependency), and I left them unchanged to avoid unrelated modifications.
