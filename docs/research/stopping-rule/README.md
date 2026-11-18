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
95% bootstrap intervals over scenarios (D-050). A _false completion_ is an
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
