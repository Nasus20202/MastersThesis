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
