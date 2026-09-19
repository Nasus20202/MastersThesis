# Skill condition methodology

## Research question

Does the Skill condition improve Kubernetes task execution over Baseline and the selected Prompt condition when all three use the same model, sandbox, Bash capability, task text, execution limits and deterministic scoring?

The intended advantage of Skill is progressive disclosure: the agent can inspect the environment, then voluntarily load procedural or domain knowledge only when it is useful. The working hypothesis is **Baseline < Prompt < Skill**, but the benchmark decides the ordering; results are not adjusted to fit that hypothesis.

## Skill design

The Skill condition has three model tools in addition to normal model reasoning:

- `bash` executes commands in the benchmark sandbox;
- `load_skill` loads one skill body selected by the agent;
- `load_reference` loads one focused reference listed by a skill.

Loading is voluntary. The harness does not select a skill or reference from scenario IDs, keywords or grader knowledge. The router exposes only the available skill names and short descriptions. A loaded skill can then expose its references.

The intended routing path is:

1. inspect the sandbox;
2. identify the relevant procedure or Kubernetes area from observed evidence;
3. load a useful skill when additional knowledge is needed;
4. if that skill exposes a reference matching the observed subsystem, load the focused reference before making a specialized change;
5. repair and verify the final state.

This preserves progressive disclosure while making the ownership of domain knowledge explicit.

## Development evidence

### Original comparison

The original comparison used 5 scenarios × 3 conditions × 5 repetitions = **75 attempts**. It is exploratory development evidence, not a final generalization result.

Configuration: Gemma 4 E4B QAT Q4_0 (`google/gemma-4-E4B-it-qat-q4_0-gguf`, revision `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`, SHA-256 `676c35070db6dbe52f93e9c864ee0fba4eddea94b9c875d9cb10daff453fbaee`); llama.cpp `server-vulkan-b10964`, digest `sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53`; `--kv-unified-per-slot 32768`; server parallelism 2; reasoning on with a 4096-token budget; 25 turns, 50 tool calls, 60 seconds per tool call and 300 seconds per attempt; benchmark parallelism 4. Temperature and maximum output tokens were not overridden.

| Condition | Macro score | Full success | Tokens (total) | Tool calls                                      | Mean agent duration |
| --------- | ----------: | -----------: | -------------: | ----------------------------------------------- | ------------------: |
| Baseline  |   **0.640** |  15/25 (60%) |        304,594 | 152 Bash                                        |              147.5s |
| Prompt    |   **0.740** |  18/25 (72%) |        562,073 | 217 Bash                                        |              201.3s |
| Skill     |   **0.000** |    0/25 (0%) |         75,024 | 46 `load_skill`, 1 `load_reference`, **0 Bash** |               27.2s |

| Scenario                    | Baseline scores   | Prompt scores   | Skill scores  |
| --------------------------- | ----------------- | --------------- | ------------- |
| `container-crash-loop`      | 0, 0, 0, 0, 0     | 0, 1, 0, 0, 0   | 0, 0, 0, 0, 0 |
| `image-pull-failure`        | 1, 1, 1, 1, 1     | 1, 1, 1, 1, 1   | 0, 0, 0, 0, 0 |
| `missing-rbac-binding`      | 0, 0, 0.5, 0.5, 0 | 0, 1, 1, 0.5, 0 | 0, 0, 0, 0, 0 |
| `service-selector-mismatch` | 1, 1, 1, 1, 1     | 1, 1, 1, 1, 1   | 0, 0, 0, 0, 0 |
| `unschedulable-cpu-request` | 1, 1, 1, 1, 1     | 1, 1, 1, 1, 1   | 0, 0, 0, 0, 0 |

The Skill result does **not** show that the skill knowledge was worse than prompting. The Skill agent made no Bash calls in any of its 25 attempts, so it never interacted with the environment it was graded on. It mostly loaded material, asked for information that was available in the sandbox, or stopped after diagnosis. The low token count and short duration are consequences of this early termination.

The original Skill condition therefore exposed a harness/prompt contract failure before it could meaningfully test skill content.

### Corrected Skill diagnostic

A separate diagnostic added an explicit instruction to inspect and repair the sandbox with Bash and reran Skill for 5 repetitions on the same 5 scenarios. This was **not** a replacement controlled three-condition comparison.

