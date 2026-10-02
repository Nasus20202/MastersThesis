# RAG development evaluation

## Status

`R1`, the RAG condition as implemented in #22, is proposed for freezing as the
development RAG condition; the decision is pending approval. On the full
46-scenario benchmark it reached **0.725 macro score and 91/138 full
successes**, against 0.634 and 78/138 for the prompt agent in the same run
(+0.091, 95% CI −0.006 to +0.185) and 0.710 and 90/138 for the frozen skill
agent (+0.015, CI −0.068 to +0.091), at 56% of the skill agent's prompt tokens.
Retrieval itself works: 73% of searches returned the scenario's source page.
The agent, however, searched in only a third of the attempts, and neither a
stronger search instruction (`R2`) nor a different result presentation (`R3`)
improved outcomes.

## Question

Does giving the prompt agent a `search_docs` tool over the frozen documentation
corpus improve repairs, and how often does the agent search and retrieve the
documentation page each scenario was written from? The study separates
failures to retrieve the right page from failures to use it, compares RAG with
the prompt agent (what retrieval adds) and with the frozen skill agent
(retrieval versus curated skills), and selects the development RAG condition to
freeze.

## Method

### Condition

The RAG agent is the prompt agent with one extra tool (#22). Its system prompt
is the frozen prompt followed by search instructions, so the prompt → RAG gap
isolates retrieval. No context is retrieved before the first turn: the agent
writes every query from what it has observed. The retrieval configuration is
fixed by the retrieval design study (D-035): hybrid search over 1.5 KiB windows
with 256 B overlap, top 5, model-visible result capped at 8 KiB. Ranking and
the index are not changed here; candidates change only the search instructions
and how results are presented.

### Scenarios, runs and outcome measures

Runs use the 46 development scenarios with three attempts each (138 attempts
per condition) and the same model, runtime, sandbox, scoring and agent limits as
the [skills optimization study](../skills-optimization/README.md): Gemma 4 E4B
QAT `UD-Q4_K_XL` with multi-token prediction, 25 turns, 50 tool calls, 60
seconds per tool call, 600 seconds per attempt and 8 KiB of model-visible Bash
output. Candidates after the first full run are screened on the same
18-scenario subset as that study ([subset.txt](../skills-optimization/subset.txt)).
The primary measure is the macro-average score (mean score per scenario, then
the unweighted mean over scenarios); full success requires every grading
criterion. Attempts that end in an error after the agent started stay in the
denominator with their recorded score.

Attempts whose scenario setup failed before the agent started (cluster
preparation timed out under load, or the registry cache served an expired
Traefik layer as an empty blob) say nothing about the agent. Unlike the earlier
studies, whose archived runs contain none, these runs had several. Such
attempts were deleted and rerun with `RESUME` on the same code and
configuration; the analysis reports how many attempts ran without an agent.

### Candidates

Each candidate is a complete prompt file or code patch under
[`candidates/`](candidates/). None contains scenario answers; example queries
in prompts name resources no scenario uses.

| ID   | Change from the previous candidate                                                                                                                                         | Files                                                                                    |
| ---- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `R1` | Condition as implemented in #22: frozen prompt plus one paragraph asking for a search once the affected resource is identified and before the first change.                | [prompt/r1.md](candidates/prompt/r1.md)                                                  |
| `R2` | Searching is stated as a required step with two triggers: before the first change, and after a change that fails or does not remove the symptom.                           | [prompt/r2.md](candidates/prompt/r2.md), [r2.yaml](candidates/r2.yaml)                   |
| `R3` | `R2` prompt; results grouped by page (title and path once, overlapping windows merged, heading anchors removed) and a tool description that says what a query should name. | [r3.yaml](candidates/r3.yaml), [r3-presentation.patch](candidates/r3-presentation.patch) |

### Search and retrieval measures

Scored afterwards, evaluator-side only, with `make analyze-runs`:

- **Search use:** attempts with at least one `search_docs` call, number of
  searches, and whether a search came before the first Bash command that
  changes the cluster (a `kubectl` write subcommand such as `apply`, `patch`,
  `set`, `delete`, `scale` or `rollout restart`; a text heuristic).
- **Retrieval success:** a search succeeds when one of its returned chunks
  comes from the scenario's `source.md` path, the page the scenario was written
  from. The path is never shown to the model. An attempt _retrieved the
  source_ when any of its searches did. This is file-level and strict: other
  pages that also describe the fix count as misses.
- **Effect of searching:** attempts are split into _no search_, _searched,
  source not retrieved_ and _source retrieved_, and each group is compared with
  the prompt agent on the same scenarios, as done for skill loading. The groups
  are chosen by the agent, so this is an association, not a causal effect.
- **Cost:** mean turns, prompt tokens and completion tokens per attempt, over
  attempts where the agent ran, next to the prompt agent on the same scenarios.
  Search results enter the context, so this shows what retrieval costs.
- **Per-criterion pass rate:** the share of attempts passing each grading
  criterion of each scenario, next to the prompt agent's rate. Criterion IDs
  repeat across scenarios, so rates are kept per scenario. It shows which part
  of the repair retrieval helped or hurt.

## Results

### Candidate screening on the subset

Three attempts on each of the 18 subset scenarios (54 attempts per candidate).
`R1` and the prompt agent are their full runs restricted to these scenarios.

| Candidate | Macro score | Full success | Attempts searching | Searches | Source retrieved (attempts) |
| --------- | ----------: | -----------: | -----------------: | -------: | --------------------------: |
| prompt    |       0.578 |        27/54 |                  — |        — |                           — |
| `R1`      |   **0.722** |        33/54 |              20/54 |       20 |                       17/54 |
| `R2`      |       0.685 |    **34/54** |          **34/54** |       38 |                       26/54 |
| `R3`      |       0.560 |        27/54 |              27/54 |       33 |                       21/54 |

Stating the search as a required step (`R2`) raised the share of attempts that
searched from 37% to 63% and the number of attempts that retrieved their
scenario's source page from 17 to 26, but did not raise the score: `R2` is
0.037 below `R1` with one more full success, a difference well within the
per-scenario noise of three attempts. The second trigger, searching after a
failed change, was rarely followed (8 of 38 `R2` searches came after the first
change). `R3`'s presentation change scored lowest, but the drop also appears in
the attempts that never searched (0.560 against 0.726 for `R2` on the same
scenarios), where the result format cannot matter, and this run was exposed to
host memory pressure and a power loss. The data therefore give no support for
the presentation change rather than evidence against it; it was not adopted.

### Full benchmark: RAG against the prompt and skill agents

`R1` and the prompt agent ran in the same invocation (46 scenarios, three
attempts, 138 attempts each). The skill agent is the frozen `S10` run from the
skills optimization study; the same-run prompt agent reproduced that study's
prompt result exactly (0.634, 78/138), so the archived skill run is a fair
reference. Differences are macro-score differences with a 95% paired bootstrap
interval over the 46 scenarios (10,000 resamples, seed 1).

| Condition         | Macro score |      Full success | Mean turns | Mean prompt tokens | Mean completion tokens |
| ----------------- | ----------: | ----------------: | ---------: | -----------------: | ---------------------: |
| prompt (same run) |       0.634 |     78/138 (0.57) |       10.4 |             35,450 |                  2,989 |
| skill `S10`       |       0.710 |     90/138 (0.65) |       12.2 |             80,091 |                  3,229 |
| RAG `R1`          |   **0.725** | **91/138 (0.66)** |       10.8 |             44,684 |                  3,142 |

| Comparison     | Macro difference |           95% CI |
| -------------- | ---------------: | ---------------: |
| RAG − prompt   |           +0.091 | [−0.006, +0.185] |
| skill − prompt |           +0.076 | [−0.004, +0.158] |
| RAG − skill    |           +0.015 | [−0.068, +0.091] |

RAG adds 0.091 macro score and 13 full successes over the prompt agent at 26%
more prompt tokens per attempt; the interval just includes zero. It matches
the skill agent's outcome at 56% of the skill agent's prompt tokens, since
skills load long procedural files while a search returns at most 8 KiB once.

### Search use and retrieval success

In `R1` the agent searched in 44 of 138 attempts (32%), almost always once (one
attempt searched twice); 33 of 45 searches came before the first change. The
first search was typically the fourth tool call (median index 3, quartiles
2–5): a few observations, then a search, then the repair. Queries were short
(median five words) and named the resource and symptom, as instructed, for
example `kubernetes pvc pending storageclass not found` or
`Kubernetes NetworkPolicy ingress allow from namespace`. No result reached the
8 KiB cap.

Retrieval worked when it was used: 33 of 45 searches (73%) returned a chunk
from the scenario's source page, 21 of them at rank 1. Most misses returned an
adjacent page that also covers the problem, such as
`force-delete-stateful-set-pod.md` instead of the cited `finalizers.md`, or
`assign-pods-nodes.md` instead of `assign-pod-node.md`, so the strict
file-level measure understates useful retrieval. This matches the retrieval
design study, where agent-style rewrites found a cited file for 79% of queries.

### Effect of searching

Attempts grouped by what search returned, against the same-run prompt agent on
the same scenarios:

| Group (`R1`)                   | Attempts | Scenarios | Macro |  Full | Prompt, same scenarios |
| ------------------------------ | -------: | --------: | ----: | ----: | ---------------------: |
| no search                      |       94 |        43 | 0.757 | 68/94 |         0.651 (76/129) |
| searched, source not retrieved |       11 |        10 | 0.517 |  5/11 |          0.456 (11/30) |
| source retrieved               |       33 |        20 | 0.617 | 18/33 |          0.544 (27/60) |

The agent searched on harder scenarios (the prompt agent scored 0.544 and 0.456
there, against 0.651 where RAG did not search), and attempts that retrieved the
source page beat the prompt on those scenarios by 0.073. But the largest gap is
in the attempts that **did not search** (+0.106). Retrieved text cannot explain
that part. It may come from the added prompt paragraph or the presence of the
tool, for example by making the agent identify the affected resource before
acting, or from run-to-run variation; the design cannot separate these. The same pattern appears by knowledge tag: the gain over the prompt is
small on the 26 documentation-dependent scenarios and larger on the 11
general-knowledge ones.

| Scenario tag                         | Scenarios | Prompt |   RAG | Skill | RAG − prompt, 95% CI    |
| ------------------------------------ | --------: | -----: | ----: | ----: | ----------------------- |
| `knowledge: documentation-dependent` |        26 |  0.665 | 0.693 | 0.715 | +0.029 [−0.112, +0.161] |
| `knowledge: general`                 |        11 |  0.616 | 0.778 | 0.788 | +0.162 [+0.020, +0.303] |
| `knowledge: multi-source`            |         6 |  0.591 | 0.736 | 0.624 | +0.145 [−0.189, +0.431] |
| `knowledge: version-specific`        |         3 |  0.519 | 0.778 | 0.556 | +0.259 (3 scenarios)    |
| `difficulty: easy`                   |        11 |  0.838 | 0.848 | 0.869 | +0.010 [−0.162, +0.152] |
| `difficulty: medium`                 |        22 |  0.520 | 0.658 | 0.668 | +0.138 [−0.025, +0.292] |
| `difficulty: hard`                   |        11 |  0.590 | 0.684 | 0.583 | +0.094 [−0.083, +0.261] |

### Where retrieved knowledge was not enough

The traces show three recurring failure patterns that searching did not fix:

- **Change mechanics.** Many failed repairs had the right diagnosis but could
  not apply it: JSON merge patches that replace a whole container list
  (`spec.template.spec.containers[0].image: Required value`), patches to
  immutable fields (`StatefulSet.spec.volumeClaimTemplates`, Pod containers)
  and `kubectl edit` without an editor. The agent repeated variants of the
  same failing patch instead of searching for the command, even when `R2`
  asked it to.
- **Using a retrieved page correctly.** In `service-targetport-mismatch` an
  agent retrieved the `Service` page that shows `targetPort` under `ports`,
  then patched `spec.targetPort`. In `daemonset-missing-toleration` all three
  attempts retrieved the DaemonSet page and none repaired the scenario.
- **Not searching.** Two thirds of attempts never searched, including 26 of
  the 47 attempts that did not fully succeed (`container-crash-loop`,
  `startup-probe-slow-app`, `sidecar-shared-volume`); the agent stated searching as optional in its own
  plan ("Use `search_docs` if the error is unclear").

### Required search on the full benchmark

`R2` was run on all 46 scenarios to test, with more power than the subset,
whether searching more often helps.

| Condition | Macro score | Full success | Attempts searching | Searches | Source in top 5 (searches) | Mean prompt tokens |
| --------- | ----------: | -----------: | -----------------: | -------: | -------------------------: | -----------------: |
| `R1`      |   **0.725** |   **91/138** |       44/138 (32%) |       45 |                33/45 (73%) |             44,684 |
| `R2`      |       0.711 |       89/138 |       81/138 (59%) |       98 |                56/98 (57%) |             51,760 |

`R2` nearly doubled search use and the number of attempts that retrieved their
source page (33 to 50), but scored 0.014 below `R1` (95% CI −0.094 to +0.069)
and used 16% more prompt tokens. Its additional searches were less targeted:
the share of searches returning the source page fell from 73% to 57%. Attempts
that retrieved the source page scored at the prompt agent's level on the same
scenarios (0.666 against 0.682), and the attempts that did not search again
scored highest (0.817 against 0.665). Making retrieval more frequent therefore
did not turn it into better repairs.

## Conclusions

- **RAG improves on the prompt agent about as much as curated skills do.**
  `R1` scored 0.091 above the same-run prompt agent and 0.015 above the frozen
  skill agent; both adapted conditions are about 0.08–0.09 above the prompt,
  with intervals that just include zero. RAG gets there with far less context
  than skills (44.7k against 80.1k prompt tokens per attempt).
- **Retrieval is not the bottleneck; use is.** The hybrid index returned the
  scenario's own source page for 73% of `R1` searches, usually at rank 1, from
  short queries the agent wrote itself. But the agent searched in a third of
  the attempts, almost always once, and rarely after a failed change.
- **More retrieval did not mean better repairs.** Requiring the search (`R2`)
  raised search use to 59% of attempts without improving the score, and
  attempts that retrieved the source page did not outperform the prompt agent
  on the same scenarios in `R2`. Most of `R1`'s gain over the prompt is in
  attempts that never searched, and on general-knowledge rather than
  documentation-dependent scenarios. The benefit of the RAG condition is
  therefore not clearly attributable to retrieved text.
- **Remaining failures are mostly about applying a change.** Wrong patch types,
  immutable fields and misread field paths are the common causes; the
  documentation covers them, but the model either does not search for them or
  does not apply what it read.
- **Selection.** `R1` is proposed for freezing: it is the condition specified in
  #22, has the best observed score, the lowest cost of the RAG candidates, and
  changes nothing in the code. `R2` and `R3` are kept as evidence. Lexical and
  semantic retrieval are not evaluated as separate conditions: #21 selected one
  hybrid configuration (D-035), and the results here show that retrieval
  quality is not what limits the condition.

## Limitations

- Three attempts per scenario: per-scenario differences of one attempt are
  noise, and the macro differences between `R1`, `R2`, the skill agent and the
  prompt agent all have 95% intervals that include or nearly include zero.
- The prompt → RAG difference combines the retrieved text, the added prompt
  paragraph and the presence of the tool. A control with the paragraph but no
  tool was not run, so the share of the gain due to retrieval itself is not
  identified.
- Grouping attempts by search outcome is observational: the agent decides when
  to search, and it searches on harder scenarios.
- Retrieval success is file-level and limited to the one source page per
  scenario; adjacent pages that also describe the fix count as misses.
- The skill agent comparison uses the archived `S10` run from the skills
  optimization study, not a run in the same invocation. The same-run prompt
  agent reproduced that study's prompt result exactly, which suggests the
  harness changes since then did not shift scores.
- The change-before-search ordering uses a text heuristic over Bash commands.
- Several attempts were rerun after setup failures (15 in total across the
  runs), and two power losses interrupted runs that were then resumed. `R3` ran
  under host memory pressure and across a power loss, so its lower score is not
  attributed to the presentation change.
- Candidates were screened on the 18-scenario subset; only `R1` and `R2` have
  full-set results.

## Evidence

Analysis outputs (`report.txt` as printed, `summary.json`, and one line per
attempt in `attempts.jsonl`):

| Analysis                                           | Directory                                      |
| -------------------------------------------------- | ---------------------------------------------- |
| `R1` against the same-run prompt agent (full set)  | [analysis/r1/](analysis/r1/)                   |
| `R1` against the skill agent `S10` (full set)      | [analysis/r1-vs-skill/](analysis/r1-vs-skill/) |
| `R2` against the same-run prompt agent (full set)  | [analysis/r2/](analysis/r2/)                   |
| `R1`, `R2`, `R3` against the prompt agent (subset) | [analysis/subset/](analysis/subset/)           |

Raw runs, each with `run.json`, `results.json`, per-attempt records and the
configuration used:

| Run                                     | Run ID                         | Raw results                                                        |
| --------------------------------------- | ------------------------------ | ------------------------------------------------------------------ |
| `R1` and prompt agent, full set         | `run-2026-10-02-07-35-43-038Z` | [raw/r1/](raw/r1/)                                                 |
| `R2`, full set                          | `run-2026-10-02-17-28-03-947Z` | [raw/r2/](raw/r2/)                                                 |
| `R2`, subset                            | `run-2026-10-02-10-32-47-138Z` | [raw/r2-subset/](raw/r2-subset/)                                   |
| `R3`, subset                            | `run-2026-10-02-12-21-54-157Z` | [raw/r3-subset/](raw/r3-subset/)                                   |
| skill `S10` (skills optimization study) | `run-2026-09-26-00-21-23-444Z` | [../skills-optimization/raw/s10/](../skills-optimization/raw/s10/) |

Reproduce an analysis from `benchmark/` with the run directories under
`results/` (raw directories are named by candidate, so link them back to their
run IDs first), for example:

```sh
make analyze-runs ANALYZE_ARGS='--condition run-2026-10-02-07-35-43-038Z/rag --reference run-2026-10-02-07-35-43-038Z/prompt'
```
