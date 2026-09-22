# Benchmark

## Document scope

This document defines how benchmark operational tasks are represented, executed and verified.

The research motivation and comparison rationale are described in [research-design.md](research-design.md). Technology choices are described in [tech-stack.md](tech-stack.md).

## Initial benchmark scope

The development benchmark uses reproducible Kubernetes operational tasks in local `kind` clusters. It may include both troubleshooting tasks and constructive implementation or configuration tasks. The execution environment does not determine the final knowledge domain or source corpus.

A benchmark scenario should require the model to inspect or modify an observable system state. Simple knowledge questions without an executable environment are outside the intended benchmark scope.

Changes to the benchmark scope are recorded in the [decision log](decision-log.md).

## Source corpus status

The common experiment corpus is the official English Kubernetes 1.37 documentation at the immutable `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15` revision recorded in [sources.md](sources.md) and [the knowledge check](research/kubernetes-knowledge-check/README.md).

All experiment conditions that use the common corpus must use this exact revision. This requirement preserves version alignment and reproducibility without selecting the runtime Kubernetes cluster configuration in this document.

## Scenario definition

Each scenario should record the following fields in a small, human-readable form:

- a stable scenario identifier and title,
- a known-good Kubernetes environment,
- an application, workload or starting environment with expected behaviour,
- a task description visible to the model,
- optional controlled fault injection and fault verification for troubleshooting tasks,
- observable grading criteria for the expected final state,
- source references and ground-truth metadata used during construction and validation.

The task description is the model-facing part of the scenario. Fault details, expected diagnosis, grading criteria, verifier implementation details and source/ground-truth metadata are evaluator data. They must not disclose a hidden fault or answer to the model.

Grading criteria should be observable, binary and distinct. Criteria should not check the same outcome more than once where practical. Each criterion should have an explicit positive weight reflecting its importance to the expected final state; criteria therefore need not contribute equally to the score. A criterion may use bounded deterministic waiting or polling when Kubernetes convergence is asynchronous; its timeout, polling interval and success condition must be fixed and recorded as part of the verifier.

## Scenario lifecycle

Every scenario follows the same core lifecycle:

1. prepare a known-good starting environment,
2. verify that the initial state behaves as expected,
3. optionally inject and verify a controlled fault,
4. provide the task and condition-specific capabilities to the model,
5. allow the model to inspect and modify the environment,
6. verify the resulting system behaviour,
7. preserve the raw result and supporting evidence,
8. delete the disposable cluster.

All setup, fault-handling and grading commands run evaluator-side in a pinned setup container, not in the model sandbox, and are not part of the model-visible transcript.

`inject_fault` and `verify_fault` are independently optional. Troubleshooting scenarios may use both to create and confirm a reproducible failure, while a scenario may also verify an already-prepared failure without injecting it or inject a state change without a dedicated pre-model verification step. Constructive tasks can omit both and start directly from the verified environment.

A scenario is suitable for evaluation only when its starting state, optional fault setup and grading are repeatable.

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

## Agent execution limits

The common model/tool loop uses per-attempt limits of 25 model turns, 50 tool calls, 60 seconds per tool call and 600 seconds total agent runtime. The per-tool deadline and total deadline are independent: a timed-out tool call returns an error to the model so it may continue, while the 600-second deadline stops the entire attempt.

Model-visible Bash output is capped at 8 KiB so a single command cannot fill the context; the complete output remains in raw evidence.

The loop retains the model's reasoning content between tool calls and re-sends it with the next request, so the model keeps its reasoning context and the inference prompt cache can reuse the shared prefix. Reasoning is recorded in the raw attempt evidence.

The selected limits are preserved in each raw run result so later conditions can be compared under the same execution budget.

## Model visibility

The model may see:

- the operational task,
- the capabilities available in its condition,
- its own reasoning retained from earlier tool-call turns in the same attempt,
- command output and other observations produced during execution.

The model must not see:

- the injected fault, when present,
- the expected root cause,
- verifier implementation details,
- hidden ground-truth data, including scenario source references.

Scenario source references are evaluator and analysis metadata. They must never be supplied as a hint or used to guide retrieval. RAG or Harness may independently retrieve the same source through an approved mechanism using only model-visible information, but the stored scenario reference itself must not be provided as a query, context or other external-knowledge input.

## Verification and scoring

Verification should be deterministic and based on observable system behaviour. It should check whether the expected final state has been reached, rather than require one exact command sequence or configuration representation.

Initial-state checks validate that the scenario is usable. When fault verification is configured, it must confirm the intended observable failure or pre-model state. These checks do not award task credit.

After the model attempt, each grading criterion is evaluated independently as pass or fail. If criterion `i` has positive weight `w_i` and pass value `p_i` (`1` for pass and `0` for fail), the partial score is:

`partial score = sum(w_i * p_i) / sum(w_i)`

The score is therefore normalized to `[0, 1]`. Full task success is recorded separately as a boolean and is true only when every repair criterion passes, regardless of the weights. A scenario must have at least one repair criterion, and its total criterion weight must be positive.

Primary scoring does not use an LLM judge. The verifier may use bounded polling for convergence, but it must use a fixed bound and deterministic pass/fail condition rather than an unbounded wait or subjective interpretation.

Each run should preserve the outcome and raw evidence for every grading criterion, together with the aggregate partial score and full-success result. Verifier internals and hidden scenario metadata remain evaluator-side data and must not be included in the model-visible task, capabilities or execution transcript.

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
- failure and cleanup information.

For a scenario benchmark invocation, one logical run contains every selected
agent, scenario and repetition. Agent attempt evidence is stored under
`results/<run-id>/<scenario-id>/<agent-name>/<attempt>.json`; validation
results retain their separate scenario/case layout.

Raw results must be preserved, including negative and inconclusive results.

## Scenario acceptance

A scenario should be accepted only when:

- the prepared environment passes initial-state verification,
- any configured fault produces its intended observable failure,
- an approved validation action reaches the expected final state,
- scoring is deterministic,
- the lifecycle is repeatable in a fresh disposable cluster.

Operational benchmark decisions are maintained in the [decision log](decision-log.md).
