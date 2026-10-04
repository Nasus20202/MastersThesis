# RAG development evaluation

## Summary

- `R1` (the RAG condition from #22): 0.725 macro, 91/138 full successes.
  Prompt agent in the same run: 0.634, 78/138 (+0.091, CI −0.006 to +0.185).
  Skill agent `S10`: 0.710, 90/138 (+0.015, CI −0.068 to +0.091).
- Retrieval works: 73% of searches returned the scenario's source page.
- The agent searched in 32% of attempts. Most of the gain over the prompt agent
  is in attempts without a search.
- `R2` (required search), `R3` (result presentation), `R4` (change guidance)
  and `R6` (observation-based queries, outlined results) did not improve on
  `R1`. Gemma 4 E4B follows few of the added search instructions.

## Question

1. Does a `search_docs` tool over the frozen corpus improve repairs over the
   prompt and skill agents?
2. Does the agent retrieve the scenario's source page, and does that help?
3. Which candidate should be frozen as the development RAG condition?

## Method

- **Condition.** Prompt agent plus `search_docs` (#22). System prompt: frozen
  prompt plus one search paragraph. No retrieval before the first turn; the
  agent writes its own queries. Retrieval fixed by D-035: hybrid search, 1.5 KiB
  windows, 256 B overlap, top 5, 8 KiB result cap.
- **Runs.** Model, runtime, sandbox, scoring and limits as in the
  [skills optimization study](../skills-optimization/README.md). Full runs: 46
  scenarios × 3 attempts. Screening: that study's 18-scenario
  [subset](../skills-optimization/subset.txt) × 3. Skill agent: archived `S10`
  run; the same-run prompt agent reproduced that study's prompt result
  (0.634, 78/138).
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

### Candidates

Files under [`candidates/`](candidates/). None contains scenario answers.

| ID   | Change                                                                                                                                                                               | Files                                                                                                                             |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------- |
| `R1` | Frozen prompt + search once the affected resource is known, before the first change.                                                                                                 | [prompt/r1.md](candidates/prompt/r1.md)                                                                                           |
| `R2` | Search required: before the first change and after a failed change.                                                                                                                  | [prompt/r2.md](candidates/prompt/r2.md), [r2.yaml](candidates/r2.yaml)                                                            |
| `R3` | `R2` + results grouped by page, overlapping windows merged, query guidance in the tool.                                                                                              | [r3.yaml](candidates/r3.yaml), [r3-presentation.patch](candidates/r3-presentation.patch)                                          |
| `R4` | `R1` + a paragraph on change mechanics and `kubectl rollout status` verification.                                                                                                    | [prompt/r4.md](candidates/prompt/r4.md), [r4.yaml](candidates/r4.yaml)                                                            |
| `P4` | Control for `R4`: prompt agent + the same paragraph, same invocation as `R4`.                                                                                                        | [prompt/p4.md](candidates/prompt/p4.md), [r4.yaml](candidates/r4.yaml)                                                            |
| `R6` | `R1` with queries from observed status or errors, retrieved causes checked before acting, each change tied to the results; results with HTML as plain text and each page's sections. | [prompt/r6.md](candidates/prompt/r6.md), [r6.yaml](candidates/r6.yaml), [r6-presentation.patch](candidates/r6-presentation.patch) |
| `R7` | `R1` (same prompt and tool) + one search with the task text before the first turn, excerpts appended to the task message.                                                            | [r7.yaml](candidates/r7.yaml), [r7-task-search.patch](candidates/r7-task-search.patch)                                            |

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

- 45 searches in 44/138 attempts (32%); 33 before the first change.
- Queries: median five words, resource plus symptom.
- 33/45 searches (73%) returned the source page, 21 at rank 1. Most misses were
  adjacent pages on the same problem.

### Where the gain comes from

| Group (`R1`)                   | Attempts | Macro | Prompt, same scenarios |
| ------------------------------ | -------: | ----: | ---------------------: |
| no search                      |       94 | 0.757 |                  0.651 |
| searched, source not retrieved |       11 | 0.517 |                  0.456 |
| source retrieved               |       33 | 0.617 |                  0.544 |

| Scenario tag                         | Scenarios | Prompt |   RAG | Skill | RAG − prompt, 95% CI    |
| ------------------------------------ | --------: | -----: | ----: | ----: | ----------------------- |
| `knowledge: documentation-dependent` |        26 |  0.665 | 0.693 | 0.715 | +0.029 [−0.112, +0.161] |
| `knowledge: general`                 |        11 |  0.616 | 0.778 | 0.788 | +0.162 [+0.020, +0.303] |
| `knowledge: multi-source`            |         6 |  0.591 | 0.736 | 0.624 | +0.145 [−0.189, +0.431] |
| `knowledge: version-specific`        |         3 |  0.519 | 0.778 | 0.556 | +0.259 (3 scenarios)    |

- The largest gap to the prompt agent (+0.106) is in attempts without a search.
  It may come from the search paragraph, the tool's presence or run-to-run
  variation; the runs do not separate these.
- The gain is small on documentation-dependent scenarios.
- Seven scenarios score below 0.5 under prompt, `R1`, `S10` and `P4`:
  `automount-token-disabled`, `container-crash-loop`,
  `daemonset-missing-toleration`, `networkpolicy-cross-namespace`,
  `overprivileged-serviceaccount`, `readonly-rootfs`, `stuck-terminating-pod`.

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
- No failure was caused by the corpus lacking the fact.
- At most 16 of 37 failures (wrong query, fact not applied) are within reach of
  retrieval; the other 21 are shared with the prompt agent.

### `R2`: required search

| Condition | Macro | Full success | Attempts searching | Source in top 5 (searches) | Mean prompt tokens |
| --------- | ----: | -----------: | -----------------: | -------------------------: | -----------------: |
| `R1`      | 0.725 |       91/138 |       44/138 (32%) |                33/45 (73%) |             44,684 |
| `R2`      | 0.711 |       89/138 |       81/138 (59%) |                56/98 (57%) |             51,760 |

- `R2` − `R1`: −0.014 [−0.094, +0.069].
- Attempts that retrieved the source page: 33 → 50; their score equals the
  prompt agent's on the same scenarios (0.666 vs 0.682).
- Searches after a failed change were rare.

### `R3`: result presentation (inconclusive)

- Subset: `R3` 0.560 (27/54), `R1` 0.722 (33/54), `R2` 0.685 (34/54).
- Attempts without a search also dropped (0.560 vs 0.624 for the prompt agent),
  so the format is not the cause. The run had host memory pressure and a power
  loss. Not adopted; no full run.

### `R4` and its control `P4`: change guidance

The paragraph: no editor; a merge patch replaces whole lists, a strategic merge
patch merges `containers` and `volumes` by name; JSON patch shape; read the
rejected field before retrying; a workload is repaired only when
`kubectl rollout status` succeeds.

| Condition   | Macro | Full success | `kubectl` writes failed | `kubectl edit` | Rollout status (attempts) | Completed without full success | Mean prompt tokens |
| ----------- | ----: | -----------: | ----------------------: | -------------: | ------------------------: | -----------------------------: | -----------------: |
| prompt      | 0.634 |       78/138 |           163/375 (43%) |             19 |                        24 |                             57 |             35,450 |
| `P4`        | 0.699 |       85/138 |           171/380 (45%) |              0 |                        40 |                             48 |             35,376 |
| `R1`        | 0.725 |       91/138 |           186/422 (44%) |             17 |                        17 |                             45 |             44,684 |
| `R4`        | 0.687 |       86/138 |           192/403 (48%) |              0 |                        54 |                             48 |             46,203 |
| skill `S10` | 0.710 |       90/138 |           153/341 (45%) |              2 |                        58 |                             44 |             80,091 |

| Comparison    | Macro difference |           95% CI |
| ------------- | ---------------: | ---------------: |
| `R4` − `P4`   |           −0.012 | [−0.093, +0.070] |
| `R4` − `R1`   |           −0.038 | [−0.129, +0.054] |
| `P4` − prompt |           +0.065 | [−0.009, +0.138] |

- `kubectl edit` disappeared and rollout checks increased; failed writes did
  not decrease.
- `R4` searched in 26% of attempts and scored at `P4`'s level.
- `P4` − prompt and `R4` − `R1` cross invocations.

### `R6`: observation-based queries and outlined results (short test)

Aimed at the two retrieval-reachable causes above. Ten scenarios × 2 attempts:
six with retrieval-reachable failures under `R1`
(`daemonset-missing-toleration`, `networkpolicy-egress-dns`,
`service-targetport-mismatch`, `automount-token-disabled`, `readonly-rootfs`,
`stale-image-ifnotpresent`) and four `R1` solved (`pvc-pending`,
`antiaffinity-unschedulable`, `missing-rbac-binding`, `liveness-restart-loop`).

| Condition       | Macro | Full success | Attempts searching | Mean prompt tokens |
| --------------- | ----: | -----------: | -----------------: | -----------------: |
| `R1` (full run) | 0.567 |        16/30 |              20/30 |             59,438 |
| `R6`            | 0.517 |         9/20 |              12/20 |             72,718 |

- `R6` − `R1`: −0.050 [−0.194, +0.083], across invocations.
- Some queries named the observed symptom (`CrashLoopBackOff Exit Code 1`),
  others still a hypothesis (`daemonset replicas default behavior`).
- No `daemonset-missing-toleration` attempt inspected the node taints.
- No attempt searched twice, so the page sections never led to a second
  search; no attempt stated which result a change relied on.
- Not run on the full set.

### `R7`: documentation retrieved with the task (short test)

Retrieval before the first turn, as in classic retrieve-then-generate RAG, so
that the agent sees documentation before forming a hypothesis. Searching each
scenario's task text returns the source page for 21 of 46 scenarios (14 at
rank 1); the results for `daemonset-missing-toleration` and
`networkpolicy-egress-dns` contain the toleration and the DNS caution that the
agent's own queries missed. Same ten scenarios × 2 attempts as `R6`.

| Condition       | Macro | Full success | Attempts searching (`search_docs`) | Mean prompt tokens |
| --------------- | ----: | -----------: | ---------------------------------: | -----------------: |
| `R1` (full run) | 0.567 |        16/30 |                              20/30 |             59,438 |
| `R7`            | 0.387 |         7/20 |                               4/20 |             88,204 |

- `R7` − `R1`: −0.180 [−0.363, +0.004], across invocations; scores dropped on
  scenarios `R1` solved (`service-targetport-mismatch` 0/2, `pvc-pending` 1/2).
- The appended excerpts replaced the agent's own searching.
- They were copied rather than checked: in `daemonset-missing-toleration` the
  agent applied the example manifest URL from the excerpts, in
  `service-targetport-mismatch`, where the source page was only at rank 5, it
  guessed port 8080.
- Not run on the full set.

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

- Qwen searched in every `R1` attempt, after inspecting the cluster, with
  queries that name the observed symptom (`Service targetPort containerPort
mismatch`, `Kubernetes NetworkPolicy egress DNS port 53`). The search
  instruction that Gemma follows in a third of attempts is followed by Qwen in
  all of them.
- Qwen uses most of the 25-turn and 600-second budget: one command per turn,
  long verification. Search results add context to every later turn, so `R1`
  attempts hit the time limit in 9 of 18 attempts against 1 of 18 for the
  prompt agent. All five `R1` failures ended at a limit; the graded state is
  the cluster when the attempt was stopped.
- `R1` − prompt on Qwen: −0.130 [−0.389, +0.116]. Under these limits the
  difference measures how fast an attempt finishes, not whether the retrieved
  text helped.

### Run-to-run variation

A prompt replicate (first attempt of 35 scenarios) scored 0.600 vs 0.645 for
the same attempts in the first run (−0.058 against all three attempts, CI
−0.196 to +0.071). Cross-invocation differences of about 0.05 are noise.

## Conclusions

- RAG matches the skill agent and beats the prompt agent by +0.091, at 56% of
  the skill agent's prompt tokens; the CI includes zero.
- Retrieval quality is not the bottleneck: the source page is found in 73% of
  searches.
- The gain is not clearly due to retrieved text: it lies in attempts without a
  search, and forcing searches (`R2`) did not help.
- Remaining failures are in diagnosis, applying changes and verifying them. Of
  the failures that retrieved the source page, 10 of 37 searched for the wrong
  hypothesis and 6 of 37 did not apply a retrieved fact; the corpus was never
  the missing piece.
- Prompt instructions are a weak lever for this model: requiring searches
  (`R2`), change guidance (`R4`) and observation-based, grounded searching
  (`R6`) were followed in part or not at all.
- `R1` is the best observed candidate; freezing it is pending in the
  [decision log](../../decision-log.md).

## Limitations

- Three attempts per scenario; most differences have CIs that include zero.
- The prompt → RAG difference mixes retrieved text, the search paragraph and
  the tool. No empty-index control was run.
- Search outcome groups are observational: the agent searches on harder
  scenarios.
- Failure causes were classified by a single rater from keyword matches and reading the traces.
- Retrieval success counts only one source page per scenario.
- `S10`, `P4` and `R4` come from other invocations than `R1` and its prompt
  agent.
- Search ordering and change counts use a text heuristic.
- `R3`: subset only, disturbed run.
- `R6`, `R7`: 20 attempts each on scenarios chosen for the change, compared
  across invocations.
- The `P4` paragraph was written from development traces; it names no
  scenario, resource or fix.

## Evidence

Analyses (`report.txt`, `summary.json`, `attempts.jsonl`):

| Analysis                                        | Directory                                                |
| ----------------------------------------------- | -------------------------------------------------------- |
| `R1` vs same-run prompt (full set)              | [analysis/r1/](analysis/r1/)                             |
| `R1` vs `S10`                                   | [analysis/r1-vs-skill/](analysis/r1-vs-skill/)           |
| `R2` vs same-run prompt                         | [analysis/r2/](analysis/r2/)                             |
| `R1`, `R2`, `R3` vs prompt (subset)             | [analysis/subset/](analysis/subset/)                     |
| `P4` vs prompt                                  | [analysis/p4/](analysis/p4/)                             |
| `R4` vs `P4` (same run)                         | [analysis/r4/](analysis/r4/)                             |
| `R4` vs `R1`                                    | [analysis/r4-vs-r1/](analysis/r4-vs-r1/)                 |
| `P4`, `R4`, prompt vs `S10`                     | [analysis/vs-skill/](analysis/vs-skill/)                 |
| Prompt replicate vs first prompt run            | [analysis/prompt-replicate/](analysis/prompt-replicate/) |
| `R7` vs `R1` (ten scenarios)                    | [analysis/r7-short/](analysis/r7-short/)                 |
| Qwen3.5 9B `R1` vs prompt (subset)              | [analysis/qwen9b-subset/](analysis/qwen9b-subset/)       |
| `R6` vs `R1` (ten scenarios)                    | [analysis/r6-short/](analysis/r6-short/)                 |
| Failure causes after retrieving the source page | [analysis/failure-causes.md](analysis/failure-causes.md) |

Raw runs (`run.json`, `results.json`, per-attempt records, configuration):

| Run                               | Run ID                         | Raw results                                                        |
| --------------------------------- | ------------------------------ | ------------------------------------------------------------------ |
| `R1` + prompt, full set           | `run-2026-10-02-07-35-43-038Z` | [raw/r1/](raw/r1/)                                                 |
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
