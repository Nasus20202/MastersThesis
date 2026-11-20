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

The expected order of the conditions is baseline < prompt < {skill, RAG} < harness. Skill and RAG are parallel knowledge methods that add knowledge to an instructed agent, and no order between them is assumed. Fine-tuning is compared with the prompt condition without an assumed direction.

This is a hypothesis, not a tuning target: an order that does not hold is reported as a result.

## Research scope

The work covers one domain: diagnosing and repairing Kubernetes incidents in reproducible local `kind` clusters. The common experiment corpus is the official English Kubernetes 1.37 documentation at the immutable revision recorded in [sources.md](sources.md) and [the knowledge check](research/kubernetes-knowledge-check/README.md).

Scope changes are recorded in the [decision log](decision-log.md).

## Adaptation methods

The comparison includes six conditions (D-002):

- **baseline** — the model with a shell and file tools in the sandbox and no system prompt,
- **prompt** — baseline with a troubleshooting system prompt (D-023),
- **skill** — baseline with a system prompt of its own and procedural troubleshooting skills the agent loads on demand (D-026, D-032),
- **RAG** — the prompt condition with one added search paragraph and a `search_docs` tool over the frozen documentation corpus (D-036),
- **fine-tuning** — the prompt condition on a model adapted with LoRA, trained on reviewed agent trajectories from the development scenarios (D-053, D-054),
- **harness** — the skill and RAG conditions combined, with structured operational tools and harness checks (D-048).

The RAG prompt is the prompt condition's prompt plus one paragraph, so prompt → RAG differs only by retrieval. The skill condition's system prompt was written during skill development and is a separate text, so prompt → skill changes the instructions as well as adding skills.

The capabilities and execution rules of each condition are defined in [benchmark.md](benchmark.md) and [tech-stack.md](tech-stack.md).

## Research outcomes

The comparison examines:

- how often incidents are repaired,
- how completely and consistently they are repaired,
- how much time and model/tool usage each condition requires,
- what trade-offs each adaptation method introduces.

The benchmark defines the scoring and the result format. Fine-tuning cost is reported separately from inference-time cost (D-055).

## Comparison principles

All conditions use the same benchmark, limits, scoring and result format. The intended experimental difference is the adaptation method, not a change in the task or scoring.

Conditions are developed and evaluated on four models, Gemma 4 E2B and E4B and Qwen3.5 4B and 9B (D-037), so that no condition is fitted to one model's weaknesses and condition differences can be checked across model families and sizes. Each model is reported separately; the models are listed in [tech-stack.md](tech-stack.md). Fine-tuning is evaluated on all four as well (D-053).

The conditions did not receive equal tuning effort during development (skill 11 candidates, RAG 6, prompt 3). After the shared settings were fixed, one targeted round on the prompt found no effect and left the prompts unchanged ([stopping-rule study](research/stopping-rule/README.md)); the difference in effort is reported as a limitation.

Any necessary difference in model, runtime, tools or limits is recorded and considered when interpreting the results.

Raw results are preserved, including negative and inconclusive outcomes. Conclusions rest on repeated runs, not on a single model execution.

## Development and final evaluation

Development scenarios may be used to improve prompts, skills, retrieval, training data and harness tools.

The final evaluation runs on held-out scenarios whose faults do not occur in the development scenarios (D-044). They are not used for tuning, training or model selection, and no agent runs on them before the final evaluation. Their results are the official result of the comparison. The development scenarios are run again in the final evaluation for comparison only: the conditions were tuned on them, and the fine-tuning adapters are trained on them. Each held-out scenario runs five times and each development scenario once (D-045).

Condition differences are reported as paired differences per scenario with a 95% bootstrap interval (D-049).

The hardware of the final evaluation is open ([decision log](decision-log.md#open-decisions)).

The final benchmark, conditions and evaluation protocol are frozen before the final evaluation. Final results must not be used to change the protocol retrospectively.

Research and technical decisions are recorded in the [decision log](decision-log.md), with supporting references collected in [sources.md](sources.md).
