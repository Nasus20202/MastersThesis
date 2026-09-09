# Benchmark

## Document scope

This document defines how benchmark incidents are represented, executed and verified.

The research motivation and comparison rationale are described in [research-design.md](research-design.md). Technology choices are described in [tech-stack.md](tech-stack.md).

## Initial benchmark scope

The initial benchmark direction is troubleshooting technical incidents in reproducible local `kind` clusters. The execution environment does not determine the final knowledge domain or source corpus.

A benchmark scenario should require the model to inspect and modify a running system. Simple knowledge questions without an observable system state are outside the intended benchmark scope.

Changes to the benchmark scope are recorded in the [decision log](decision-log.md).

## Source corpus status

The common experiment corpus is the official English Kubernetes 1.37 documentation at the immutable `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15` revision recorded in [sources.md](sources.md) and [the knowledge check](research/kubernetes-knowledge-check/README.md).

All experiment conditions that use the common corpus must use this exact revision. This requirement preserves version alignment and reproducibility without selecting the runtime Kubernetes cluster configuration in this document.

## Scenario definition

Each scenario should record the following fields in a small, human-readable form:

- a stable scenario identifier and title,
- a known-good Kubernetes environment,
- an application or workload with expected behaviour,
- a controlled injected fault,
- a task description visible to the model,
- observable repair criteria for the expected repaired state,
- a reset procedure,
- source references and ground-truth metadata used during construction and validation.

The task description is the model-facing part of the scenario. The injected fault, expected diagnosis, repair criteria, verifier implementation details, reset procedure and source/ground-truth metadata are evaluator data. They must not disclose the fault or answer to the model.

Repair criteria should be observable, binary and distinct. Criteria should not check the same outcome more than once where practical. Each criterion should have an explicit positive weight reflecting its importance to the repaired state; criteria therefore need not contribute equally to the score. A criterion may use bounded deterministic waiting or polling when Kubernetes convergence is asynchronous; its timeout, polling interval and success condition must be fixed and recorded as part of the verifier.

## Scenario lifecycle

Every scenario should follow the same lifecycle:

1. prepare a known-good environment,
2. verify that the clean state behaves as expected,
3. inject the fault,
4. verify that the fault produces an observable failure,
5. provide the task and condition-specific capabilities to the model,
6. allow the model to inspect and modify the environment,
7. verify the resulting system behaviour,
8. preserve the raw result and supporting evidence,
9. reset the environment before the next run.

A scenario is suitable for evaluation only when its clean state, injected fault, repair verification and reset procedure are repeatable.

## Condition capabilities

The benchmark conditions initially have these operational differences:

| Condition   | Additional capability                                                                                                             |
| ----------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Baseline    | Bash access and the command-line tools required to interact with the Kubernetes environment                                       |
| Prompt      | Baseline capabilities plus an approved troubleshooting system prompt                                                              |
| Skill       | Baseline capabilities plus approved reusable troubleshooting skills                                                               |
| RAG         | Baseline capabilities plus context retrieved from the technical documentation corpus selected through source-corpus qualification |
| Fine-tuning | Baseline capabilities using a model adapted with separate technical troubleshooting examples                                      |
| Harness     | Baseline capabilities plus approved structured operational tools and controlled external knowledge access                         |

The exact tools, prompts, skills, retrieved context, model artifacts and access limits must be recorded for each evaluated condition.

## Model visibility

The model may see:

- the troubleshooting task,
- the capabilities available in its condition,
- command output and other observations produced during execution.

The model must not see:

- the injected fault,
- the expected root cause,
- verifier implementation details,
- hidden ground-truth data, including scenario source references.

Scenario source references are evaluator and analysis metadata. They must never be supplied as a hint or used to guide retrieval. RAG or Harness may independently retrieve the same source through an approved mechanism using only model-visible information, but the stored scenario reference itself must not be provided as a query, context or other external-knowledge input.

## Verification and scoring

Verification should be deterministic and based on observable system behaviour. It should check whether the expected behaviour has been restored, rather than require one exact command sequence or configuration representation.

The clean-state and fault-injection checks validate that the scenario is usable: the clean state must pass before fault injection, and the injected fault must produce the intended observable failure. These checks do not award repair credit.

After the model attempt, each repair criterion is evaluated independently as pass or fail. If repair criterion `i` has positive weight `w_i` and pass value `p_i` (`1` for pass and `0` for fail), the partial score is:

`partial score = sum(w_i * p_i) / sum(w_i)`

The score is therefore normalized to `[0, 1]`. Full task success is recorded separately as a boolean and is true only when every repair criterion passes, regardless of the weights. A scenario must have at least one repair criterion, and its total criterion weight must be positive.

Primary scoring does not use an LLM judge. The verifier may use bounded polling for convergence, but it must use a fixed bound and deterministic pass/fail condition rather than an unbounded wait or subjective interpretation.

Each run should preserve the outcome and raw evidence for every repair criterion, together with the aggregate partial score and full-success result. Verifier internals and hidden scenario metadata remain evaluator-side data and must not be included in the model-visible task, capabilities or execution transcript.

The current status of scoring decisions is recorded in the [decision log](decision-log.md).

## Repeated runs

Each scenario and condition should be executed multiple times because model behaviour is stochastic.

The current repetition decision is recorded in the [decision log](decision-log.md).

Each run should preserve, where available:

- scenario and condition identifiers,
- model and runtime identifiers,
- criterion-level verification results,
- partial score,
- complete-success status,
- execution time,
- token and tool usage,
- model and tool transcript,
- relevant runtime metadata,
- failure and reset information.

Raw results must be preserved, including negative and inconclusive results.

## Scenario acceptance

A scenario should be accepted only when:

- the clean environment passes verification,
- the injected fault produces an observable failure,
- an approved repair restores the expected behaviour,
- scoring is deterministic,
- the lifecycle can be repeated from a reset state.

Operational benchmark decisions are maintained in the [decision log](decision-log.md).
