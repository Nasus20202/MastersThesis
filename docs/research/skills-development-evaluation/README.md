# Skills development evaluation

## Status

Development evaluation for tickets [#15](https://github.com/Nasus20202/MastersThesis/issues/15), [#16](https://github.com/Nasus20202/MastersThesis/issues/16), [#17](https://github.com/Nasus20202/MastersThesis/issues/17), and [#18](https://github.com/Nasus20202/MastersThesis/issues/18); all remain **In progress**. Five runs (175 attempts) are complete: the initial three-condition comparison and four skill-only runs. Run 5 tested mandatory skill loading, domain-reference routing, and the revised kubectl guidance. Its analysis is recorded below; no further prompt revision or benchmark run has been selected.

## Research question

On five development Kubernetes incidents, how does the skill condition perform relative to baseline and prompt conditions, and what changes in task outcomes and tool traces when the troubleshooting workflow directs loading domain skills and focused references?

## Design

The first run compared three agent conditions. Runs 2–4 repeated the skill condition after successive changes to skill routing and troubleshooting guidance. Each run used five attempts per scenario. Baseline and prompt conditions were not repeated after run 1.

| Setting             | Value                                                                                                                                          |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| Conditions in run 1 | `baseline`, `prompt`, `skill`                                                                                                                  |
| Scenarios           | `container-crash-loop`, `image-pull-failure`, `missing-rbac-binding`, `service-selector-mismatch`, `unschedulable-cpu-request`                 |
| Replication         | Five attempts per condition and scenario; 75 attempts in run 1 and 25 attempts in each of runs 2–5                                             |
| Model               | Gemma 4 E4B QAT Q4_0 (`google/gemma-4-E4B-it-qat-q4_0-gguf`), revision `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`                              |
| Runtime             | llama.cpp Vulkan `server-vulkan-b10964`, digest `sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53`; two inference slots |
| Attempt limits      | 25 turns, 50 tool calls, 60 seconds per tool call, 300 seconds total                                                                           |
| Scoring             | Deterministic weighted normalized score; full-success status reported separately                                                               |

Baseline used the common Bash sandbox. Prompt added the selected troubleshooting prompt. Skill added the routing prompt, load_skill and load_reference tools, troubleshooting instructions, and Kubernetes references. Prompt and skill shared the instruction to inspect the sandbox, repair the owning resource, and verify the result. Run 2 required troubleshooting before the first Bash call. Run 3 added Kubernetes-skill and relevant-reference guidance. Run 4 required Kubernetes-skill and reference routing, kubectl guidance before mutation, and checks for current desired state and incomplete outcomes. Run 5 required troubleshooting and Kubernetes skills before Bash, relevant references after diagnosis, kubectl before mutation, bounded waits, explicit replica convergence checks, and an incomplete outcome when verification remained unresolved.

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

### Run 3: broader domain guidance

Run `run-2026-09-20-11-11-33-707Z` ran from 2026-09-20 11:11:33 to 11:50:13 UTC. Command: `make benchmark REPEAT=5 AGENT=skill BENCHMARK_PARALLEL=4`. Source revision: `13b3526ead27030454ab88719fb3fee3949618cf`. Run metadata records a clean tree and all 25 expected attempts. The command returned nonzero because one RBAC attempt reached the 300-second agent limit.

| Measure                     | Run 2 required-routing condition | Run 3 broader-guidance condition |
| --------------------------- | -------------------------------: | -------------------------------: |
| Macro-average score         |                            0.840 |                            0.820 |
| Full success                |                      21/25 (84%) |                      20/25 (80%) |
| Mean tokens per attempt     |                           30,371 |                           34,664 |
| Mean agent duration         |                          157.3 s |                          171.9 s |
| Mean tool calls per attempt |                             9.64 |                            10.08 |
| 300-second agent timeouts   |                             3/25 |                             1/25 |

| Scenario                    | Score | Full success | Agent timeouts |
| --------------------------- | ----: | -----------: | -------------: |
| `container-crash-loop`      |  0.40 |          2/5 |            0/5 |
| `image-pull-failure`        |  1.00 |          5/5 |            0/5 |
| `missing-rbac-binding`      |  0.70 |          3/5 |            1/5 |
| `service-selector-mismatch` |  1.00 |          5/5 |            0/5 |
| `unschedulable-cpu-request` |  1.00 |          5/5 |            0/5 |

All 25 attempts loaded `troubleshooting` before the first Bash call. Two attempts also loaded `kubernetes`, both before the first Bash call. No attempt loaded `kubectl` or `bash`, and no focused reference was loaded.

#### Trace observations in run 3

- **Crash-loop:** Three of five attempts failed both deployment-rollout and replica-capacity grading. The final deployment state had three ready replicas but only one updated replica. All three final responses claimed recovery; the grader's state check showed that the rollout was incomplete.
- **RBAC:** Three of five attempts passed both criteria. One timed out without restoring either criterion. Another restored application readiness but failed least-privilege grading because the service account could list ConfigMaps; its final response nevertheless described the access as least-privilege.
- **Tool calls:** 39 of 225 Bash calls returned errors. Traces include unsupported interactive `kubectl edit` calls, malformed patch commands, and watch commands that reached the 60-second tool limit.
- **Skill selection:** Kubernetes guidance was loaded in 2/25 attempts, and focused references in 0/25. Both Kubernetes-skill loads occurred in a failed crash-loop attempt and a successful service-selector attempt. These observations are too sparse to attribute outcome differences to skill loading.

The benchmark records incomplete work through grading score, `full_success`, termination, and error fields, so a separate “not done” return value is unnecessary for scoring. However, the failed crash-loop and RBAC attempts show that the final response can claim success when deterministic checks show otherwise.

### Run 4: required domain skills and references

Run `run-2026-09-20-12-04-51-596Z` ran from 2026-09-20 12:04:51 to 12:50:46 UTC. Command: `make benchmark REPEAT=5 AGENT=skill BENCHMARK_PARALLEL=4`. Source revision: `20b0e54ae5eb56baf1ab81090089a0650f35f923`. All 25 attempt records were persisted. The command returned nonzero because eight attempts reached the model deadline and one exceeded the 32,768-token context limit.

| Measure                     | Run 3 broader-guidance condition | Run 4 required-routing condition |
| --------------------------- | -------------------------------: | -------------------------------: |
| Macro-average score         |                            0.820 |                            0.720 |
| Full success                |                      20/25 (80%) |                      17/25 (68%) |
| Mean tokens per attempt     |                           34,664 |                           52,620 |
| Mean agent duration         |                          171.9 s |                          210.8 s |
| Mean tool calls per attempt |                            10.08 |                            12.84 |
| Agent errors                |                             1/25 |                             9/25 |

| Scenario                    | Score | Full success | Agent errors |
| --------------------------- | ----: | -----------: | -----------: |
| `container-crash-loop`      |  0.60 |          3/5 |          1/5 |
| `image-pull-failure`        |  0.80 |          4/5 |          1/5 |
| `missing-rbac-binding`      |  0.70 |          3/5 |          5/5 |
| `service-selector-mismatch` |  1.00 |          5/5 |          0/5 |
| `unschedulable-cpu-request` |  0.50 |          2/5 |          2/5 |

#### Trace observations in run 4

- **Routing:** All 25 attempts loaded `troubleshooting` before Bash. The `kubernetes` skill was loaded in 24/25 attempts, and in 23/25 before the first Bash call. `kubectl` was loaded in all attempts, but only 18 of the 24 attempts that made a mutation loaded it before their first mutation.
- **References:** The agent made 14 `load_reference` calls; 12 succeeded. Successful loads were `configuration.md` (5), `authorization.md` (4), `resources.md` (2), and `workloads.md` (1). No attempt loaded `images.md` or `networking.md`, the listed references for two recurring scenario areas. Two calls used invalid arguments: `skill: "workloads", reference: "deployment.md"` and `skill: "kubernetes", reference: "service"`.
- **Command errors:** 41 of 233 Bash calls returned nonzero status: 38 exited with status 1 and three reached the 60-second tool limit. Traces include interactive `kubectl edit`, open-ended watches, rollout waits longer than the outer tool limit, malformed patches, and incorrect `type/name` resource addressing.
- **Verification:** Two completed responses claimed recovery despite a failed grading criterion. One crash-loop attempt reported all replicas healthy while only one of three replicas was updated. One CPU-scheduling attempt reported a healthy rollout with only two of three replicas present. Three other RBAC attempts passed grading but ended at the agent timeout, so grader success and agent completion are separate outcomes.
- **Runtime:** Eight attempts ended with `context deadline exceeded`; one image-pull attempt failed because the request exceeded the model's 32,768-token context limit. All 25 expected results were still recorded.

Run 4 achieved higher skill-loading rates than run 3, but the score and full-success rate were lower, and command error frequency was similar (41/233 versus 39/225 Bash calls). The comparison is not a causal estimate: the runs were unpaired, the skill condition changed in several ways, and run 4 had more runtime errors.

### Run 5: mandatory skill and reference routing

Run run-2026-09-20-13-03-09-026Z ran from 2026-09-20 13:03:09 to 13:50:42 UTC. Command: make benchmark REPEAT=5 AGENT=skill BENCHMARK_PARALLEL=4. Source revision: 64afafa623af9f81e5d31bb19cd6bef77748defc. All 25 attempt records were persisted. The command returned nonzero after nine model-agent deadlines and a Kind cluster cleanup failure; the leftover cluster was deleted after the run.

| Measure                        |          Run 4 |          Run 5 |
| ------------------------------ | -------------: | -------------: |
| Macro-average normalized score |          0.720 |          0.620 |
| Full success                   |    17/25 (68%) |    14/25 (56%) |
| Mean tokens per attempt        |         52,620 |         66,372 |
| Mean agent duration            |        210.8 s |        219.5 s |
| Mean tool calls per attempt    |          12.84 |          13.64 |
| Bash calls with nonzero status | 41/233 (17.6%) | 31/235 (13.2%) |
| Agent timeouts/errors          |           9/25 |           9/25 |

Run 4 had eight deadline errors and one context-limit error. Run 5 had nine deadline errors; one timed-out attempt nevertheless passed all grading criteria. The runner also reported a separate cluster-deletion failure for one RBAC attempt. The results summary records all 25 attempts.

| Scenario                  | Score | Full success | Agent timeouts |
| ------------------------- | ----: | -----------: | -------------: |
| container-crash-loop      |  0.00 |          0/5 |            3/5 |
| image-pull-failure        |  0.80 |          4/5 |            0/5 |
| missing-rbac-binding      |  0.50 |          1/5 |            4/5 |
| service-selector-mismatch |  1.00 |          5/5 |            0/5 |
| unschedulable-cpu-request |  0.80 |          4/5 |            2/5 |

#### Trace observations in run 5

- **Skill loading:** All 25 attempts loaded troubleshooting and Kubernetes before the first Bash call. Kubectl was loaded before the first mutation in 24 of 25 attempts; 38 of 39 state-changing calls followed a kubectl load.
- **References:** All 28 reference calls loaded a listed file with the correct Kubernetes skill name. Loads were pods.md (6), networking.md (5), resources.md (4), configuration.md (4), authorization.md (4), images.md (2), workloads.md (2), and observability.md (1).
- **Command errors:** 31 of 235 Bash calls returned nonzero status. The count fell from 41/233 in run 4, but invalid calls and failed patches remained, especially in crash-loop repair.
- **Crash-loop:** All five attempts failed both rollout and replica-capacity criteria. Three ended at the model deadline. Two completed responses said Outcome: complete even though grading showed incomplete rollout. Traces include JSON merge patches that replaced the container list without preserving its required image, malformed patch commands, and repeated rollout waits.
- **RBAC:** One of five attempts passed both criteria; the scenario mean score was 0.50 and four attempts timed out. Traces include broad RBAC listings, creation of a separate Role and RoleBinding, and an invalid subject apiGroup patch. The only full-credit attempt repaired the existing RoleBinding's ServiceAccount subject and passed both least-privilege checks.
- **Completion reporting:** Fifteen final responses said Outcome: complete, and two contradicted the grader. No final response said Outcome: incomplete; ten attempts had no outcome status. The benchmark stores grading and termination separately but does not parse the model's outcome sentence as a structured field.

Run 5 improved compliance with skill and reference routing and had fewer failed Bash calls than run 4. Its macro score fell by 0.10, its full-success count fell by three, and mean token use rose by about 26%. The comparison is descriptive: runs are unpaired, the skill instructions changed between runs, and both runs had model errors.

## Interpretation and limitations

Across runs 1–5, skill use changed from optional selection to required routing. Run 5 achieved the specified loading sequence in all 25 attempts and all 28 reference calls used valid Kubernetes filenames. This procedural improvement did not raise task scores: macro score and full-success rate were lower than run 4, while token use was higher. Since the runs were unpaired and had model errors, the difference cannot be attributed to the instruction changes alone.

Run 5 shows that correct skill and reference loading did not ensure a correct repair. Crash-loop performance was 0/5 despite complete skill loading, and RBAC performance remained low. The two false completion claims also show that the final-answer instruction did not reliably prevent unsupported success reports. These traces point to patch selection, repair of the existing RBAC object, and completion reporting as the main areas for researcher review.

This is a development evaluation on five scenarios that were available during prompt and skill development, so performance may be optimistic for these cases. The sample is small; runs 2–4 are unpaired skill-only repetitions after separate changes, and baseline and prompt conditions were not rerun. Runs 4 and 5 each recorded nine agent errors; run 4 included a context-limit error, while all nine run 5 errors were deadlines. Results are descriptive; no inferential statistical analysis was conducted. The model and runtime represent one configuration. Tool-trace observations were reviewed qualitatively.

## Reproducibility and evidence

Raw attempt records and run metadata are preserved with each run:

- [Run 1 raw results](raw/run-2026-09-19-22-00-57-779Z/), including all 75 attempts and `run.json`.
- [Run 1 configuration snapshot](raw/run-2026-09-19-22-00-57-779Z/configuration/), containing the prompt and skill source files used in that run.
- [Run 2 raw results](raw/run-2026-09-20-08-34-18-420Z/), including all 25 attempts and `run.json`.
- [Run 2 configuration snapshot](raw/run-2026-09-20-08-34-18-420Z/configuration/), containing the prompt and skill source files used in that run.
- [Run 3 raw results](raw/run-2026-09-20-11-11-33-707Z/), including all 25 attempts, `run.json`, and the incremental summary in `results.json`.
- [Run 3 configuration snapshot](raw/run-2026-09-20-11-11-33-707Z/configuration/skill/), containing the router prompt and skill and reference files at source revision `13b3526`.
- [Run 4 raw results](raw/run-2026-09-20-12-04-51-596Z/), including all 25 attempts, `run.json`, and `results.json`.
- [Run 4 configuration snapshot](raw/run-2026-09-20-12-04-51-596Z/configuration/skill/), containing the prompt and skill/reference files at source revision `20b0e54`.
- [Run 5 raw results](raw/run-2026-09-20-13-03-09-026Z/), including all 25 attempts, run metadata, and the results summary.
- [Run 5 configuration snapshot](raw/run-2026-09-20-13-03-09-026Z/configuration/skill/), containing the prompt, skills, and references at source revision 64afafa.
- [Skill source traceability](../../../benchmark/internal/agent/skill/SOURCES.md).
- Project-wide skill-routing decision: [D-026](../../decision-log.md).

Runs 2–5 changed the skill condition in successive iterations. The project-wide router change from run 2 is recorded in D-026. Run 5 results and its configuration are archived with the earlier evidence; no raw outcomes have been replaced. No further research or skill change is recorded as approved.
