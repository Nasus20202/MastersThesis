# Agent Instructions

This is a research project. The human researcher owns the methodology and coordinates the work.

## Workflow

1. Start from an issue ticket on GitHub and move it to **In Progress**.
2. Analyze the task, relevant sources and open decisions. Discuss them with the human before implementation.
3. Implement only the agreed scope.
4. Move to **In Review** and review the code and results against the ticket.
5. If changes are requested, return to **In Progress** and address them.
6. After code approval, update only documentation agreed with the human.
7. Review and approve the documentation.
8. Merge only after code and required documentation are approved.

## Commits

- Use Conventional Commits (`feat:`, `fix:`, `refactor:`, `chore:`, ...) with an imperative summary.
- One logical change per commit.
- Do not commit unless the human asks.

## Approval gate

Research and technical decisions must be explicitly approved by the human researcher before they are implemented, executed, frozen, or recorded as approved. First present the proposed decision, affected scope, rationale and alternatives, then wait for explicit approval in the conversation. Do not infer approval from a ticket, prior discussion, an implementation request, or silence. Exploratory checks may inspect the current state, but must not change the agreed methodology or document an unapproved decision as accepted. If an implementation request contains an unstated research or technical choice, stop and ask for approval before making that choice.

## Useful commands

Run `make` from the repository root to list available commands. Before committing run `make format` and `make check` (tests, lint and corpus validation). When scenarios change, also run `make benchmark-validate`.

## Documentation

- Update documentation when an approved research or technical decision changes the content it describes.
- Record every research and technical decision, including changes and reversals, in [docs/decision-log.md](docs/decision-log.md). Use a sequential `D-NNN` identifier, an ISO `YYYY-MM-DD` date, one of the four status values, a rationale or reason and a reference to the affected document; one row per decision.
- Keep [docs/decision-log.md](docs/decision-log.md) as the single source of truth for open decisions. Do not repeat the same open-decision list in other documentation files; link to the decision log instead.
- All records in [docs/decision-log.md](docs/decision-log.md) must be human-reviewed and approved before being added. Do not modify the file without explicit approval from the human researcher.
- Only decisions not yet merged into the default branch can and should be modified. Once a decision is merged, it is frozen and cannot be changed without explicit approval.
- Keep subject-specific details in their owning documentation file and avoid copying full sections between files.
- Don't save non human-readable or non-research-focused content in documentation files. If it is not clear why something exists, it is not ready to be documented.
- Write research documents around the question, method, results, and limitations. Include commands and hashes only when they help reproduce or interpret an experiment; keep links to the raw evidence.
- Keep [docs/README.md](docs/README.md) as the routing index: one line per document with a short description and a link; update it when documentation files are added, removed or renamed.
- Documentation changes remain human-gated: do not change research decisions, frozen data, sources, conditions or metrics without explicit approval.

### Document ownership

- [research-design.md](docs/research-design.md) — research goal, questions, scope, adaptation conditions, outcomes, comparison principles and the development/final-evaluation split.
- [benchmark.md](docs/benchmark.md) — scenario contract and lifecycle, condition capabilities, execution limits, model visibility, verification and scoring, repeated runs and acceptance.
- [tech-stack.md](docs/tech-stack.md) — runner, model serving and multi-token prediction, `kind`, registry caches, sandbox and setup container, adaptation components and provenance.
- [scenarios.md](docs/scenarios.md) — evaluator-only incident catalogue: states, faults, expected repairs, grading, difficulty and cluster profiles.
- [glossary.md](docs/glossary.md) — one line per term: definition and link to the owning document or an external source.
- [sources.md](docs/sources.md) — external sources and the frozen corpus, with purpose, status and revision or location.
- [decision-log.md](docs/decision-log.md) — dated decisions with status, rationale and reference; the only place open decisions live.
- [`docs/research/<topic>/README.md`](docs/research/) — one study: question, method and configuration, results, limitations and links to its raw evidence.
- [benchmark/README.md](benchmark/README.md) — benchmark module: local runner commands, logging and scenario configuration.
- [docs/README.md](docs/README.md) — routing index: one line and link per document.

Keep detail in its owning document and link instead of copying it. Update the owning document when its trigger changes: scenario edits update [scenarios.md](docs/scenarios.md) and [benchmark.md](docs/benchmark.md); model, runtime or `kind` changes update [tech-stack.md](docs/tech-stack.md) and [sources.md](docs/sources.md); new terminology updates [glossary.md](docs/glossary.md); scope or protocol changes update [research-design.md](docs/research-design.md).

`benchmark/results/` and `benchmark/models/` are gitignored; copy accepted raw evidence under `docs/research/<topic>/` so it is committed with its study.

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
