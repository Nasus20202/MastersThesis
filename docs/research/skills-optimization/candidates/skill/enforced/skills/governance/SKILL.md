---
name: governance
description: Repairing admission rejections from LimitRange and ResourceQuota without weakening the namespace policy.
---

# Governance

Admission rejects an object before it is created. Read the exact rejection reason and make the object compliant; do not raise or delete the policy to clear the error.

## LimitRange

- A LimitRange sets namespace defaults and bounds. A rejection names the field and the allowed range.
- Set `resources.requests` and `resources.limits` within the minimum and maximum, and keep the ratio implied by the defaults. A container missing requests or limits receives the defaults only if the LimitRange defines them.

## ResourceQuota

- A quota caps the total requested resources and object counts in a namespace. Sum what the namespace already requests and compare with the quota's `used` and `hard` values.
- Reduce or remove only the workload that exceeds the quota, and only if the task allows it; preserve stated limits and unrelated workloads.
- A Pod with no requests may be rejected by a quota that requires them (`requests.cpu`, `requests.memory`).

## Verify

The object is admitted and reaches its expected state (Pod Running/Ready), and the LimitRange and ResourceQuota are unchanged from their original values.
