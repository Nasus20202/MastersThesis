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

## Hypothesis

The expected order of the conditions is baseline < prompt < {skill, RAG} < harness. Skill and RAG are parallel knowledge methods built on the prompt condition, and no order between them is assumed. Fine-tuning is compared with the prompt condition without an assumed direction.

This is a hypothesis, not a tuning target: an order that does not hold is reported as a result.

## Initial research scope

The initial execution direction is troubleshooting technical incidents in reproducible local `kind` environments. The common experiment corpus is the official English Kubernetes 1.37 documentation at the immutable revision recorded in [sources.md](sources.md) and [the knowledge check](research/kubernetes-knowledge-check/README.md); the final benchmark protocol remains open.

The scope is intentionally a starting point. Scope changes are recorded in the [decision log](decision-log.md).

## Adaptation methods

The comparison is intended to include the following conditions:

- **baseline** — model knowledge with the minimal execution capabilities needed to interact with the environment,
- **prompt** — baseline with a carefully designed troubleshooting system prompt,
- **skill** — baseline with reusable procedural troubleshooting knowledge,
- **RAG** — baseline with retrieved information from the technical documentation corpus selected through source-corpus qualification,
- **fine-tuning** — the prompt condition on a model adapted with LoRA, trained on reviewed agent trajectories from the development scenarios (D-046, D-047),
- **harness** — the skill and RAG conditions combined, with structured operational tools and harness checks (D-049).

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

Conditions are developed and evaluated on four models, Gemma 4 E2B and E4B and Qwen3.5 4B and 9B (D-037), so that no condition is fitted to one model's weaknesses and condition differences can be checked across model families and sizes. Each model is reported separately; the models are listed in [tech-stack.md](tech-stack.md). Fine-tuning is evaluated on Gemma 4 E4B and Qwen3.5 4B only (D-046).

Any necessary difference in model, runtime, tools or limits must be recorded and considered when interpreting the results.

The study should preserve raw results, including negative and inconclusive outcomes. Conclusions should be based on repeated runs rather than on a single model execution.

## Development and final evaluation

Development scenarios may be used to improve prompts, skills, retrieval, training data and harness tools.

The final evaluation runs on held-out scenarios whose faults do not occur in the development scenarios (D-044). They are not used for tuning, training or model selection, and no agent runs on them before the final evaluation. Their results are the official result of the comparison. The development scenarios are run again in the final evaluation for comparison only: the conditions were tuned on them, and the fine-tuning adapters are trained on them. Each held-out scenario runs five times and each development scenario once (D-045).

Condition differences are reported as paired differences per scenario with a 95% bootstrap interval (D-051).

The final benchmark, conditions and evaluation protocol should be frozen before final evaluation. Final results must not be used to change the protocol retrospectively.

Research and technical decisions are recorded in the [decision log](decision-log.md), with supporting references collected in [sources.md](sources.md).
