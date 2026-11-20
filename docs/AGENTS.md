# Documentation instructions

Rules for everything under `docs/`, in addition to the [repository instructions](../AGENTS.md).

## Rules

- Record every research and technical decision, including changes and reversals, in [decision-log.md](decision-log.md). Use a sequential `D-NNN` identifier, an ISO `YYYY-MM-DD` date, one of the four status values, a rationale or reason and a reference to the affected document; one row per decision. A row states only the decision and its reason; measured results, run details and implementation belong in the owning document.
- Keep [decision-log.md](decision-log.md) as the single source of truth for open decisions. Do not repeat the same open-decision list in other documentation files; link to the decision log instead.
- All records in [decision-log.md](decision-log.md) must be human-reviewed and approved before being added. Do not modify the file without explicit approval from the human researcher.
- Only decisions not yet merged into the default branch can and should be modified. Once a decision is merged, it is frozen and cannot be changed without explicit approval.
- Keep subject-specific details in their owning documentation file and avoid copying full sections between files.
- Don't save non human-readable or non-research-focused content in documentation files. If it is not clear why something exists, it is not ready to be documented.
- Start a new study from [research/TEMPLATE.md](research/TEMPLATE.md). Write research documents around the question, method, results, and limitations. Include commands and hashes only when they help reproduce or interpret an experiment; keep links to the raw evidence.
- The research centres on the final run. Development tries, failed attempts and abandoned paths get a short description in their study (what was tried, what happened, why it was dropped), without raw evidence; remove their sections and files once nothing depends on them.
- Keep [README.md](README.md) as the routing index: one line per document with a short description and a link; update it when documentation files are added, removed or renamed.
- Documentation changes remain human-gated: do not change research decisions, frozen data, sources, conditions or metrics without explicit approval.

## Document ownership

- [research-design.md](research-design.md) — research goal, questions, scope, adaptation conditions, outcomes, comparison principles and the development/final-evaluation split.
- [benchmark.md](benchmark.md) — scenario contract and lifecycle, condition capabilities, execution limits, model visibility, verification and scoring, repeated runs and acceptance.
- [tech-stack.md](tech-stack.md) — runner, model serving and multi-token prediction, `kind`, registry caches, sandbox and setup container, adaptation components and provenance.
- [scenarios.md](scenarios.md) — evaluator-only incident catalogue: states, faults, expected repairs, grading, difficulty and cluster profiles.
- [glossary.md](glossary.md) — one line per term: definition and link to the owning document or an external source.
- [sources.md](sources.md) — external sources and the frozen corpus, with purpose, status and revision or location.
- [decision-log.md](decision-log.md) — dated decisions with status, rationale and reference; the only place open decisions live.
- [`research/<topic>/README.md`](research/) — one study: question, method and configuration, results, limitations and links to its raw evidence.
- [benchmark/README.md](../benchmark/README.md) — benchmark module: local runner commands, logging and scenario configuration.
- [README.md](README.md) — routing index: one line and link per document.

Keep detail in its owning document and link instead of copying it. Update the owning document when its trigger changes: scenario edits update [scenarios.md](scenarios.md) and [benchmark.md](benchmark.md); model, runtime or `kind` changes update [tech-stack.md](tech-stack.md) and [sources.md](sources.md); new terminology updates [glossary.md](glossary.md); scope or protocol changes update [research-design.md](research-design.md).

## Raw evidence

`benchmark/results/` and `benchmark/models/` are gitignored; copy accepted raw evidence under `research/<topic>/` so it is committed with its study. Accepted means the study reports its results; logs of failed or superseded tries are not kept.
