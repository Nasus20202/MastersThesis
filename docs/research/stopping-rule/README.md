# Stopping rule for the prompt agent

Every condition fails in two ways at the end of an attempt: Gemma stops before
the repair is verified, and Qwen keeps checking after it is done. This study
tests whether one added paragraph in the system prompt fixes either failure
(#123). Raw evidence is in [raw/](raw/), analyses in
[analysis/](analysis/).

## Summary

**The stopping rule did not help.** Neither failure changed beyond noise, and
the score did not change.

- **Not stopping.** Qwen3.5 4B ends 25 of 54 attempts at the turn limit with or
  without the rule, after the same 41 turns on average.
- **Stopping too early.** Gemma reports an unverified repair in 65 of 108
  attempts without the rule and in 62 with either version, including the one
  that forbids exactly this.
- **Score.** P5 gains +0.010 macro [−0.050, +0.071] over the four models;
  random sign flips give a gain at least this large 76% of the time.

**The prompts stay unchanged.** P5 formally passes the acceptance rule fixed
before the run, but a prompt with no effect passes that rule about one
time in six, and P5 did not change the behaviour it was written for. The
researcher therefore kept the current prompt, skill and RAG prompts, overriding
the rule's outcome; the round is reported as negative.

## Question

Does a stopping rule in the system prompt reduce the two stopping failures
without lowering the score?

- **Stopping too early.** Gemma ends the attempt while the workload is visibly
  unready: 20 to 23 of 138 attempts per condition in the archived full runs
  (#126).
- **Not stopping.** Qwen keeps re-checking after the repair is done: in the
  [shared-settings study](../shared-settings/README.md) Qwen3.5 4B ended 26 of
  54 attempts at the turn limit, 11 of them already repaired.

## Method

Two candidate paragraphs were appended to the prompt agent's system prompt:

- **P5**, for Qwen: finish as soon as a fresh observation after the last
  change shows that the outcome and the constraints hold, and do not repeat a
  check that has passed ([p5.md](candidates/p5.md)).
- **P6**, P5 plus a sentence for Gemma: do not give the final report while a
  check is failing, has timed out or has not been made
  ([p6.md](candidates/p6.md)).

Each candidate and the unchanged prompt (_reference_) ran the 18-scenario
[subset](../skills-optimization/subset.txt) three times on Gemma 4 E2B, Gemma 4
E4B, Qwen3.5 4B and Qwen3.5 9B: 648 attempts on Kaggle under the shared
settings (D-038 to D-041). Scores are macro scores with paired differences and
95% bootstrap intervals over scenarios (D-049). A _false completion_ is an
attempt the agent ended itself without full success.

Fixed before any result was read:

- **Acceptance.** A candidate replaces the prompt only if the mean
  macro over models is higher, no model's interval lies entirely below zero,
  and both false completions and turn-limit endings decrease. If both pass,
  the higher mean macro wins.
- **Carry-over.** An accepted paragraph also goes into the skill and RAG
  prompts. Both agents then run once on the subset and are compared with the
  same agents in the [development suite](../kaggle-development-suite/README.md);
  only an interval entirely below zero would stop the freeze.

## Results

### Did the rule change the stopping failures?

| Ending (attempts)                 | Reference |   P5 |   P6 |
| --------------------------------- | --------: | ---: | ---: |
| Gemma: false completion (of 108)  |        65 |   62 |   62 |
| Qwen3.5 4B: turn limit (of 54)    |        25 |   25 |   24 |
| … of which already fully repaired |        13 |    7 |    7 |
| Qwen3.5 4B: mean turns            |      40.9 | 40.6 | 40.9 |
| Qwen3.5 9B: turn limit (of 54)    |         8 |    6 |    9 |
| Qwen3.5 9B: mean turns            |      30.3 | 27.6 | 29.1 |
| All models: false completion      |        75 |   70 |   68 |
| All models: turn limit            |        33 |   31 |   33 |

- **Not stopping: no.** Qwen3.5 4B runs to the limit as often and for as long
  as before. Fewer of its limit endings were already repaired (13 → 7) only
  because it fully repaired fewer scenarios (39 → 32 of 54), not because it
  stopped earlier. Qwen3.5 9B ends two fewer attempts at the limit with P5 and
  one more with P6.
- **Stopping too early: no.** Gemma's false completions fall by 3 of 108 with
  either version. P6, which forbids reporting an unverified outcome, does no
  better than P5: an instruction alone does not change Gemma's early stops.

### Did the score change?

Macro score, and difference from the reference with 95% interval:

| Model       | Reference |        P5 | P5 − reference              |        P6 | P6 − reference              |
| ----------- | --------: | --------: | --------------------------- | --------: | --------------------------- |
| Gemma 4 E2B |     0.311 |     0.326 | +0.015 [−0.099, +0.136]     |     0.330 | +0.019 [−0.130, +0.167]     |
| Gemma 4 E4B |     0.627 |     0.662 | +0.034 [−0.100, +0.172]     |     0.653 | +0.025 [−0.077, +0.142]     |
| Qwen3.5 4B  |     0.779 |     0.705 | −0.074 [−0.210, +0.071]     |     0.720 | −0.059 [−0.190, +0.051]     |
| Qwen3.5 9B  |     0.881 |     0.946 | +0.066 [+0.002, +0.140]     |     0.893 | +0.012 [−0.066, +0.093]     |
| **Mean**    | **0.650** | **0.660** | **+0.010 [−0.050, +0.071]** | **0.649** | **−0.001 [−0.060, +0.060]** |

No. Only Qwen3.5 9B with P5 has an interval above zero, at a lower bound of
+0.002; Qwen3.5 4B, the model P5 was written for, moves the other way.

### Acceptance checks

The four checks fixed before the run (see [Method](#method)):

| Check                                      | P5             | P6             |
| ------------------------------------------ | -------------- | -------------- |
| 1. Mean macro higher than the reference    | yes (+0.010)   | no (−0.001)    |
| 2. No model's interval entirely below zero | yes            | yes            |
| 3. Fewer false completions                 | yes (70 vs 75) | yes (68 vs 75) |
| 4. Fewer turn-limit endings                | yes (31 vs 33) | no (33 vs 33)  |

P5 passes all four and P6 does not, so under the rule P5 would replace the
prompt. Two checks made after the results were read show that P5's pass is
within noise:

- a random sign on each scenario's difference gives a mean gain at least as
  large as +0.010 in 76% of 20,000 flips;
- pooling the reference and P5 attempts per model and scenario and splitting
  them at random into halves of three, 16% of 5,000 splits pass checks 1, 3
  and 4. Check 2 rarely fails without an effect, so a prompt with no effect
  passes all four about one time in six.

### Decision

The prompts are kept unchanged. Every difference that let P5 pass is within
noise, and the paragraph changed neither failure it targets; adding it to
every condition's prompt would add text with no shown effect. Keeping the
prompts overrides the acceptance rule.

### Carry-over into skill and RAG

The carry-over check ran after P5 had passed the rule and before the decision
to keep the prompts. Difference of the skill and RAG agents with the P5
paragraph from the same agents in the development suite, one attempt per
scenario:

| Model       | Skill + P5                  | RAG + P5                    |
| ----------- | --------------------------- | --------------------------- |
| Gemma 4 E2B | −0.083 [−0.333, +0.167]     | −0.014 [−0.250, +0.222]     |
| Gemma 4 E4B | +0.134 [−0.069, +0.343]     | +0.186 [−0.069, +0.441]     |
| Qwen3.5 4B  | +0.019 [−0.183, +0.233]     | −0.061 [−0.311, +0.180]     |
| Qwen3.5 9B  | +0.028 [+0.000, +0.083]     | +0.000 [+0.000, +0.000]     |
| **Mean**    | **+0.025 [−0.090, +0.111]** | **+0.028 [−0.102, +0.141]** |

No interval lies below zero: the paragraph does not harm the skill and RAG
agents either. One attempt rules out only a large regression.

**Concurrency matters for Qwen3.5 9B.** Its first carry-over run had 12 agents
in flight instead of the suite's 6. With the longer skill and RAG contexts the
median time per turn doubled (40 s and 28 s against 17 s and 15 s), 6 of 36
attempts hit the 1800 s limit after only 18 to 44 turns, and the RAG
difference fell to −0.125 [−0.292, +0.000]. At 6 agents no attempt timed out;
the table uses that run. Skill and RAG runs of Qwen3.5 9B are only comparable
at the same number of agents.

## Why the failures persist

A post-hoc reading of the same 648 attempts, with no new runs.

### Qwen: repaired, but no passing check

All 27 Qwen3.5 4B turn-limit endings with full success were read by hand. Two
turns were recorded for each:

- the last change the final state depends on;
- the first turn the agent's own output showed the outcome.

The repair lands a median of 16 turns before the limit (range 0–39), and the
agent first sees the outcome a median of 11 turns before it.

| After the repair, the agent…                                           | ref |  P5 |  P6 | All |
| ---------------------------------------------------------------------- | --: | --: | --: | --: |
| tries and fails to build a request through the Service or Gateway      |   4 |   2 |   5 |  11 |
| fixed only in the last 7 turns (a late repair, not a stopping failure) |   5 |   2 |   1 |   8 |
| re-reads objects that already show the outcome                         |   1 |   3 |   1 |   5 |
| waits for a slow controller (ReplicaSet or init-container backoff)     |   3 |   0 |   0 |   3 |

Most of the time after the outcome is visible goes into the request check:
232 of 357 tool calls, 24% of which fail. The failing requests have these causes:

- `kubectl run` with wrong flags;
- curl images given wget arguments;
- images without curl or wget;
- requests to Traefik without the `Host` header;
- `app.example.com` resolved from the sandbox.

The prompt asks for exactly this check ("for a Service, confirm a request
through it succeeds") and to keep diagnosing while any check fails. The agent
never sees a passing check, so P5's "finish when a fresh observation shows the
outcome" never applies.

On `gateway-route-misconfig`:

| Model       | Attempts at the turn limit | Of which repaired |
| ----------- | -------------------------: | ----------------: |
| Qwen3.5 4B  |                     7 of 9 |                 6 |
| Qwen3.5 9B  |                     7 of 9 |                 7 |
| Gemma 4 E4B |                     0 of 9 |                 — |

Gemma 4 E4B ends all 9 attempts with full success and never sends a request.

Qwen3.5 9B sometimes does get a correct answer through the Gateway, but doubts
it ("That might be default") and keeps testing other hosts and paths.

The 47 Qwen3.5 4B turn-limit endings without full success are still repairing:
30 changed the cluster in their last 10 turns.

### Gemma: change not applied, failure misread, or wrong check

The 189 Gemma false completions were classified by a heuristic on the agent's
own trace and final report. Agreement with a manual reading:

| Sample                                            | Categories | Reports |
| ------------------------------------------------- | ---------: | ------: |
| 40 attempts used to tune the heuristic            |      36/40 |   38/40 |
| 20 further attempts, before the last tuning round |      15/20 |   20/20 |
| 20 further attempts, after the last tuning round  |      18/20 |   20/20 |

| After the last accepted change  | Claims success | "Should" work | Admits failure | Total |
| ------------------------------- | -------------: | ------------: | -------------: | ----: |
| No change reached the cluster   |             21 |             7 |             11 |    39 |
| No observation after the change |              7 |             1 |              5 |    13 |
| An observation showed failure   |             35 |             4 |             42 |    81 |
| Rollout still in progress       |              2 |             0 |              1 |     3 |
| Observations looked healthy     |             46 |             1 |              6 |    53 |
| **Total**                       |        **111** |        **13** |         **65** |   189 |

- **Unverified reports are rare.** Only 13 attempts made a change and never
  looked again.
- **Changes that never reached the cluster are common.** In 39 attempts no
  change reached the cluster, for one or both of two reasons:
  - a local manifest was edited and never applied;
  - the API server rejected the change (a `Conflict` from a stale
    `resourceVersion`, an invalid patch, or `kubectl edit` without an editor).

  21 of the 39 still report the change as made, some with invented output.

- **The largest group misreads a failing observation.** Examples:
  - READY `3/3` from the old ReplicaSet while one replica is updated;
  - an `nslookup` that times out, followed by "DNS restored".
- **The second group checks a proxy or makes a different repair:**
  - compares the image tag instead of the pulled build;
  - checks that a Pod is Running instead of sending a DNS or dependency request;
  - patches the failing command instead of rolling back;
  - removes the read-only root filesystem.
- **Transient and destructive changes.** 40 attempts made only transient
  changes (Pod deletes, `rollout restart`); 13 deleted the Deployment or
  StatefulSet that owns the fault.
- **A third are honest give-ups.** 65 admit that the outcome was not reached.
  30 of these name a limitation that does not exist, such as "without permission
  to modify the deployment".

In 181 of the 189 at least one outcome criterion fails; the most frequent are:

| Criterion          | Failures |
| ------------------ | -------: |
| `deployment-ready` |       63 |
| `replica-capacity` |       37 |
| `service-response` |       21 |

Most early stops are scenarios Gemma does not repair; it stops anyway.

Reports by condition, over both Gemma models:

| Condition | Claim success | Admit failure |
| --------- | ------------: | ------------: |
| ref       |            42 |            20 |
| P5        |            36 |            18 |
| P6        |            33 |            27 |

The number of early stops does not change across conditions. P6 may turn some
success claims into admissions; this is a post-hoc reading of small counts.

### What this suggests

A further stopping paragraph addresses neither failure:

- **Qwen** lacks the passing observation that a stopping rule waits for.
- **Gemma** mostly believes a change or check that its own output contradicts.

Both point to the Harness mechanisms (#35), which can work from the agent's own
trace without grader information:

- **Loop notice:** fires after N turns without a cluster change, or K failed
  commands in a row.
- **Check before finishing:** shows which changes the API server accepted,
  which local files were never applied, and the latest state of the changed
  objects.

The request-check sentence of the prompt is a separate candidate, tested in the
[request-check fallback pilot](#request-check-fallback-pilot) below. It did not
help.

## Request-check fallback pilot

Exploratory and post-hoc, run after the decision above. No prompt was changed
and no decision is recorded.

### Question

Qwen3.5 4B often reaches the turn limit after repairing a Service or Gateway
because it cannot build the request the prompt asks for ("confirm a request
through it succeeds"). Does allowing a fallback to resource status, after two
failed request attempts, make it stop on these scenarios without lowering the
score?

### Method

**Candidate.** [P7](candidates/p7.md) is the prompt agent's system prompt with
one clause extended. The new text is in bold:

> for a Service, confirm a request through it succeeds, **and if two attempts
> to send such a request fail because of the request command itself rather than
> the Service, use instead the strongest status the resources report (ready
> endpoints, accepted routes) and state which check you could not make**;

**Runs.** Qwen3.5 4B, prompt agent, on the three development scenarios whose
check is a request: `gateway-route-misconfig`, `service-targetport-mismatch` and
`readiness-probe-breaks-endpoints`. Both arms ran locally, one after the other,
on one host (AMD RX 5700, Vulkan; llama.cpp on port 8090; 2 agents in flight),
at revision `e6c88f6b` with the shared settings (D-038 to D-041) and P7 through
[a config overlay](candidates/p7.yaml). Two batches per arm, 3 then 2 attempts
per scenario: 15 attempts per arm, P7 first within each batch. The raw runs are
in [raw/](raw/), and `make analyze-runs` compares P7 with the same-host reference
([report](analysis/p7-vs-reference.txt)). The Kaggle attempts above on the same
three scenarios are context only, because host and concurrency differ.

### Results

**The fallback did not help.** P7 ended more attempts at the turn limit than the
same-host reference, and the agent never used the fallback.

| Arm (15 attempts each) | Turn limit | Full success | Macro | Mean turns | Fallback stated in report |
| ---------------------- | ---------: | -----------: | ----: | ---------: | ------------------------: |
| P7, local              |          7 |           14 | 0.967 |       41.4 |                         0 |
| Reference, local       |          3 |           15 | 1.000 |       41.6 |                         — |
| Kaggle reference (9)   |          5 |            8 | 0.889 |       46.0 |                         — |
| Kaggle P5 and P6 (18)  |          6 |           17 | 0.963 |       43.8 |                         — |

Turn-limit endings per scenario, out of 5 attempts: `gateway-route-misconfig` 4
for P7 against 3 for the reference, `service-targetport-mismatch` 2 against 0,
`readiness-probe-breaks-endpoints` 1 against 0.

- **Turn-limit endings: 7 against 3.** Within chance (Fisher's exact test,
  two-sided p = 0.25), but a large reduction would have shown, and none did.
- **Score.** The macro difference is −0.033 [95% CI −0.100, +0.000]. The one
  lost criterion is `target-port-correct` in one `service-targetport-mismatch`
  attempt, which ended at the limit with the port still wrong.
- **The fallback was never used.** No P7 report states which check could not be
  made. In the Gateway attempts at the limit the agent stops trying requests
  after a few attempts but does not finish: it keeps looking for objects that do
  not exist (`xdsroute`, `lrroute`, "gatewayroutes") and re-reads a route status
  it has already seen as Accepted. Its reasoning still asks to "verify that
  traffic actually flows through".
- **Same-host reference.** It ended at the limit less often than the Kaggle
  reference on these scenarios (3 of 15 against 5 of 9), so the stopping failure
  partly depends on the run.

**Reading.** Permission to fall back is not enough for Qwen3.5 4B to stop: once
the request has failed, it does not judge that the request command was at fault
and keeps exploring. This supports leaving the stopping failure to the Harness
(#35), for example a loop notice after turns without a change, rather than
further prompt text.

### Pilot limitations

- Three scenarios and few attempts per arm detect only a large change in
  turn-limit endings, and say nothing about other scenarios or models.
- The fallback sentence relaxes a verification rule. On Gemma it could raise
  false completions; this was not tested.
- Local and Kaggle runs differ in host and concurrency.

## Limitations

- Three attempts on 18 scenarios detect only a clear effect: single runs of
  one agent on this subset differed by up to 0.19 macro in earlier studies.
- The chance checks were defined after the results were read, and the
  decision to keep the prompts overrides the rule fixed in advance. Both
  favour the null result; the rule's outcome is reported alongside.
- The carry-over check has one attempt per scenario and compares with runs
  made a month earlier at another revision.
- "Fewer than the reference" passes on a difference of one attempt, so a
  prompt with no effect passes the rule about one time in six. A future round
  should require the failure reductions, like the score, to exceed noise.
- The acceptance rule counts the failures over all models together, so Gemma dominates the
  false completions and Qwen3.5 4B the turn-limit endings.
- The prompt, skill and RAG conditions received unequal tuning effort before
  this round.
- The failure diagnosis in [Why the failures persist](#why-the-failures-persist)
  was made after the decision, by one coder who knew the conditions. The Gemma
  categories come from a heuristic whose agreement with a manual reading is
  given above; treat its counts as ±10%.
