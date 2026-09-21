---
name: troubleshooting
description: General diagnosis, repair, and verification workflow for structured troubleshooting.
---

# Troubleshooting

Diagnose from observed evidence, repair the source of the faulty state, and verify recovery before reporting success.

## Diagnose

Start from the resources and context provided by the task. If information is missing, discover it with focused queries and expand only when the evidence requires it.

Inspect the current state, relevant events or logs, and resource ownership. Trace the symptom to the component or configuration responsible for producing the faulty state rather than treating only its visible effects. Before repair, identify the requested outcome and any constraints that must remain true.

Load supporting guidance:

- For Kubernetes tasks, load the `kubernetes` skill before diagnosing resource behavior.
- After identifying the affected subsystem or subsystems, load each relevant focused reference before choosing a repair. Load references when their subject applies to the observed problem.
- Use `kubectl` when command syntax, resource addressing, output, or operation behavior needs guidance.

## Repair

Prefer the smallest change that addresses the diagnosed cause.

Modify the resource or configuration that owns the faulty state rather than transient output derived from it. Preserve unrelated configuration and existing constraints. Avoid speculative or unrelated changes.

Make one focused change at a time and verify its effect before making another.

## Verify

Use fresh observations after the repair.

Confirm both that the change took effect and that the requested outcome has recovered. Do not treat a successful command or accepted configuration change as proof of recovery.

Use bounded waits where available. If a wait times out, inspect fresh state and continue from that evidence.

If verification fails, use the new evidence to continue diagnosis, make a focused correction when supported, and verify again. Continue until the requested outcome is verified or no further progress is possible. Only when progress is no longer possible, report the blocker and unmet outcome; do not claim success.
