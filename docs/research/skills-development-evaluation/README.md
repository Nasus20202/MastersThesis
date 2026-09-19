# Skills development evaluation

## Status

Active development study for tickets [#15](https://github.com/Nasus20202/MastersThesis/issues/15), [#16](https://github.com/Nasus20202/MastersThesis/issues/16), [#17](https://github.com/Nasus20202/MastersThesis/issues/17), and [#18](https://github.com/Nasus20202/MastersThesis/issues/18). All four remain **In progress**. The first approved 75-attempt run and its analysis are complete. The router-only follow-up approved in [D-026](../../decision-log.md) is implemented; its rerun is deferred.

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

The benchmark runner limits active agent loops to the two configured inference slots. Cluster setup, grading and cleanup remain outside that limit. This run is the first measurement with that limit. The run used repository revision `63ad7ef3fb2c270486a7e3a1d1ccdab68f0adbe0`; temperature and maximum output tokens were unset, so runtime defaults applied. The command was `make benchmark REPEAT=5 AGENT=all BENCHMARK_PARALLEL=4`.

## Measures and analysis

The primary measure is macro-average normalized score: average the five repetitions within each scenario, then average scenarios equally. Secondary measures are full-success rate, per-scenario outcome distribution, tokens, tool use, agent duration, termination reason, inference errors and observable failure modes. The complete run, including all 75 attempt records and `run.json`, is preserved in [`raw/run-2026-09-19-22-00-57-779Z/`](raw/run-2026-09-19-22-00-57-779Z/).

The five scenarios were available while the prompt and skill were developed. Results are development evidence, not a generalization estimate. They must not be used to alter a recorded result or claim final benchmark performance.

## Results

Run `run-2026-09-19-22-00-57-779Z` started at 2026-09-19 22:00:57 UTC and completed at 23:32:41 UTC. All 75 attempts completed and were persisted. A full success means every deterministic grading criterion passed. The score is the normalized partial score, macro-averaged equally across the five scenarios.

| Condition | Macro score | Full success | Total tokens (mean per attempt) | Tool calls                | Mean agent duration | 300-second timeouts |
| --------- | ----------: | -----------: | ------------------------------: | ------------------------- | ------------------: | ------------------: |
| Baseline  |       0.680 |  16/25 (64%) |                371,544 (14,862) | 166 Bash                  |             115.7 s |                0/25 |
| Prompt    |       0.680 |  17/25 (68%) |                578,121 (23,125) | 221 Bash                  |             138.7 s |                4/25 |
| Skill     |       0.740 |  17/25 (68%) |                827,193 (33,088) | 230 Bash; 12 `load_skill` |             165.2 s |                4/25 |

Each of the four prompt and four skill timeouts ended with `context deadline exceeded`. One skill timeout still passed every grading criterion, so agent termination and task success are reported separately.

| Scenario                    | Baseline score (full success) | Prompt score (full success) | Skill score (full success) |
| --------------------------- | ----------------------------: | --------------------------: | -------------------------: |
| `container-crash-loop`      |                    0.40 (2/5) |                  0.20 (1/5) |                 0.20 (1/5) |
| `image-pull-failure`        |                    1.00 (5/5) |                  1.00 (5/5) |                 1.00 (5/5) |
| `missing-rbac-binding`      |                    0.20 (0/5) |                  0.40 (2/5) |                 0.50 (1/5) |
| `service-selector-mismatch` |                    1.00 (5/5) |                  1.00 (5/5) |                 1.00 (5/5) |
| `unschedulable-cpu-request` |                    0.80 (4/5) |                  0.80 (4/5) |                 1.00 (5/5) |

The skill condition scored 0.06 above both baseline and prompt, and completed one more task than baseline. Its full-success rate matched prompt at 68%. The score gain came mainly from RBAC and scheduling; crash-loop performance was lower than baseline, and the two consistently solved scenarios were already at ceiling in all conditions. With five repetitions per scenario, these are descriptive results and do not establish a stable general improvement.

## Skill selection and use

| Observation in the 25 skill attempts             |                 Count |
| ------------------------------------------------ | --------------------: |
| Loaded at least one skill                        |                 12/25 |
| Loaded no skill                                  |                 13/25 |
| `load_skill` calls                               | 12, all for `kubectl` |
| Loaded `troubleshooting`, `kubernetes` or `bash` |                     0 |
| Loaded a reference                               |                     0 |

The 12 `kubectl` loads occurred before the first Bash call in those attempts. The troubleshooting procedure and Kubernetes knowledge references were never exposed to the model. Therefore this run does not measure the effect of the troubleshooting skill content; it measures the skill condition as implemented, where skill selection is optional, with sparse use of the `kubectl` skill. The aggregate skill score cannot be attributed to the unused troubleshooting material.

Attempts that loaded `kubectl` averaged 0.625 score and 6/12 full successes; attempts that loaded no skill averaged 0.846 score and 11/13 full successes. This is a descriptive split only: the model chose whether to load a skill, so the groups are not randomized and should not be interpreted as evidence that loading `kubectl` reduced performance.

## Failure patterns

- **Crash loop:** Skill achieved full success in 1/5 attempts, versus 2/5 for baseline and 1/5 for prompt. In the four failed skill attempts, the agent found and patched the invalid Nginx arguments but the graders observed only `1/3` replicas updated (three were ready); rollout checks timed out. The final responses sometimes claimed recovery before the required updated-replica condition was met. Several traces used unbounded pod watches or rollout checks that exceeded the 60-second tool limit.
- **RBAC:** Skill achieved full success in 1/5 attempts. Three attempts received partial credit because the deployment recovered but the least-privilege criterion failed. Their repairs added permissions including `list` (and sometimes `watch`) on ConfigMaps, while the verifier requires access to the named ConfigMap and denies listing ConfigMaps and reading Secrets. The one full repair corrected the existing RoleBinding's ServiceAccount subject and passed both criteria.
- **Timeouts:** Ten skill tool calls across four attempts exceeded the 60-second tool limit. Calls included `kubectl get ... --watch` and rollout checks. Four skill attempts also reached the 300-second agent limit; the scheduling attempt still passed grading despite timing out.

## Interpretation

This run estimates the skill condition with optional skill routing as implemented. Since troubleshooting was loaded in none of the attempts, the results do not estimate the effect of applying its procedure. The observed outcomes and use traces are preserved above for researcher review.

## Follow-up iteration placeholder

**Approved change:** require loading the `troubleshooting` skill before the first Bash call on troubleshooting tasks. Keep skill content and all other benchmark settings fixed, as recorded in [D-026](../../decision-log.md).

**Status:** Router change implemented in [`benchmark/internal/agent/skill/prompt.md`](../../../benchmark/internal/agent/skill/prompt.md); benchmark rerun pending at the researcher's request.

**Planned run:** the same 3-condition × 5-scenario × 5-repetition matrix (75 attempts). Run ID: pending. Raw evidence: pending under `raw/<run-id>/`. Analysis: pending.

## Iteration rule

After each complete 75-attempt run, analyze the preserved evidence and present any proposed prompt, skill, execution or protocol change to the researcher. No change will be made or measured until it is explicitly approved. An approved change starts a new complete 75-attempt run; earlier evidence remains labeled with the configuration that produced it.

## Sources

The skill package's source traceability is maintained in [`benchmark/internal/agent/skill/SOURCES.md`](../../../benchmark/internal/agent/skill/SOURCES.md). Common benchmark and scoring decisions are recorded in the [decision log](../../decision-log.md).
