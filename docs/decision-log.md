# Decision log

## Purpose

This file records research and project decisions, including when they were made, their current status and later changes.

Detailed descriptions belong in the linked project documents. Historical decisions should not be silently rewritten. When a decision changes, add a new entry and mark the previous entry as superseded.

## Status values

- **Approved** — currently accepted for the project.
- **Provisional** — accepted as a starting point but not frozen for final evaluation.
- **Open** — not decided yet.
- **Superseded** — replaced by a later decision.

## Decisions

| ID | Date | Decision | Rationale | Status | Reference |
|---|---|---|---|---|---|
| D-001 | 2026-09-07 | The initial research domain is Kubernetes troubleshooting in reproducible local environments. | Provides practical, observable repair tasks rather than isolated questions. | Approved, initial scope | [research-design.md](research-design.md) |
| D-002 | 2026-09-07 | The initial comparison includes baseline, prompt, skill, RAG, fine-tuning and harness conditions. | Covers the planned runtime, knowledge and weight-adaptation approaches. | Approved, initial comparison | [research-design.md](research-design.md) |
| D-003 | 2026-09-07 | The benchmark runner and experiment orchestration will be implemented in Go as a CLI tool. | Keeps the benchmark implementation focused and practical to operate. | Approved, implementation direction | [tech-stack.md](tech-stack.md) |
| D-004 | 2026-09-07 | Docker will provide the execution and sandboxing boundary. Bash will be exposed as the raw execution interface, with Kubernetes interaction through tools such as `kubectl`. | Provides a concrete execution environment for the baseline while limiting host access. | Approved, implementation direction | [tech-stack.md](tech-stack.md) |
| D-005 | 2026-09-07 | Kubernetes incidents will run in local `kind` clusters. | Provides a reproducible local Kubernetes environment for fault injection and repair. | Approved, initial environment | [tech-stack.md](tech-stack.md) |
| D-006 | 2026-09-07 | Local model execution will initially use llama.cpp in Docker with the Vulkan backend. | Supports local execution while keeping the model-serving path common across conditions. | Approved, provisional runtime direction | [tech-stack.md](tech-stack.md) |
| D-007 | 2026-09-07 | Gemma 4 E4B is the initial model candidate, but it is not a permanent project requirement. | Establishes a starting model without making the research dependent on one artifact. | Provisional | [tech-stack.md](tech-stack.md) |
| D-008 | 2026-09-07 | If the initial model cannot reliably use the required execution or tool-calling interface, Gemma 4 12B or Qwen 3.5 9B may be evaluated after checking the model template, runtime and integration. | Separates model capability failures from runtime or integration failures. | Provisional fallback plan | [tech-stack.md](tech-stack.md) |
| D-009 | 2026-09-07 | Scenario verification should be deterministic, support partial credit and track complete task success separately. | Makes repairs auditable and captures progress from incomplete repairs. | Approved, evaluation principle | [benchmark.md](benchmark.md) |
| D-010 | 2026-09-07 | Benchmark conditions must be executed repeatedly because model behaviour is stochastic, while preserving raw results. | Allows effectiveness and stability to be compared instead of relying on one run. | Approved, evaluation principle | [research-design.md](research-design.md) |
| D-011 | 2026-09-07 | The RAG condition will use the complete selected collection of official Kubernetes documentation as its intended corpus basis. | Tests retrieval over broad product documentation rather than scenario-specific excerpts. | Provisional; snapshot to be recorded | [benchmark.md](benchmark.md) |
| D-012 | 2026-09-07 | Fine-tuning data must remain separate from benchmark incidents, and final benchmark incidents must not be included in training data. | Prevents direct benchmark leakage into the adapted model. | Approved, data-separation principle | [benchmark.md](benchmark.md) |
| D-013 | 2026-09-07 | The research and technical documentation will extend the existing files and add focused files for the benchmark, technology stack, decisions, glossary and sources. | Keeps each document focused while preserving a clear project entrypoint. | Approved, documentation structure | [research-design.md](research-design.md) |

## Open decisions

Open decisions are maintained only in this section. Other documents should describe their subject area and link here rather than repeat this list.

- exact model artifact, revision, hash and quantization,
- exact llama.cpp image and revision,
- Kubernetes version and cluster configuration,
- benchmark scenario set and size,
- repeated-run count,
- execution, resource and network limits,
- final scoring criteria and aggregation,
- documentation corpus snapshot,
- embedding and retrieval configuration,
- fine-tuning dataset and training configuration,
- final harness tools,
- whether `README.md` needs a small project-description update.

Changes to `AGENTS.md` require explicit approval.

## Change procedure

When a decision changes:

1. keep the previous entry,
2. mark it as **Superseded**,
3. add a new entry with the date and reason,
4. update the affected document,
5. preserve any source or result that explains the change.
