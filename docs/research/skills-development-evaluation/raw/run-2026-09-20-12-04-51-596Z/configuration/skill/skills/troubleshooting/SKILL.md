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
- After identifying the affected subsystem or subsystems, load every Kubernetes reference that covers them before choosing a repair. If more than one subsystem is involved, load the relevant reference for each; do not load unrelated references.
- Before the first Kubernetes mutation, load `kubectl` for command and operation guidance.

Do not skip a required skill or reference because the problem seems familiar.

## Repair

Prefer the smallest change that addresses the diagnosed cause.

Modify the resource or configuration that owns the faulty state rather than transient output derived from it. Preserve unrelated configuration and existing constraints. Avoid speculative or unrelated changes.

Make one focused change at a time and verify its effect before making another.

Use noninteractive commands and bounded operations. Avoid interactive editors and open-ended watch streams.

## Verify

Use fresh observations after the repair.

Confirm that the change took effect, the requested outcome has recovered, and every stated constraint still holds. Include prohibited actions or access in the checks when the task names them.

For controller-managed resources, verify that the current desired state has converged. Check that updated and ready instances match the requested state; ready instances from an older revision do not prove recovery.

Report success only when fresh evidence passes every required check. If any check fails or remains uncertain, continue diagnosis and repair when possible; otherwise report that the task remains incomplete and name the unmet or uncertain check.

Use bounded waits where available. If a wait times out, inspect fresh state and continue from that evidence.

If verification fails, use the new evidence to continue diagnosis, make a focused correction when supported, and verify again. Continue until the requested outcome is verified or no further progress is possible. Only when progress is no longer possible, report the blocker and unmet outcome; do not claim success.
