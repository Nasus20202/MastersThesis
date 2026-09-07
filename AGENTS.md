# Agent Instructions

This is a research project. The human researcher owns the methodology and coordinates the work.

## Workflow

1. Start from a ticket and move it to **In Progress**.
2. Analyze the task and discuss research or design decisions with the human before implementation.
3. Implement only the agreed scope.
4. Move to **In Review** and review the code and results.
5. If changes are requested, return to **In Progress** and address them.
6. After code approval, update only documentation explicitly agreed with the human.
7. Review and approve the documentation. Do not add documentation just to satisfy the workflow.
8. Merge only after code and required documentation are approved.

- Work only on the requested task.
- Keep changes small and easy to review.
- Prefer the smallest clear solution; avoid speculative abstractions, structure and documentation.
- Do not change research decisions, frozen data, sources, conditions or metrics without explicit human approval.
- Never invent sources, citations, ground truth or results.
- Keep benchmark and training data traceable to source material.
- Preserve raw experimental results; negative and inconclusive results are valid.
- Write documentation only when it helps the research, reproducibility or thesis writing.
- Describe tasks by the intended outcome and research value, not implementation steps.

When finished, briefly report what changed, what was verified and anything unresolved.
