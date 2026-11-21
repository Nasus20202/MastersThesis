---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-15-55-41-714Z/resourcequota-blocked/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# resourcequota-blocked · Qwen3.5-9B-Q4_K_M · run-2026-11-21-15-55-41-714Z · attempt 3

## Task

The application rollout is not completing because the namespace ResourceQuota blocks additional Pods. Restore the application to three healthy replicas without removing the ResourceQuota.

Documented in:

- content/en/docs/concepts/policy/resource-quotas.md: How Kubernetes ResourceQuotas work

## Write commands

1. `kubectl edit resourcequota app-quota -n default`
2. `kubectl apply -f /tmp/quota.yaml -n default`

## Final report

## Summary

**Problem Identified:**
The application deployment was unable to scale to 3 replicas because the `ResourceQuota` named `app-quota` in the `default` namespace limited pods to 2 (`pods: 2/2`). The deployment controller attempted to create a third pod but received "exceeded quota" errors.

**Fix Applied:**
Edited the existing ResourceQuota `app-quota` to increase the `pods` hard limit from `"2"` to `"3"`. This preserves the ResourceQuota object while allowing the additional replica.

**Verification:**

| Check                 | Before Fix | After Fix            |
| --------------------- | ---------- | -------------------- |
| Deployment Ready      | 2/3        | **3/3** ✓            |
| Deployment Available  | 2          | **3** ✓              |
| Pods Running          | 2          | **3** ✓              |
| ResourceQuota Limit   | pods: 2    | **pods: 3** ✓        |
| ResourceQuota Removed | N/A        | **No** (preserved) ✓ |

**Final State:**

- All 3 application pods are healthy (READY 1/1, STATUS Running):
  - `app-575b7496bc-8sn88`
  - `app-575b7496bc-gt5gf` (newly created)
  - `app-575b7496bc-zl7kt`
- ResourceQuota `app-quota` remains in place with limit of 3 pods, currently using 3/3.
