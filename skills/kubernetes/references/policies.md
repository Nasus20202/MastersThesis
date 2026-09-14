# Kubernetes policies and disruption controls

Policies act at different points in the lifecycle. Admission policies can
default or reject an object, namespace policies can constrain creation, and
runtime policies can affect traffic or voluntary disruption. Identify the
policy, selector and direction before changing it.

- For an API rejection, use the complete error and the object's namespace and
  effective fields to distinguish schema, admission, quota and security-policy
  causes.
- A `ResourceQuota` or `LimitRange` affects admission and defaults; use
  `resources.md` to trace requests, limits and effective values.
- A `NetworkPolicy` affects selected Pod traffic; use `networking.md` to trace
  source, destination and ingress/egress behavior.
- A `PodDisruptionBudget` limits voluntary disruptions such as eviction or
  drain. It does not make an unready Pod ready and is not a general availability
  guarantee.
- For mutating or validating webhooks and custom policies, check reported
  failures, selectors, scope and referenced configuration. Do not disable or
  delete a policy as a first repair.

Preserve the policy's intended scope and least privilege. Verify both the
admission result and the resulting workload behavior after a change.
