# RAG development evaluation

## Summary

- `R1`, the RAG condition from #22, scored 0.725 macro (91/138 full successes)
  against 0.634 (78/138) for the prompt agent in the same run (+0.091, CI
  −0.006 to +0.185) and 0.710 (90/138) for the skill agent `S10`, at 56% of the
  skill agent's prompt tokens.
- Retrieval works when used: 73% of searches returned the page the scenario was
  written from. Gemma 4 E4B searched in a third of the attempts.
- Most of the gain is not from the retrieved text. `R1` with an empty index,
  where every search returns nothing, scored 0.700 on the full set: `R1` −
  empty index +0.024 (CI −0.038 to +0.088), empty index − prompt +0.067. About
  three quarters of the gain remains without documentation; on the
  documentation-dependent scenarios the empty index scored as high as `R1`.
- Five attempts to make retrieved text matter (required search, result
  presentation, change guidance, observation-based queries, documentation
  attached to the task) did not improve on `R1`; attaching documentation to the
  task scored lowest (−0.180 against `R1` on ten scenarios).
- Qwen3.5 9B, a stronger model, searched in every `R1` attempt and retrieved the
  source page in 16 of 18, but it is too slow for the benchmark limits, so its
  scores do not show whether the text helped.

## Question

1. Does a `search_docs` tool over the frozen corpus improve repairs over the
   prompt and skill agents?
2. Does the agent retrieve the scenario's source page, and what does the
   retrieved text contribute to the repair?
3. Which candidate should be frozen as the development RAG condition?

## Method

