# Skills development evaluation

## Status

Active development study for tickets [#15](https://github.com/Nasus20202/MastersThesis/issues/15), [#16](https://github.com/Nasus20202/MastersThesis/issues/16), [#17](https://github.com/Nasus20202/MastersThesis/issues/17), and [#18](https://github.com/Nasus20202/MastersThesis/issues/18). All four tickets are **In progress**. The next step is the approved repeated development evaluation; this document will record its raw evidence, analysis, decisions and any later approved iteration.

## Question

On the five current development incidents, does the skill condition improve deterministic Kubernetes repair over baseline prompting and the selected troubleshooting prompt?

## Approved development protocol

The evaluation runs the three conditions on every current development scenario for five independent repetitions: **3 conditions × 5 scenarios × 5 repetitions = 75 attempts**.

| Element            | Approved value                                                                                                                                 |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Conditions         | `baseline`, `prompt`, `skill`                                                                                                                  |
| Scenarios          | `container-crash-loop`, `image-pull-failure`, `missing-rbac-binding`, `service-selector-mismatch`, `unschedulable-cpu-request`                 |
| Repetitions        | Five fresh attempts per condition and scenario                                                                                                 |
| Model              | Gemma 4 E4B QAT Q4_0 (`google/gemma-4-E4B-it-qat-q4_0-gguf` at revision `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`)                            |
| Runtime            | llama.cpp Vulkan `server-vulkan-b10964`, digest `sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53`; two inference slots |
| Per-attempt limits | 25 turns, 50 tool calls, 60 seconds per tool call, 300 seconds total                                                                           |
| Scoring            | Deterministic weighted normalized score and separate full-success result                                                                       |

Baseline exposes the common Bash sandbox. Prompt adds the selected troubleshooting prompt. Skill adds the routing prompt, `load_skill`, `load_reference`, the reusable troubleshooting skill, and Kubernetes reference material. Prompt and skill include the shared execution contract: inspect the sandbox with Bash, repair the owning resource, and verify the outcome before stopping.

The benchmark runner limits active agent loops to the two configured inference slots. Cluster setup, grading and cleanup remain outside that limit. This run is the first measurement with that limit.

## Measures and analysis

The primary measure is macro-average normalized score: average the five repetitions within each scenario, then average scenarios equally. Secondary measures are full-success rate, per-scenario outcome distribution, tokens, tool use, agent duration, termination reason, inference errors and observable failure modes. Every attempt record will be copied to `raw/<run-id>/` with its `run.json`; derived summaries will cite that directory.

The five scenarios were available while the prompt and skill were developed. Results are development evidence, not a generalization estimate. They must not be used to alter a recorded result or claim final benchmark performance.

## Iteration rule

After each complete 75-attempt run, analyze the preserved evidence and present any proposed prompt, skill, execution or protocol change to the researcher. No change will be made or measured until it is explicitly approved. An approved change starts a new complete 75-attempt run; earlier evidence remains labeled with the configuration that produced it.

## Sources

The skill package's source traceability is maintained in [`benchmark/internal/agent/skill/SOURCES.md`](../../../benchmark/internal/agent/skill/SOURCES.md). Common benchmark and scoring decisions are recorded in the [decision log](../../decision-log.md).
