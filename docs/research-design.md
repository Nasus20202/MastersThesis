# Research design

## Research goal

Compare how different ways of adapting a language model affect its ability to diagnose and repair technical incidents that require product-specific knowledge in a controlled local environment.

The research focuses on practical troubleshooting rather than question answering. The model should use observations from a faulty system, reason about the failure and take actions that restore the expected behaviour.

## Research questions

The current research questions are:

1. To what extent do different adaptation methods improve incident repair compared with the baseline condition?
2. Which adaptation methods produce more complete repairs, rather than only partial recovery?
3. How consistently does each adaptation method perform across repeated runs?
4. What time, tool-use and computational-cost trade-offs accompany the effectiveness and stability of each method?

These questions describe the current research framing. They may be refined before the final evaluation protocol is frozen.

## Initial research scope

The initial execution direction is troubleshooting technical incidents in reproducible local `kind` environments. Official vanilla Kubernetes documentation is the initial candidate source corpus. Its exact Kubernetes minor version and upstream documentation revision remain open until candidate qualification.

The scope is intentionally a starting point. Scope changes are recorded in the [decision log](decision-log.md).

## Adaptation methods

The comparison is intended to include the following conditions:

- **baseline** — model knowledge with the minimal execution capabilities needed to interact with the environment,
- **prompt** — baseline with a carefully designed troubleshooting system prompt,
- **skill** — baseline with reusable procedural troubleshooting knowledge,
- **RAG** — baseline with retrieved information from the technical documentation corpus selected through source-corpus qualification,
- **fine-tuning** — a model adapted using separate technical troubleshooting examples,
- **harness** — baseline with richer structured operational tools and controlled external knowledge access.

These names describe the research conditions at a high level. Their concrete capabilities and execution rules are defined in [benchmark.md](benchmark.md) and [tech-stack.md](tech-stack.md).

## Research outcomes

The comparison should examine:

- how often incidents are repaired,
- how completely and consistently they are repaired,
- how much time and model/tool usage each condition requires,
- what trade-offs are introduced by each adaptation method.

The benchmark defines the observable scoring and result format. Fine-tuning cost should be considered separately from inference-time cost.

## Comparison principles

The conditions should use the same benchmark and evaluation process wherever possible. The intended experimental difference is the adaptation method, not an unrelated change in the task or scoring.

Any necessary difference in model, runtime, tools or limits must be recorded and considered when interpreting the results.

The study should preserve raw results, including negative and inconclusive outcomes. Conclusions should be based on repeated runs rather than on a single model execution.

## Development and final evaluation

Development scenarios may be used to improve prompts, skills, retrieval, training data and harness tools.

The final benchmark, conditions and evaluation protocol should be frozen before final evaluation. Final results must not be used to change the protocol retrospectively.

Research and technical decisions are recorded in the [decision log](decision-log.md), with supporting references collected in [sources.md](sources.md).
