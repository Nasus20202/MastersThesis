# Benchmark

## Document scope

This document defines how benchmark incidents are represented, executed and verified.

The research motivation and comparison rationale are described in [research-design.md](research-design.md). Technology choices are described in [tech-stack.md](tech-stack.md).

## Initial benchmark scope

The initial benchmark direction is Kubernetes troubleshooting in reproducible local `kind` clusters.

A benchmark scenario should require the model to inspect and modify a running system. Simple knowledge questions without an observable system state are outside the intended benchmark scope.

Changes to the benchmark scope are recorded in the [decision log](decision-log.md).

## Scenario definition

Each scenario should define:

- a known-good Kubernetes environment,
- an application or workload with expected behaviour,
- a controlled injected fault,
- a task description visible to the model,
- observable criteria for the expected repaired state,
- a reset procedure,
- source references used during construction and validation.

The source references, injected fault, expected diagnosis and verifier details must not be exposed to the model.

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

| Condition | Additional capability |
|---|---|
| Baseline | Bash access and the command-line tools required to interact with the Kubernetes environment |
| Prompt | Baseline capabilities plus an approved troubleshooting system prompt |
| Skill | Baseline capabilities plus approved reusable troubleshooting skills |
| RAG | Baseline capabilities plus context retrieved from the approved Kubernetes documentation corpus |
| Fine-tuning | Baseline capabilities using a model adapted with separate Kubernetes troubleshooting examples |
| Harness | Baseline capabilities plus approved structured operational tools and controlled external knowledge access |

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
- hidden ground-truth data,
- scenario source references unless the condition explicitly provides them through an approved retrieval or external-knowledge mechanism.

## Verification and scoring

Verification should be deterministic and based on observable system behaviour.

Independent criteria may award partial credit for partial repairs. Complete task success must be recorded separately from the partial score.

Verification should evaluate whether the expected behaviour has been restored. It should not require one exact command sequence or configuration representation.

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
