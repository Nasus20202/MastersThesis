# Kubernetes Policies

Admission is the gate between an accepted write and a persisted object. Mutating admission runs first and may change the object (injecting defaults, sidecars, or security settings); validating admission then accepts or rejects it. Built-in controls, Pod Security admission, and LimitRange/ResourceQuota enforcement all act here, alongside custom validating and mutating webhooks (see `extensions.md`).

Pod Security admission enforces a Pod Security Standard per namespace (`enforce`, `audit`, `warn` modes at increasing strictness). A rejection names the violated control and the offending field; the Pod object is never created, so workload children counts diverge from expectations.

LimitRange constrains individual objects (defaults, minimums, maximums for requests, limits, or ratios), while ResourceQuota constrains namespace aggregates (total requests/limits, object counts). Quota counts at creation: an object that exceeds the remaining budget is rejected before scheduling or provisioning begins.

NetworkPolicy and RBAC (see `networking.md`, `authorization.md`) are also policy, but they govern runtime traffic and API access rather than admission. Disabling or widening any policy removes a guardrail for every current and future object it selects, not just for the workload under investigation.