| Scenario                    | Diagnostic scores |    Full success |
| --------------------------- | ----------------- | --------------: |
| `container-crash-loop`      | 0, 0, 0, 0, 1     |             1/5 |
| `image-pull-failure`        | 1, 0, 1, 0, 1     |             3/5 |
| `missing-rbac-binding`      | 0, 0, 0, 1, 1     |             2/5 |
| `service-selector-mismatch` | 1, 1, 1, 0, 1     |             4/5 |
| `unschedulable-cpu-request` | 1, 0, 1, 1, 1     |             4/5 |
| **All scenarios**           |                   | **14/25 (56%)** |

The diagnostic macro score was **0.560**. It used 605,975 tokens and 203 tool calls: 166 Bash, 34 `load_skill`, and only **3 `load_reference`** calls. Four attempts still made no Bash calls. Four attempts reached the 300-second limit.

This run shows that restoring environment interaction removed the catastrophic 0/25 failure, but routing remained inconsistent. The Skill condition still trailed the original Baseline (0.640) and Prompt (0.740) results.

## Why Skill underperformed

The raw traces point to three separate problems.

### 1. The first run did not execute the task

The original router described skills as optional knowledge but did not make the hands-on execution contract strong enough for Gemma 4 E4B. All 25 Skill attempts made zero Bash calls. That run primarily measured whether the model would infer the missing operational contract, not whether skills helped Kubernetes work.

The branch now uses the shared execution contract for hands-on agents. That change was not present in either raw run and therefore still needs measurement.

### 2. Skill selection worked better than reference selection

In the corrected diagnostic the model frequently loaded `troubleshooting`, `kubectl`, or `kubernetes`, but it almost never progressed from a broad skill to a focused reference. Across 25 attempts it made 34 skill loads but only 3 reference requests.

The old routing language only said that references could be loaded “when more detail is needed.” It did not tell the model to treat a diagnosed subsystem as a routing decision. The Kubernetes skill listed `authorization.md`, `scheduling.md`, `networking.md`, and other references, but Gemma often continued from its own knowledge instead of consulting the matching reference.

### 3. Reference ownership was unclear

The RBAC attempts expose the routing problem directly:

- **Attempt 1 — score 0.** It loaded `troubleshooting`, `kubectl` and `kubernetes`, saw `authorization.md` in the Kubernetes reference list, but never loaded it. It inspected the already-correct Role, deleted the failing Pod, then incorrectly patched the Role API group to `.`.
- **Attempt 4 — score 1.** It loaded `kubernetes`, `troubleshooting`, `configuration.md` and `authorization.md`. It later observed HTTP 403, inspected the RoleBinding, found subject `wrong-reader`, and patched it to `config-reader`.
- **Attempt 5 — score 1.** It loaded only `kubectl`; after finding the bad RoleBinding it requested a nonexistent `RoleBinding structure` reference from the `kubectl` skill. It recovered by reasoning from the object and patched the correct subject anyway.

Attempt 4 is only one observation, so it does not establish that loading `authorization.md` caused success. It does show that the reference contained directly relevant RBAC semantics and that the agent could use the progressive-disclosure path successfully. Attempts 1 and 5 show why the routing contract needed to make topic-to-skill and skill-to-reference ownership clearer.

## Routing revision after the diagnostic

The Skill package was revised from the raw evidence without adding scenario IDs, grader criteria, exact repairs, or automatic routing:

- the system prompt now says to inspect first and use skills as on-demand knowledge rather than a checklist;
- after observations identify a domain, the agent is told to load a matching skill voluntarily;
- after a skill exposes references, the agent is told to load a focused reference when it matches the observed failure or the subsystem about to be changed;
- the Kubernetes skill description now advertises its broader domain scope, including RBAC, scheduling, networking, storage, configuration and lifecycle;
- the Kubernetes skill maps generic evidence such as `403`/`Forbidden`, ServiceAccounts, Roles and RoleBindings to `authorization.md`, with equivalent examples for scheduling, resources, images, networking, configuration and workload failures;
- the `kubectl` skill now routes authorization findings to the Kubernetes skill and its `authorization.md` reference instead of implying that command knowledge owns RBAC semantics;
- the troubleshooting skill now reminds the agent to consult a matching Kubernetes reference before a specialized repair when exact semantics or constraints matter.

Skill and reference loading remain model-selected. No keyword matcher or harness-side router chooses content for the model.

These changes are a new development configuration. **No benchmark result is attributed to them yet.**

