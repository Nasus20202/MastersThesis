# Skill condition methodology

## Research question

Does the skill condition improve deterministic Kubernetes repair over baseline prompting and the selected troubleshooting prompt? All conditions use the same model, sandbox, Bash tool, task text, execution limits and scoring. The skill condition combines its router prompt, loaders, procedural skills and Kubernetes references.

## Design

The skill prompt stays short and lists available skills. The agent can load a skill or one of its references when needed. The troubleshooting skill supplies a general inspect, diagnose, repair and verify procedure; the Kubernetes material explains domain behavior. All conditions have the same Bash capability.

Loading material on demand can reduce unused initial context. It adds a routing decision, and a missed, incorrect or irrelevant load can waste context or stop useful work. Loaded content remains in the conversation, so token use may increase. Routing and token counts are secondary measures; task success is primary.

The prompt condition retains its detailed troubleshooting guidance. The skill condition keeps its routing and progressive-disclosure design; its prompt does not contain the troubleshooting procedure or Kubernetes reference content.

## Evidence

| Study           | Task and model                                                                                                                                                                                   | Comparison, metric and result                                                                                                                                                         | Limitation                                                                                                      |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| SkillsBench [2] | 87 terminal and container tasks across 8 domains; 18 model and harness configurations, including OpenHands with Claude Opus 4.7, GPT-5.4 Mini and Gemini 3.1 Flash Lite; 3 trials per condition. | No skill vs curated skills; task-macro pass rate **33.9% to 50.5% (+16.6 pp)**.                                                                                                       | No strong prompt-only comparison; task-skill curation and context size differ; 13/87 tasks regressed. Preprint. |
| Liu et al. [3]  | 84 SkillsBench tasks, 3 repetitions; Claude Opus 4.6/Claude Code, Kimi K2.5/Terminus-2 and Qwen3.5-397B-A17B/Qwen-Code.                                                                          | For Claude, no skill **35.4%**, forced skill **55.4%**, agent selection **51.2%**, and skill distractors **43.5%** task pass rate.                                                    | Retrieval, selection and skill quality are coupled; no prompt-only comparator. Preprint.                        |
| SkillJuror [4]  | 82 SkillsBench tasks; GPT-5.4 high with Codex/Harbor; 5 trials per condition.                                                                                                                    | No skill **29.0%**, flat skill **42.0%**, progressive disclosure **46.1%**. Progressive vs flat: **+4.1 pp** (17/410); task-level 95% CI spans zero. Tokens per pass: 0.22M vs 0.21M. | One model and harness; format effect is small and uncertain. Preprint.                                          |
| Huang [5]       | 56 data-science tasks; 9 OpenAI, Google and Anthropic model configurations; 3 repetitions.                                                                                                       | Task-only prompt vs generated flat skills; majority-vote pass rate **68.3% vs 67.5%**; no condition significant (_p_ ≥ .396).                                                         | Generated data-science skills and different tasks and execution setup. Preprint.                                |
| Skill-Use [6]   | 177 tasks, 79 skills, 9 domains; 8 LLMs and 2 harnesses.                                                                                                                                         | Studies trigger, procedure compliance and boundary behavior; highest combined score **0.613** (GPT-5.5/Claude Code).                                                                  | Does not compare task success with and without skills. Preprint.                                                |
| ReAct [1]       | ALFWorld tasks; PaLM-540B.                                                                                                                                                                       | Action-only prompt vs interleaved reasoning and action; reported task success **45% vs 71%**.                                                                                         | Prompt intervention, not skills; different model and tasks; reports selected best runs. ICLR 2023.              |

These studies do not estimate the gain expected on this Gemma 4 E4B Kubernetes benchmark. Skill gains vary by task and routing method; prompt changes can also improve tool use. Measure baseline, prompt and skill directly.

## Metrics

