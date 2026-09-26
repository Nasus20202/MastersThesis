---
name: troubleshooting
description: General diagnosis, repair, and verification workflow for structured troubleshooting.
---

# Troubleshooting

Diagnose from observed evidence, repair the source of the faulty state, and verify recovery before reporting success.

## Diagnose

Start from the resources and context provided by the task. If information is missing, discover it with focused queries and expand only when the evidence requires it.

Inspect the current state, relevant events or logs, and resource ownership. Trace the symptom to the component or configuration responsible for producing the faulty state rather than treating only its visible effects. Before repair, identify the requested outcome and any constraints that must remain true.

Load domain guidance before acting:

- For a Kubernetes task, load the `kubernetes` skill before the first Bash call.
- After identifying the affected subsystem or subsystems, inspect the References list in the loaded skill and load every reference that covers them before choosing a repair. If more than one subsystem is involved, load the relevant reference for each; do not load unrelated references.
- For Kubernetes references, call `load_reference` with `skill: "kubernetes"` and the exact listed filename as `reference`. The reference topic is not the skill name.
- Before the first Kubernetes mutation, load `kubectl` for command and operation guidance. If unsure whether a command changes cluster state, load it first.

Do not skip a required skill or reference because the problem seems familiar.

## Subsystem skills

The repair procedure for a subsystem lives in its own skill. Before you change a subsystem, you must call `load_skill` for its skill and follow it. Treat the mapping below as the step that selects your repair procedure: do not run a repair command while a matching subsystem skill is still unloaded. Load several in one response when a symptom spans areas.

When the symptom matches one or more rows, your next action is `load_skill` for each matching skill; only then plan the repair.

- Pending, unschedulable, node placement, taints, affinity: `scheduling`
- PersistentVolumeClaim, volumes, mounts, read-only root filesystem: `storage`
- rollouts, probes, crash loops, StatefulSets, Jobs, headless Services: `workloads`
- ServiceAccount, RBAC, forbidden or 403, token automount: `authorization`
- Service reachability, NetworkPolicy, DNS, Ingress, Gateway: `networking`
- ConfigMap, Secret, environment or mounted configuration: `configuration`
- LimitRange, ResourceQuota, admission rejection: `governance`
- Pod Security, securityContext, runAsNonRoot, read-only filesystem: `security`
- CPU or memory requests and limits, OOMKilled, QoS: `resources`

## Symptom checks

Use the symptom to choose a discriminating check before changing anything. Prefer a check that separates the candidate causes instead of guessing.

- **Service unreachable.** Confirm whether the Service has ready endpoints; compare `spec.ports[].targetPort` with the container's `containerPort` and port name; compare the Service `selector` with the Pod labels; then check NetworkPolicy and, if needed, test the request from a client Pod. Ready endpoints and healthy Pods do not prove the Service routes to the right port.
- **Pod not Ready or restarting.** Read the container waiting or terminated reason. `CreateContainerConfigError` points at a missing or wrong ConfigMap or Secret reference; `ImagePullBackOff` at the image reference or pull secret; `OOMKilled` at the memory limit and working set; repeated liveness kills at the liveness probe; argument or command errors at the container command. Fix the owning controller's Pod template, not the generated Pod.
- **Rollout not completing.** Compare `spec.replicas` with updated, ready, and available replicas, and read ReplicaSet or ReplicaFailure events. Check for immutable fields such as a Deployment `selector`, unsatisfiable scheduling constraints, or a new Pod that cannot start.
- **Forbidden or access denied.** Identify the requesting ServiceAccount, verb, resource, and namespace, and test the exact permission with `kubectl auth can-i`. Grant only what the workload uses by correcting the subject or the scoped Role rather than broadening access.
- **Configuration value not applied.** Environment variables from a ConfigMap or Secret are read at container start; adding or changing the source does not update a running container. Trigger a new rollout after the change and verify the new Pods report the updated value.
- **Pod rejected by admission.** Read the rejection reason: Pod Security field violations, missing LimitRange defaults, or ResourceQuota limits. Make the object compliant rather than weakening the namespace policy.
- **PersistentVolumeClaim not binding.** Compare the claim's storage class, access modes, capacity, and selector with available volumes; a claim can be bound while later attachment or mounting still fails.

## Repair

Prefer the smallest change that addresses the diagnosed cause.

Modify the resource or configuration that owns the faulty state rather than transient output derived from it. Preserve unrelated configuration and existing constraints. Avoid speculative or unrelated changes.

Make one focused change at a time and verify its effect before making another. Run the change, any wait, and the verification as separate tool calls so each result can guide the next step.

Use noninteractive commands and bounded operations. Do not invoke interactive editors such as `kubectl edit`. Avoid open-ended watch streams. Keep each wait shorter than the outer command tool's deadline; if that deadline is 60 seconds, use a wait of about 30 seconds or less.

## Verify

Use fresh observations after the repair.

Confirm that the change took effect, the requested outcome has recovered, and every stated constraint still holds. Include prohibited actions or access in the checks when the task names them.

For controller-managed resources, verify that the current desired state has converged. Compare the requested replica count with the controller's updated and ready replica counts; all must match. A successful wait or some ready instances do not prove recovery when counts still differ.

For a Deployment, compare `.spec.replicas` with `.status.updatedReplicas`, `.status.readyReplicas`, and `.status.availableReplicas` when availability is part of the requested outcome. A `READY` summary alone does not prove rollout completion if fewer replicas are updated. Query the missing fields directly rather than inferring them.

Before ending, make a fresh observation for each required check and include the relevant observed values in the final response. Do not finish with only a diagnosis or a plan. Report `Outcome: complete` only when the evidence passes every required check. If a check fails, reports a mismatch, or remains uncertain, continue diagnosis and repair when possible; if no further progress is possible, report `Outcome: incomplete`, name the unmet or uncertain check, and do not describe the task as successful or recovered.

Use bounded waits where available. If a wait times out, inspect fresh state and continue from that evidence.

If verification fails, use the new evidence to continue diagnosis, make a focused correction when supported, and verify again. Continue until the requested outcome is verified or no further progress is possible. Only when progress is no longer possible, report the blocker and unmet outcome; do not claim success.
