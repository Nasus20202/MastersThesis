# Agent Instructions

This is a research project. The human researcher owns the methodology and coordinates the work.

## Workflow

1. Start from a ticket and move it to **In Progress**.
2. Analyze the task, relevant sources and open decisions. Discuss them with the human before implementation.
3. Implement only the agreed scope.
4. Move to **In Review** and review the code and results against the ticket.
5. If changes are requested, return to **In Progress** and address them.
6. After code approval, update only documentation agreed with the human.
7. Review and approve the documentation.
8. Merge only after code and required documentation are approved.

## Guardrails

- Keep work concise, research-focused and human-readable. If the human cannot understand why something exists, it is not ready.
- Prefer the smallest clear solution. Do not add speculative abstractions, structure, configuration or documentation.
- Work only on the requested ticket; avoid unrelated cleanup and future-proofing.
- Do not change research decisions, frozen data, sources, conditions or metrics without explicit human approval.
- Never invent sources, citations, ground truth or results. Record sources used, including a stable revision or version when possible.
- Keep benchmark and training data traceable to source material.
- Preserve raw experimental results; negative and inconclusive results are valid.
- Document only what helps the research, reproducibility or thesis writing. Do not restate code, repository structure or ticket history.
- Avoid boilerplate, filler, repeated explanations and generic "best practices" text.

When finished, briefly report what changed, what was verified and anything unresolved.