Primary effectiveness metric: macro-average normalized score, averaging repetitions within each scenario and then scenarios equally. Key secondary metric: full-success rate. Report all repetition outcomes, failure modes, tokens, tool calls and agent duration. For the skill condition, also report skill and reference calls, loaded names and evident incorrect or unnecessary requests.

## Leakage and generalization

The five development scenarios existed while the skills were being created. The audit found no scenario IDs, scenario text or exact scenario repair recipes in the model-visible skill material. The references cover the same broad topics, and the troubleshooting procedure was informed by development weaknesses. Results on these five scenarios are development evidence, not a generalization test.

Before authoring the expanded set, audit the current model-visible skill prompt and files, record their digest and freeze them. Create new scenarios after the freeze. Do not tune the frozen package from those results.

## Development comparison

The original comparison used 5 scenarios × 3 conditions × 5 repetitions = **75 attempts**. It is exploratory; five repetitions give only coarse per-scenario rates. After this run, prefer broader scenario coverage to more repetitions of these five scenarios.

Configuration: Gemma 4 E4B QAT Q4_0 (`google/gemma-4-E4B-it-qat-q4_0-gguf`, revision `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`, SHA-256 `676c35070db6dbe52f93e9c864ee0fba4eddea94b9c875d9cb10daff453fbaee`); llama.cpp `server-vulkan-b10964`, digest `sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53`; `--kv-unified-per-slot 32768`; server parallelism 2; reasoning on with a 4096-token budget; 25 turns, 50 tool calls, 60 seconds per tool call and 300 seconds per attempt; benchmark parallelism 4. Temperature and maximum output tokens were not set, so runtime defaults applied. Scenarios: `container-crash-loop`, `image-pull-failure`, `missing-rbac-binding`, `service-selector-mismatch` and `unschedulable-cpu-request`.

| Condition    | Macro score | Full success | Tokens (prompt / completion / total) |                                  Tool calls | Agent duration (sum / mean) |
| ------------ | ----------: | -----------: | -----------------------------------: | ------------------------------------------: | --------------------------: |
| Baseline     |   **0.640** |  15/25 (60%) |           232,289 / 72,305 / 304,594 |                                    152 Bash |             3,686s / 147.5s |
| Final prompt |   **0.740** |  18/25 (72%) |           465,210 / 96,863 / 562,073 |                                    217 Bash |             5,031s / 201.3s |
| Skill        |   **0.000** |    0/25 (0%) |             51,860 / 23,164 / 75,024 | 46 `load_skill`, 1 `load_reference`; 0 Bash |                679s / 27.2s |

| Scenario                    | Baseline scores, repetitions 1–5 | Prompt scores, repetitions 1–5 | Skill scores, repetitions 1–5 |
| --------------------------- | -------------------------------- | ------------------------------ | ----------------------------- |
| `container-crash-loop`      | 0, 0, 0, 0, 0                    | 0, 1, 0, 0, 0                  | 0, 0, 0, 0, 0                 |
| `image-pull-failure`        | 1, 1, 1, 1, 1                    | 1, 1, 1, 1, 1                  | 0, 0, 0, 0, 0                 |
| `missing-rbac-binding`      | 0, 0, 0.5, 0.5, 0                | 0, 1, 1, 0.5, 0                | 0, 0, 0, 0, 0                 |
| `service-selector-mismatch` | 1, 1, 1, 1, 1                    | 1, 1, 1, 1, 1                  | 0, 0, 0, 0, 0                 |
| `unschedulable-cpu-request` | 1, 1, 1, 1, 1                    | 1, 1, 1, 1, 1                  | 0, 0, 0, 0, 0                 |

The final prompt exceeded baseline by 0.100 macro-score points and 12 percentage points in full success. The original skill run made no Bash calls in all 25 attempts; it requested cluster details or gave a diagnosis without inspecting the sandbox. Its lower token use and duration came from stopping early. Seven attempts recorded llama.cpp `context deadline exceeded` (two baseline and five prompt); two prompt attempts still passed grading.

## Corrected skill diagnostic and follow-up

