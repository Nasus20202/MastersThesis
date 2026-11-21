---
attempt: finetune/datasets/qwen35-9b/runs/run-2026-11-21-17-25-40-589Z/rollback-bad-release/prompt/003.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# rollback-bad-release · Qwen3.5-9B-Q4_K_M · run-2026-11-21-17-25-40-589Z · attempt 3

## Task

A recent application release is not completing its rollout and the application is running with fewer available replicas than intended. Restore the application to a healthy state.

Documented in:

- content/en/docs/concepts/workloads/controllers/deployment.md: Rolling Back a Deployment

## Write commands

1. `kubectl rollout undo deployment/app -n default --to-revision=1 2>/dev/null`

## Final report

## Summary

**Root Cause:** The application deployment was configured with a broken image/container specification in revision 2. The pod template had `busybox:1.37.0` with command `sh -c exit 1`, which immediately caused containers to crash with exit code 1. This prevented the new replica set from scaling properly.

**Fix Applied:** Rolled back the deployment to revision 1 using:

```bash
kubectl rollout undo deployment/app -n default --to-revision=1
```

This restored the correct configuration (`nginx:1.31.5`) that was working in the old replica set.

**Verified Outcome:**

| Check                | Expected     | Observed     | Status |
| -------------------- | ------------ | ------------ | ------ |
| Ready replicas       | 3/3          | 3/3          | ✓      |
| Available replicas   | 3/3          | 3/3          | ✓      |
| Unavailable replicas | 0            | 0            | ✓      |
| Image used           | nginx:1.31.5 | nginx:1.31.5 | ✓      |

**Final State:** The deployment `app` in namespace `default` is now fully healthy with all 3 replicas running, ready, and available using the correct `nginx:1.31.5` image.