## Next test

First run a focused RBAC smoke test to verify the routing behavior itself. The useful signal is not merely whether the task passes: inspect whether the model recognizes authorization evidence, loads `kubernetes`, then loads `authorization.md` before making an RBAC change.

If routing behaves as intended, rerun the complete controlled comparison with Baseline, Prompt and Skill under the same current runner and runtime settings. Preserve every run, including failures. Compare macro score and full-success rate first; tool use, reference selection, tokens, duration and termination state are supporting evidence.

The desired research hypothesis is Baseline < Prompt < Skill because Skill adds optional domain knowledge on top of the common execution capability. The experiment must still report the observed ordering if that hypothesis is not supported.

The expanded benchmark is development/tuning data for later Prompt/Skill optimization. Once the best observed configurations are selected, freeze them before the final evaluation set; do not tune on final evaluation results.

## Raw data

The committed raw attempt records are the source of truth for the numbers and failure analysis above:

- [original three-condition comparison](raw/run-2026-09-19-12-13-42-245Z/)
- [corrected Skill diagnostic](raw/run-2026-09-19-17-51-41-483Z/)
- [machine-readable summary](results.json)

Raw records include complete model messages, tool calls and outputs, token usage, termination state and deterministic grading evidence. They are preserved unchanged; later routing revisions must create new runs rather than overwrite these results.

## Literature context

Published and preprint evidence does not imply that skills must outperform prompting on this benchmark. Reported gains depend on skill quality, routing and harness behavior.

| Study           | Relevant result                                                                                           | Limitation                                                    |
| --------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| SkillsBench [2] | Curated skills improved task-macro pass rate from **33.9% to 50.5%** across its evaluated configurations. | Different tasks, models and harnesses; some tasks regressed.  |
| Liu et al. [3]  | For Claude, no skill **35.4%**, forced skill **55.4%**, agent selection **51.2%**, distractors **43.5%**. | Selection quality and skill quality are coupled.              |
| SkillJuror [4]  | Progressive disclosure **46.1%** vs flat skill **42.0%**.                                                 | One model/harness; task-level confidence interval spans zero. |
| Huang [5]       | Task-only **68.3%** vs generated flat skills **67.5%**; no significant condition difference.              | Data-science tasks and generated skills.                      |
| Skill-Use [6]   | Separates trigger, procedure compliance and boundary behavior.                                            | Does not estimate task-success gain from skills.              |
| ReAct [1]       | Interleaved reasoning/action improved ALFWorld success over action-only prompting.                        | Prompt intervention, not skills.                              |

## Sources

1. Shunyu Yao et al. 2023. “ReAct: Synergizing Reasoning and Acting in Language Models.” _ICLR 2023_, arXiv:2210.03629v2. [OpenReview](https://openreview.net/forum?id=WE_vluYUL-X); [arXiv](https://arxiv.org/abs/2210.03629v2).
2. Xiangyi Li et al. 2026. “SkillsBench: Benchmarking How Well Agent Skills Work Across Diverse Tasks.” arXiv:2602.12670v4, 14 June 2026. [arXiv](https://arxiv.org/abs/2602.12670v4).
3. Yujian Liu et al. 2026. “How Well Do Agentic Skills Work in the Wild: Benchmarking LLM Skill Usage in Realistic Settings.” arXiv:2604.04323v1. [arXiv](https://arxiv.org/abs/2604.04323v1).
4. Zhiyu Chen et al. 2026. “SkillJuror: Measuring How Agent Skill Organization Changes Runtime Behavior.” arXiv:2606.11543v1. [arXiv](https://arxiv.org/abs/2606.11543v1).
5. Wei-Jung Huang. 2026. “Do LLM-Generated Skills Make Better AI Data Scientists? A Component Ablation Across Data-Science Workflows.” arXiv:2607.07504v1. [arXiv](https://arxiv.org/abs/2607.07504v1).
6. Jinyi Han et al. 2026. “Skill-Use: Can LLMs Actually Use Skills in Agentic Harnesses?” arXiv:2608.04828v1. [arXiv](https://arxiv.org/abs/2608.04828v1).
7. Barry Zhang, Keith Lazuka and Mahesh Murag. 2025. “Equipping Agents for the Real World with Agent Skills.” Anthropic Engineering, 16 October 2025. [Article](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills).

Literature checked 2026-09-19. The arXiv studies listed here are preprints.
