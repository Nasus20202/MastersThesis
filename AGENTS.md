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

## Approval gate

Research and technical decisions must be explicitly approved by the human researcher before they are implemented, executed, frozen, or recorded as approved. First present the proposed decision, affected scope, rationale and alternatives, then wait for explicit approval in the conversation. Do not infer approval from a ticket, prior discussion, an implementation request, or silence. Exploratory checks may inspect the current state, but must not change the agreed methodology or document an unapproved decision as accepted. If an implementation request contains an unstated research or technical choice, stop and ask for approval before making that choice.

## Useful commands

Run `make` from the repository root to list available commands. Before committing, run
`make test`, `make format`, and `make lint`.

## Documentation

- Update documentation when an approved research or technical decision changes the content it describes.
- Record every research and technical decision, including changes and reversals, in [docs/decision-log.md](docs/decision-log.md) with a date, status, rationale or reason and a reference to the affected document.
- Keep [docs/decision-log.md](docs/decision-log.md) as the single source of truth for open decisions. Do not repeat the same open-decision list in other documentation files; link to the decision log instead.
- Keep subject-specific details in their owning documentation file and avoid copying full sections between files.
- Update [docs/README.md](docs/README.md) when documentation files are added, removed or renamed.
- Documentation changes remain human-gated: do not change research decisions, frozen data, sources, conditions or metrics without explicit approval.

## Guardrails

- Keep work concise, research-focused and human-readable. If the human cannot understand why something exists, it is not ready.
- Prefer the smallest clear solution. Do not add speculative abstractions, structure, configuration or documentation.
- Work only on the requested ticket; avoid unrelated cleanup and future-proofing.
- Do not change research decisions, frozen data, sources, conditions or metrics without explicit human approval.
- Never invent sources, citations, ground truth or results. Record research and technical sources used, including stable revision/version and location when possible, so they can be cited later.
- Keep benchmark and training data traceable to source material.
- Preserve raw experimental results; negative and inconclusive results are valid.
- Document only what helps the research, reproducibility or thesis writing. Do not restate code, repository structure or ticket history.
- Avoid boilerplate, filler, repeated explanations and generic "best practices" text.
- When finished, briefly report what changed, what was verified and anything unresolved.
