---
name: troubleshooting
description: Apply a general evidence-first workflow to diagnose, repair and verify problems in an interactive sandbox.
metadata:
  version: "1.0"
  scope: diagnosis and verification
---

# Troubleshooting

Treat the task as a claim about an observable end state. Work from evidence
and keep diagnosis separate from repair.

1. Define what healthy or complete means in terms that can be observed.
2. Establish the current state with the smallest useful set of focused checks.
3. Form a hypothesis that explains the observations, then test it with a
   discriminating check rather than collecting unrelated output.
4. Change the smallest authoritative object or setting that addresses the
   evidence. Preserve unrelated configuration and access scope.
5. Allow bounded convergence time, then repeat fresh checks of the end state.
   If the result is not reached, return to evidence and revise the hypothesis.
6. Finish with a concise account of observations, changes and verification.
   Do not claim success from intent, a successful mutation command or a stale
   observation.

When several components interact, trace the dependency and ownership edges
between them. Prefer an explanation supported by multiple independent
observations, and stop when the task is verified or the execution budget ends.