- **Condition.** Prompt agent plus `search_docs` (#22). System prompt: frozen
  prompt plus one search paragraph. No retrieval before the first turn; the
  agent writes its own queries. Retrieval fixed by D-035: hybrid search, 1.5 KiB
  windows, 256 B overlap, top 5, 8 KiB result cap.
- **Runs.** Gemma 4 E4B; runtime, sandbox, scoring and limits (25 turns, 600 s
  per attempt) as in the [skills optimization study](../skills-optimization/README.md).
  Full runs: 46 scenarios × 3 attempts. Screening: that study's 18-scenario
  [subset](../skills-optimization/subset.txt) × 3. Short tests: ten scenarios × 2. Skill agent: archived `S10` run; the same-run prompt agent reproduced that
  study's prompt result (0.634, 78/138).
- **Setup failures.** 15 attempts whose setup failed before the agent started
  were rerun with `RESUME`. Attempts that failed after the agent started are
  kept.
- **Measures.** Macro score (mean per scenario, then over scenarios), full
  success, 95% paired bootstrap CI over scenarios (10,000 resamples, seed 1).
  Evaluator-side, with `make analyze-runs`: search use; retrieval success (a
  returned chunk comes from the scenario's `source.md` page, never shown to the
  model; adjacent pages count as misses); scores grouped by search outcome
  (observational); tokens; `kubectl` writes and failures, `kubectl edit`,
  `kubectl rollout status` (text heuristic).

### Candidates and controls

Files under [`candidates/`](candidates/). None contains scenario answers.

| ID          | Change                                                                                                                                                                                 | Files                                                                                                                             |
| ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `R1`        | Frozen prompt + search once the affected resource is known, before the first change.                                                                                                   | [prompt/r1.md](candidates/prompt/r1.md)                                                                                           |
| empty index | Control: `R1` with every chunk removed from the index; every search returns no results.                                                                                                | [empty.yaml](candidates/empty.yaml)                                                                                               |
| `R2`        | Search required: before the first change and after a failed change.                                                                                                                    | [prompt/r2.md](candidates/prompt/r2.md), [r2.yaml](candidates/r2.yaml)                                                            |
| `R3`        | `R2` + results grouped by page, overlapping windows merged, query guidance in the tool.                                                                                                | [r3.yaml](candidates/r3.yaml), [r3-presentation.patch](candidates/r3-presentation.patch)                                          |
| `R4`        | `R1` + a paragraph on change mechanics and `kubectl rollout status` verification.                                                                                                      | [prompt/r4.md](candidates/prompt/r4.md), [r4.yaml](candidates/r4.yaml)                                                            |
| `P4`        | Control for `R4`: prompt agent + the same paragraph, same invocation as `R4`.                                                                                                          | [prompt/p4.md](candidates/prompt/p4.md), [r4.yaml](candidates/r4.yaml)                                                            |
| `R6`        | `R1` with queries from observed status or errors, retrieved causes checked before acting, each change tied to the results; HTML as plain text and each page's sections in the results. | [prompt/r6.md](candidates/prompt/r6.md), [r6.yaml](candidates/r6.yaml), [r6-presentation.patch](candidates/r6-presentation.patch) |
| `R7`        | `R1` + one search with the task text before the first turn, excerpts appended to the task message.                                                                                     | [r7.yaml](candidates/r7.yaml), [r7-task-search.patch](candidates/r7-task-search.patch)                                            |

## Results

### RAG against prompt and skill

| Condition         | Macro | Full success | Mean turns | Mean prompt tokens |
| ----------------- | ----: | -----------: | ---------: | -----------------: |
| prompt (same run) | 0.634 |       78/138 |       10.4 |             35,450 |
| skill `S10`       | 0.710 |       90/138 |       12.2 |             80,091 |
| RAG `R1`          | 0.725 |       91/138 |       10.8 |             44,684 |

| Comparison     | Macro difference |           95% CI |
| -------------- | ---------------: | ---------------: |
| RAG − prompt   |           +0.091 | [−0.006, +0.185] |
| skill − prompt |           +0.076 | [−0.004, +0.158] |
| RAG − skill    |           +0.015 | [−0.068, +0.091] |

### Search use and retrieval

- 45 searches in 44/138 attempts (32%); 33 before the first change, after a
  median of three tool calls.
- Queries: median five words, resource plus symptom.
- 33/45 searches (73%) returned the source page, 21 at rank 1. Most misses were
  adjacent pages on the same problem.

### What the retrieved text adds: empty-index control

`R1` with an index from which every chunk was removed: the prompt and the tool
are unchanged, and every search returns `No documentation matched the query.`
Full set, three attempts, compared with the `R1` run:

| Condition         | Macro | Full success | Attempts searching | Mean turns | Mean prompt tokens |
| ----------------- | ----: | -----------: | -----------------: | ---------: | -----------------: |
| prompt            | 0.634 |       78/138 |                  — |       10.4 |             35,450 |
| `R1`, empty index | 0.700 |       88/138 |       42/138 (30%) |       10.5 |             36,063 |
| `R1`              | 0.725 |       91/138 |       44/138 (32%) |       10.8 |             44,684 |

| Comparison           | Macro difference |           95% CI |
| -------------------- | ---------------: | ---------------: |
| `R1` − prompt        |           +0.091 | [−0.006, +0.185] |
| empty index − prompt |           +0.067 | [−0.016, +0.144] |
| `R1` − empty index   |           +0.024 | [−0.038, +0.088] |

| Scenario tag                         | Scenarios | Prompt | Empty index |  `R1` | `R1` − prompt, 95% CI   | `R1` − empty index, 95% CI |
| ------------------------------------ | --------: | -----: | ----------: | ----: | ----------------------- | -------------------------- |
| `knowledge: documentation-dependent` |        26 |  0.665 |       0.705 | 0.693 | +0.029 [−0.112, +0.161] | −0.012 [−0.098, +0.080]    |
| `knowledge: general`                 |        11 |  0.616 |       0.778 | 0.778 | +0.162 [+0.020, +0.303] | +0.000 [−0.091, +0.091]    |
| `knowledge: multi-source`            |         6 |  0.591 |       0.519 | 0.736 | +0.145 [−0.189, +0.431] | +0.218 [+0.056, +0.361]    |
| `knowledge: version-specific`        |         3 |  0.519 |       0.741 | 0.778 | +0.259 (3 scenarios)    | +0.037 (3 scenarios)       |

- Without documentation the agent searched as often, used as many turns and
  19% fewer prompt tokens, and kept about three quarters of `R1`'s gain over
  the prompt agent.
- On the 26 documentation-dependent scenarios, where retrieved text should
  matter most, the empty index scored as high as `R1`.
- The multi-source difference (6 scenarios) is mostly not explained by
  retrieval: `R1` retrieved the source page in 5 of its 18 attempts there, and
  the largest per-scenario gains (`sidecar-shared-volume`,
  `dns-name-resolution`) come from attempts that did not retrieve it. Tag
  intervals are not corrected for comparing four tags.
- On the subset the control also ran next to a prompt agent in one
  invocation: empty index − prompt +0.119 [−0.025, +0.265], against `R1` −
  prompt +0.144 [−0.008, +0.299] in the `R1` run; empty index − `R1` −0.007
  [−0.116, +0.090]. The prompt agent scored 0.595 there and 0.578 in the `R1`
  run.

This agrees with where `R1`'s gain lies in its own run: the largest gap to the
prompt agent is in attempts that never searched.

| Group (`R1`, full set)         | Attempts | Macro | Prompt, same scenarios |
| ------------------------------ | -------: | ----: | ---------------------: |
| no search                      |       94 | 0.757 |                  0.651 |
| searched, source not retrieved |       11 | 0.517 |                  0.456 |
| source retrieved               |       33 | 0.617 |                  0.544 |

The search paragraph asks the agent to identify the affected resource before
its first change; the control suggests that this ordering, not the excerpts,
is what improves on the prompt agent.

### Why repairs failed after retrieving the source page

All `R1` and `R2` attempts that retrieved the source page but did not fully
succeed, classified by hand from their traces
([per-attempt list](analysis/failure-causes.md)).

| Cause                                                                                 | `R1` | `R2` | Total |
| ------------------------------------------------------------------------------------- | ---: | ---: | ----: |
| Wrong diagnosis; the query followed the wrong hypothesis                              |    5 |    5 |    10 |
| Right diagnosis; the change was rejected (merge patch, immutable field, wrong object) |    4 |    3 |     7 |
| Right diagnosis; the needed fact was in the retrieved text but not applied            |    2 |    4 |     6 |
| Right diagnosis; wrong or partial fix the retrieved text does not decide              |    4 |   10 |    14 |
| Total                                                                                 |   15 |   22 |    37 |

- `daemonset-missing-toleration`: no failed attempt inspected the node taints;
  queries asked about replicas or `desiredNumberScheduled`. "toleration"
  appeared in the results twice and was not used.
- `networkpolicy-egress-dns`: the caution that a default-deny egress policy
  also blocks DNS was in the results of 4 of 5 failed attempts; the new policy
  still omitted DNS.
- `service-targetport-mismatch`: the `Service` page showing
  `ports[].targetPort` was retrieved; the agent changed the Service type or
  restarted `kube-proxy`.
- The last group needs cluster inspection or a careful reading of the task, not
  documentation: Role rules reduced only in `verbs`, a Deployment scaled below
  the requested replicas to fit a quota.
- No failure was caused by the corpus lacking the fact. At most 16 of 37
  failures (wrong query, fact not applied) are within reach of retrieval.
- Seven scenarios score below 0.5 under prompt, `R1`, `S10` and `P4`:
  `automount-token-disabled`, `container-crash-loop`,
  `daemonset-missing-toleration`, `networkpolicy-cross-namespace`,
  `overprivileged-serviceaccount`, `readonly-rootfs`, `stuck-terminating-pod`.

### Attempts to make the retrieved text matter

Each candidate targets one of the causes above. Short tests use ten scenarios
× 2 attempts: six with retrieval-reachable failures under `R1`
(`daemonset-missing-toleration`, `networkpolicy-egress-dns`,
`service-targetport-mismatch`, `automount-token-disabled`, `readonly-rootfs`,
`stale-image-ifnotpresent`) and four `R1` solved (`pvc-pending`,
`antiaffinity-unschedulable`, `missing-rbac-binding`, `liveness-restart-loop`).
On these ten scenarios `R1` scored 0.567 (16/30), searched in 20/30 attempts
and used 59,438 prompt tokens per attempt.

| Candidate | Targets                       | Scenarios × attempts | Macro (`R1` on the same) | Difference to `R1`, 95% CI | Attempts searching |
| --------- | ----------------------------- | -------------------- | -----------------------: | -------------------------- | -----------------: |
| `R2`      | too few searches              | 46 × 3               |            0.711 (0.725) | −0.014 [−0.094, +0.069]    |             81/138 |
| `R3`      | result presentation           | 18 × 3               |            0.560 (0.722) | inconclusive, see below    |              27/54 |
| `R4`      | rejected changes              | 46 × 3               |            0.687 (0.725) | −0.038 [−0.129, +0.054]    |             36/138 |
| `R6`      | wrong query, fact not applied | 10 × 2               |            0.517 (0.567) | −0.050 [−0.194, +0.083]    |              12/20 |
| `R7`      | no search, wrong hypothesis   | 10 × 2               |            0.387 (0.567) | −0.180 [−0.363, +0.004]    |               4/20 |

All differences to `R1` cross invocations, so differences of about 0.05 are within run-to-run variation.

**`R2`: required search.** Search use nearly doubled (32% → 59% of attempts)
and attempts that retrieved the source page rose from 33 to 50, but their score
equals the prompt agent's on the same scenarios (0.666 vs 0.682). The added
searches were less targeted (57% of searches returned the source page, against
73%), and searches after a failed change were rare.

**`R3`: result presentation (inconclusive).** Attempts without a search also
dropped (0.560 vs 0.624 for the prompt agent), so the format is not the cause;
the run had host memory pressure and a power loss. Not adopted; no full run.

**`R4` and its control `P4`: change guidance.** One paragraph: no editor; a
merge patch replaces whole lists, a strategic merge patch merges `containers`
and `volumes` by name; JSON patch shape; read the rejected field before
retrying; a workload is repaired only when `kubectl rollout status` succeeds.

| Condition   | Macro | Full success | `kubectl` writes failed | `kubectl edit` | Rollout status (attempts) | Completed without full success | Mean prompt tokens |
| ----------- | ----: | -----------: | ----------------------: | -------------: | ------------------------: | -----------------------------: | -----------------: |
| prompt      | 0.634 |       78/138 |           163/375 (43%) |             19 |                        24 |                             57 |             35,450 |
| `P4`        | 0.699 |       85/138 |           171/380 (45%) |              0 |                        40 |                             48 |             35,376 |
| `R1`        | 0.725 |       91/138 |           186/422 (44%) |             17 |                        17 |                             45 |             44,684 |
| `R4`        | 0.687 |       86/138 |           192/403 (48%) |              0 |                        54 |                             48 |             46,203 |
| skill `S10` | 0.710 |       90/138 |           153/341 (45%) |              2 |                        58 |                             44 |             80,091 |

`kubectl edit` disappeared and rollout checks increased, but failed writes did
not decrease. `R4` − `P4` in the same run: −0.012 [−0.093, +0.070]; `P4` −
prompt across runs: +0.065 [−0.009, +0.138].

**`R6`: observation-based queries and outlined results.** The prompt asks for
queries naming the observed status or error, for checking retrieved causes in
the cluster before acting and for naming the result each change relies on;
the results show HTML as plain text and list each page's sections. Some
queries named the symptom (`CrashLoopBackOff Exit Code 1`), others still a
hypothesis (`daemonset replicas default behavior`). No attempt searched twice,
so the section lists never led to a second search; no attempt named the result
a change relied on; no `daemonset-missing-toleration` attempt inspected the
node taints.

**`R7`: documentation attached to the task.** One search with the task text
before the first turn, as in retrieve-then-generate RAG, so the agent sees
documentation before forming a hypothesis; prompt and tool as in `R1`.
Searching each scenario's task text returns the source page for 21 of 46
scenarios (14 at rank 1), and for `daemonset-missing-toleration` and
`networkpolicy-egress-dns` it returns the toleration and the DNS caution the
agent's own queries missed. In the test the attached excerpts replaced the
agent's own searching (4 of 20 attempts searched) and were copied rather than
checked: in `daemonset-missing-toleration` the agent applied the example
manifest URL from the excerpts; in `service-targetport-mismatch`, where the
source page was at rank 5, it guessed port 8080. Scores also dropped on
scenarios `R1` solved (`service-targetport-mismatch` 0/2, `pvc-pending` 1/2).

### Stronger model: Qwen3.5 9B (exploratory)

Prompt agent and `R1` in one invocation, Qwen3.5 9B (`Q4_K_M`, the
[qwen35-9b profile](../../../benchmark/model-profiles/qwen35-9b.env)), subset ×
1 attempt, same limits as for Gemma. Gemma rows are `R1` and its prompt agent
on the same scenarios (three attempts).

| Model, condition | Macro | Full success | Attempts searching | Source retrieved (attempts) | Ended by a limit | Mean turns | Mean prompt tokens |
| ---------------- | ----: | -----------: | -----------------: | --------------------------: | ---------------: | ---------: | -----------------: |
| Gemma, prompt    | 0.578 |        27/54 |                  — |                           — |             0/54 |       11.2 |             39,947 |
| Gemma, `R1`      | 0.722 |        33/54 |              20/54 |                       17/54 |             0/54 |       10.9 |             46,495 |
| Qwen, prompt     | 0.852 |        13/18 |                  — |                           — |             9/18 |       21.3 |            117,437 |
| Qwen, `R1`       | 0.722 |        13/18 |              18/18 |                       16/18 |            15/18 |       21.9 |            164,861 |

- Qwen searched in every `R1` attempt, after inspecting the cluster (a median
  of eight tool calls before the first search; Gemma three), with queries that
  name the observed symptom (`Service targetPort containerPort mismatch`,
  `Kubernetes NetworkPolicy egress DNS port 53`). The same instruction that
  Gemma follows in a third of the attempts is followed in all of them.
- Qwen uses most of the 25-turn and 600-second budget: one command per turn
  and long verification. Search results add context to every later turn, so
  `R1` attempts hit the time limit in 9 of 18 attempts against 1 of 18 for the
  prompt agent. All five `R1` failures ended at a limit; the graded state is the
  cluster when the attempt was stopped.
- `R1` − prompt on Qwen: −0.130 [−0.389, +0.116]. Under these limits the
  difference measures how fast an attempt finishes, not whether the retrieved
  text helped.

### Run-to-run variation

The prompt agent scored 0.578 and 0.595 on the subset in two invocations
(+0.017, CI −0.125 to +0.165). A partial prompt replicate (first attempt of 35
scenarios) scored 0.600 vs 0.645 for the same attempts in the first run.
Cross-invocation differences of about 0.05 are within this variation.

## Conclusions

- **RAG improves on the prompt agent as much as curated skills do, mostly not
  through the documentation.** `R1` is +0.091 over the prompt agent and level
  with `S10` at 56% of its prompt tokens. With an empty index the same agent
  keeps about three quarters of that gain; the retrieved text adds +0.024 (CI
  −0.038 to +0.088). The remaining effect comes from the search paragraph and
  the tool, plausibly because the paragraph makes the agent identify the
  affected resource before its first change; this run does not separate the
  paragraph from the tool.
- **Retrieval quality is not the bottleneck.** The index returns the source
  page for 73% of Gemma's searches and 16 of 18 Qwen attempts.
- **Gemma 4 E4B does not turn retrieved text into repairs.** It searches in a
  third of the attempts, often for a wrong hypothesis, and ignores retrieved
  facts it needs. Instructions to search more, search differently or ground
  changes in the results were followed in part or not at all, and
  documentation attached to the task was copied rather than checked.
- **A stronger model uses the tool as intended.** Qwen3.5 9B searched in every
  attempt with symptom-based queries. Whether the text then improves repairs
  cannot be read from these runs, because Qwen hits the turn and time limits.
- **Candidate.** `R1` remains the best observed candidate; freezing it is
  pending in the [decision log](../../decision-log.md). Its gain over the prompt
  agent should be reported as the effect of the RAG condition as a whole, not
  of retrieved documentation.

## Limitations

- Three attempts per scenario; most differences have CIs that include zero.
- Search outcome groups are observational: the agent searches on harder
  scenarios.
- The full-set empty-index control ran in a different invocation than `R1`; on
  the subset, where both ran next to a prompt agent, the prompt agent's score
  was stable across invocations (+0.017).
- Failure causes were classified by a single rater from keyword matches and
  reading the traces.
- Retrieval success counts only one source page per scenario.
- `S10`, `P4` and `R4` come from other invocations than `R1` and its prompt
  agent.
- Search ordering and change counts use a text heuristic.
- `R3`: subset only, disturbed run. `R6`, `R7`: 20 attempts each on scenarios
  chosen for the change, compared across invocations.
- The `P4` paragraph was written from development traces; it names no
  scenario, resource or fix.
- Qwen: one attempt per scenario on the subset, under limits chosen for Gemma.

## Evidence

Analyses (`report.txt`, `summary.json`, `attempts.jsonl`):

| Analysis                                        | Directory                                                        |
| ----------------------------------------------- | ---------------------------------------------------------------- |
| `R1` vs same-run prompt (full set)              | [analysis/r1/](analysis/r1/)                                     |
| `R1` vs `S10`                                   | [analysis/r1-vs-skill/](analysis/r1-vs-skill/)                   |
| Empty index vs `R1` (full set)                  | [analysis/empty-full-vs-r1/](analysis/empty-full-vs-r1/)         |
| Empty index vs `R1`'s prompt agent (full set)   | [analysis/empty-full-vs-prompt/](analysis/empty-full-vs-prompt/) |
| Empty index vs same-run prompt (subset)         | [analysis/empty-subset/](analysis/empty-subset/)                 |
| Empty index vs `R1` (subset)                    | [analysis/empty-vs-r1/](analysis/empty-vs-r1/)                   |
| Failure causes after retrieving the source page | [analysis/failure-causes.md](analysis/failure-causes.md)         |
| `R1`, `R2`, `R3` vs prompt (subset)             | [analysis/subset/](analysis/subset/)                             |
| `R2` vs same-run prompt                         | [analysis/r2/](analysis/r2/)                                     |
| `P4` vs prompt                                  | [analysis/p4/](analysis/p4/)                                     |
| `R4` vs `P4` (same run)                         | [analysis/r4/](analysis/r4/)                                     |
| `R4` vs `R1`                                    | [analysis/r4-vs-r1/](analysis/r4-vs-r1/)                         |
| `P4`, `R4`, prompt vs `S10`                     | [analysis/vs-skill/](analysis/vs-skill/)                         |
| `R6` vs `R1` (ten scenarios)                    | [analysis/r6-short/](analysis/r6-short/)                         |
| `R7` vs `R1` (ten scenarios)                    | [analysis/r7-short/](analysis/r7-short/)                         |
| Qwen3.5 9B `R1` vs prompt (subset)              | [analysis/qwen9b-subset/](analysis/qwen9b-subset/)               |
| Prompt replicate vs first prompt run            | [analysis/prompt-replicate/](analysis/prompt-replicate/)         |

Raw runs (`run.json`, `results.json`, per-attempt records, configuration):

| Run                               | Run ID                         | Raw results                                                        |
| --------------------------------- | ------------------------------ | ------------------------------------------------------------------ |
| `R1` + prompt, full set           | `run-2026-10-02-07-35-43-038Z` | [raw/r1/](raw/r1/)                                                 |
| Empty index, full set             | `run-2026-10-04-04-37-44-883Z` | [raw/empty-full/](raw/empty-full/)                                 |
| Empty index + prompt, subset      | `run-2026-10-04-02-32-31-899Z` | [raw/empty-subset/](raw/empty-subset/)                             |
| `R2`, full set                    | `run-2026-10-02-17-28-03-947Z` | [raw/r2/](raw/r2/)                                                 |
| `R2`, subset                      | `run-2026-10-02-10-32-47-138Z` | [raw/r2-subset/](raw/r2-subset/)                                   |
| `R3`, subset                      | `run-2026-10-02-12-21-54-157Z` | [raw/r3-subset/](raw/r3-subset/)                                   |
| `P4` + `R4`, full set             | `run-2026-10-02-22-39-05-228Z` | [raw/p4-r4/](raw/p4-r4/)                                           |
| `R6`, ten scenarios               | `run-2026-10-03-21-54-29-642Z` | [raw/r6-short/](raw/r6-short/)                                     |
| `R7`, ten scenarios               | `run-2026-10-04-01-51-51-746Z` | [raw/r7-short/](raw/r7-short/)                                     |
| Qwen3.5 9B, prompt + `R1`, subset | `run-2026-10-03-23-40-15-673Z` | [raw/qwen9b-subset/](raw/qwen9b-subset/)                           |
| Prompt replicate (35)             | `run-2026-10-02-21-41-02-671Z` | [raw/prompt-replicate/](raw/prompt-replicate/)                     |
| skill `S10`                       | `run-2026-09-26-00-21-23-444Z` | [../skills-optimization/raw/s10/](../skills-optimization/raw/s10/) |

Reproduce from `benchmark/` (link raw directories back to their run IDs under
`results/` first):

```sh
make analyze-runs ANALYZE_ARGS='--condition run-2026-10-02-07-35-43-038Z/rag --reference run-2026-10-02-07-35-43-038Z/prompt'
```
