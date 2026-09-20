# Skills development evaluation

## Status

Development evaluation for tickets [#15](https://github.com/Nasus20202/MastersThesis/issues/15), [#16](https://github.com/Nasus20202/MastersThesis/issues/16), [#17](https://github.com/Nasus20202/MastersThesis/issues/17), and [#18](https://github.com/Nasus20202/MastersThesis/issues/18); all remain **In progress**. The initial comparison (75 attempts) and the required-routing iteration (25 attempts) are complete. The second run changed routing behavior only; troubleshooting-skill content has not been revised since the first run. Further changes await researcher review.

## Research question

On five development Kubernetes incidents, how does the skill condition perform relative to baseline and prompt conditions, and what changes in task outcomes and tool traces when troubleshooting-skill loading is required?

## Design

The first run compared three agent conditions. The second run repeated the skill condition after changing its router instruction to require loading the troubleshooting skill before the first Bash call. Both runs used five attempts per scenario. Baseline and prompt conditions were not repeated in the second run.

| Setting             | Value                                                                                                                                          |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Conditions in run 1 | `baseline`, `prompt`, `skill`                                                                                                                  |
| Scenarios           | `container-crash-loop`, `image-pull-failure`, `missing-rbac-binding`, `service-selector-mismatch`, `unschedulable-cpu-request`                 |
| Replication         | Five attempts per condition and scenario; 75 attempts in run 1 and 25 in run 2                                                                 |
| Model               | Gemma 4 E4B QAT Q4_0 (`google/gemma-4-E4B-it-qat-q4_0-gguf`), revision `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`                              |
| Runtime             | llama.cpp Vulkan `server-vulkan-b10964`, digest `sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53`; two inference slots |
| Attempt limits      | 25 turns, 50 tool calls, 60 seconds per tool call, 300 seconds total                                                                           |
| Scoring             | Deterministic weighted normalized score; full-success status reported separately                                                               |

Baseline used the common Bash sandbox. Prompt added the selected troubleshooting prompt. Skill added the routing prompt, `load_skill` and `load_reference` tools, troubleshooting instructions, and Kubernetes references. Prompt and skill shared the instruction to inspect the sandbox, repair the owning resource, and verify the result. In run 2, the router required `troubleshooting`; loading other listed skills and references remained conditional on whether the agent judged their guidance useful.

The runner limited active agent loops to the two configured inference slots. Cluster setup, grading, and cleanup were outside this limit. Temperature and maximum output tokens were unset, so runtime defaults applied.

## Measures

The primary measure is macro-average normalized score: average attempts within each scenario, then average the five scenario means equally. Secondary measures are full-success rate, per-scenario outcomes, tokens, tool calls, agent duration, termination reason, inference errors, and observed tool-trace patterns. A full success requires every deterministic grading criterion to pass.

## Results

### Run 1: optional skill routing

Run `run-2026-09-19-22-00-57-779Z` ran from 2026-09-19 22:00:57 to 23:32:41 UTC. All 75 attempts were persisted. Command: `make benchmark REPEAT=5 AGENT=all BENCHMARK_PARALLEL=4`. Source revision: `63ad7ef3fb2c270486a7e3a1d1ccdab68f0adbe0`.

| Condition | Macro score | Full success | Mean tokens per attempt | Tool calls                | Mean agent duration | 300-second timeouts |
| --------- | ----------: | -----------: | ----------------------: | ------------------------- | ------------------: | ------------------: |
| Baseline  |       0.680 |  16/25 (64%) |                  14,862 | 166 Bash                  |             115.7 s |                0/25 |
| Prompt    |       0.680 |  17/25 (68%) |                  23,125 | 221 Bash                  |             138.7 s |                4/25 |
| Skill     |       0.740 |  17/25 (68%) |                  33,088 | 230 Bash; 12 `load_skill` |             165.2 s |                4/25 |

| Scenario                    | Baseline score (full success) | Prompt score (full success) | Skill score (full success) |
| --------------------------- | ----------------------------: | --------------------------: | -------------------------: |
| `container-crash-loop`      |                    0.40 (2/5) |                  0.20 (1/5) |                 0.20 (1/5) |
| `image-pull-failure`        |                    1.00 (5/5) |                  1.00 (5/5) |                 1.00 (5/5) |
| `missing-rbac-binding`      |                    0.20 (0/5) |                  0.40 (2/5) |                 0.50 (1/5) |
| `service-selector-mismatch` |                    1.00 (5/5) |                  1.00 (5/5) |                 1.00 (5/5) |
| `unschedulable-cpu-request` |                    0.80 (4/5) |                  0.80 (4/5) |                 1.00 (5/5) |

The skill condition's macro score was 0.06 higher than baseline and prompt. Its full-success rate equaled prompt and exceeded baseline by one attempt. The difference was not uniform across scenarios: the skill condition scored lower on crash-loop repair, tied at ceiling on image-pull and service-selector repair, and scored higher on RBAC and scheduling. Four prompt attempts and four skill attempts reached the 300-second limit; one timed-out skill attempt nevertheless passed grading. Scores and termination outcomes are distinct measures.

#### Skill selection

The skill prompt listed `bash`, `kubectl`, `kubernetes`, and `troubleshooting`. In the 25 skill attempts, 12 loaded a skill and 13 loaded none. All 12 calls loaded `kubectl`; no attempt loaded `troubleshooting`, `kubernetes`, `bash`, or a reference. The troubleshooting procedure and Kubernetes reference material therefore were not presented during this run.

Attempts that loaded `kubectl` averaged 0.625 score and 6/12 full successes; attempts that loaded no skill averaged 0.846 and 11/13 full successes. This is a post hoc, nonrandomized split because the agent selected whether to load a skill. It does not estimate the effect of loading `kubectl`.

### Run 2: required troubleshooting routing

Run `run-2026-09-20-08-34-18-420Z` ran from 2026-09-20 08:34:18 to 09:08:16 UTC. Command: `make benchmark REPEAT=5 AGENT=skill BENCHMARK_PARALLEL=4`. Source revision: `e826c41cfe4a9cd92e7b2f50bf69c240835a0b38`. Model, runtime, scenarios, skill content, scoring, and attempt limits were held fixed from the first run; the routing instruction was changed. All 25 attempt records were saved. The benchmark command returned nonzero after three RBAC attempts reached the 300-second agent limit.

| Measure                     | Run 1 skill condition | Run 2 required-routing condition |
| --------------------------- | --------------------: | -------------------------------: |
| Macro-average score         |                 0.740 |                            0.840 |
| Full success                |           17/25 (68%) |                      21/25 (84%) |
| Mean tokens per attempt     |                33,088 |                           30,371 |
| Mean agent duration         |               165.2 s |                          157.3 s |
| Mean tool calls per attempt |                  9.68 |                             9.64 |
| Agent timeouts              |                  4/25 |                             3/25 |

| Scenario                    | Score | Full success | Agent timeouts |
| --------------------------- | ----: | -----------: | -------------: |
| `container-crash-loop`      |  0.40 |          2/5 |            0/5 |
| `image-pull-failure`        |  1.00 |          5/5 |            0/5 |
| `missing-rbac-binding`      |  0.80 |          4/5 |            3/5 |
| `service-selector-mismatch` |  1.00 |          5/5 |            0/5 |
| `unschedulable-cpu-request` |  1.00 |          5/5 |            0/5 |

All 25 attempts loaded `troubleshooting` before their first Bash call. **No other skill was loaded:** the available `bash`, `kubectl`, and `kubernetes` skills were not selected in any attempt. No reference was loaded. The router instruction required `troubleshooting` but made other skills and references conditional on perceived relevance. The traces establish that `troubleshooting` was loaded; they do not establish whether the agent followed all of its instructions.

Run 2's macro score was 0.10 higher than run 1's skill score, and it had four more full successes. This is a descriptive comparison between two sets of attempts, not a controlled estimate of the routing change: attempts were not paired, no baseline or prompt condition was rerun, and there were only five attempts per scenario.

### Trace observations in run 2

- **Crash-loop:** Three of five attempts failed full-success grading. In these attempts, the grader observed `updated=1` and `ready=3`; the criterion for all replicas to update was not met. Traces show that the invalid container arguments were patched, but rollout verification did not confirm full recovery. Across the run, agents issued six `kubectl edit` calls, six watch calls, and two direct Pod deletions.
- **RBAC:** Four of five attempts passed grading; three reached the 300-second limit, including two that passed grading. One attempt failed the least-privilege criterion. Its repair granted `list` (and, in some traces, `watch`) on ConfigMaps; the verifier requires access to the named ConfigMap while denying ConfigMap listing and Secret reads. The full-credit repair changed the existing RoleBinding's ServiceAccount subject.
- **Tool-call limits:** Ten skill tool calls across four attempts exceeded the 60-second per-call limit. These included watch and rollout-check calls. Three attempts also reached the 300-second total limit.

These observations come from trace inspection; no formal coding or inter-rater procedure was applied. They identify behaviors present in this run, not their prevalence outside this benchmark.

## Interpretation and limitations

In run 1, a skill was loaded in 12/25 attempts, and every load was `kubectl`; `troubleshooting` was never loaded. Under the required router in run 2, `troubleshooting` was loaded in all 25 attempts, but no additional skills or references were selected. Failures remained in crash-loop completion and RBAC repair. The results do not establish whether loading additional skills would improve performance or whether the remaining errors are caused by skill content.

This is a development evaluation on five scenarios that were available during prompt and skill development, so performance may be optimistic for these cases. The sample is small, and the only second-run change was assessed in an unpaired skill-only comparison. Results are descriptive; no inferential statistical analysis was conducted. The model and runtime represent one configuration. Tool-trace observations were reviewed qualitatively.

## Reproducibility and evidence

Raw attempt records and run metadata are preserved with each run:

- [Run 1 raw results](raw/run-2026-09-19-22-00-57-779Z/), including all 75 attempts and `run.json`.
- [Run 1 configuration snapshot](raw/run-2026-09-19-22-00-57-779Z/configuration/revision-63ad7ef/), containing the exact agent prompts, skills, references, and source used.
- [Run 2 raw results](raw/run-2026-09-20-08-34-18-420Z/), including all 25 attempts and `run.json`.
- [Run 2 configuration snapshot](raw/run-2026-09-20-08-34-18-420Z/configuration/revision-e826c41/), containing the exact 40-file agent configuration.
- [Skill source traceability](../../../benchmark/internal/agent/skill/SOURCES.md).
- Project-wide skill-routing decision: [D-026](../../decision-log.md).

The skill package was unchanged between runs. The router change and its project-wide scope are recorded in D-026. Earlier results remain associated with their original configuration; no raw outcomes have been replaced.