After the first skill result, the skill prompt received a one-sentence instruction to inspect and repair the sandbox with Bash. The separate rerun used the same five scenarios and five repetitions; it is not a replacement three-condition comparison. [`results.json`](results.json) records both run summaries and the diagnostic's attempt outcomes. The benchmark's full attempt records remain under `benchmark/results/` on the run host.

| Scenario                    | Diagnostic scores, repetitions 1–5 | Full success |
| --------------------------- | ---------------------------------- | -----------: |
| `container-crash-loop`      | 0, 0, 0, 0, 1                      |          1/5 |
| `image-pull-failure`        | 1, 0, 1, 0, 1                      |          3/5 |
| `missing-rbac-binding`      | 0, 0, 0, 1, 1                      |          2/5 |
| `service-selector-mismatch` | 1, 1, 1, 0, 1                      |          4/5 |
| `unschedulable-cpu-request` | 1, 0, 1, 1, 1                      |          4/5 |

The diagnostic scored **0.560** macro-normalized and **14/25 (56%)** full success. It used 605,975 tokens and 203 tool calls (166 Bash, 34 `load_skill`, 3 `load_reference`); mean agent duration was 168.3 seconds. Four attempts made no Bash calls: two loaded `troubleshooting` and asked for sandbox details; two made no tool calls. Four attempts reached the 300-second limit (three RBAC and one scheduling); one still passed grading. The RBAC attempt 5 requested `RoleBinding structure` from `kubectl`, which has no reference files.

The remaining zero-Bash attempts show the one-sentence correction did not prevent early stops. The shared execution contract now added to both prompt and skill conditions was not present during either comparison. Its effect has not been measured.

The four benchmark workers could also run model-agent loops against two llama.cpp slots. The runs recorded seven inference request deadline errors and four agent timeouts. Queue wait was not measured, so these results do not establish queueing as the cause. The runner now limits the complete agent phase to `LLAMA_PARALLEL`; cluster setup, grading and cleanup remain outside the limit. No inference has been run with this change.

## Sources

1. Shunyu Yao et al. 2023. “ReAct: Synergizing Reasoning and Acting in Language Models.” _ICLR 2023_, arXiv:2210.03629v2. [OpenReview](https://openreview.net/forum?id=WE_vluYUL-X); [arXiv](https://arxiv.org/abs/2210.03629v2).
2. Xiangyi Li et al. 2026. “SkillsBench: Benchmarking How Well Agent Skills Work Across Diverse Tasks.” arXiv:2602.12670v4, 14 June 2026. [arXiv](https://arxiv.org/abs/2602.12670v4).
3. Yujian Liu et al. 2026. “How Well Do Agentic Skills Work in the Wild: Benchmarking LLM Skill Usage in Realistic Settings.” arXiv:2604.04323v1. [arXiv](https://arxiv.org/abs/2604.04323v1).
4. Zhiyu Chen et al. 2026. “SkillJuror: Measuring How Agent Skill Organization Changes Runtime Behavior.” arXiv:2606.11543v1. [arXiv](https://arxiv.org/abs/2606.11543v1).
5. Wei-Jung Huang. 2026. “Do LLM-Generated Skills Make Better AI Data Scientists? A Component Ablation Across Data-Science Workflows.” arXiv:2607.07504v1. [arXiv](https://arxiv.org/abs/2607.07504v1).
6. Jinyi Han et al. 2026. “Skill-Use: Can LLMs Actually Use Skills in Agentic Harnesses?” arXiv:2608.04828v1. [arXiv](https://arxiv.org/abs/2608.04828v1).
7. Barry Zhang, Keith Lazuka and Mahesh Murag. 2025. “Equipping Agents for the Real World with Agent Skills.” Anthropic Engineering, 16 October 2025. [Article](https://www.anthropic.com/engineering/equipping-agents-for-the-real-world-with-agent-skills).

Literature checked 2026-09-19. The arXiv studies listed here are preprints.
