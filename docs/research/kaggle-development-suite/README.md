# Development suite on Kaggle

All four models under baseline, prompt, skill and RAG on the 46 development
scenarios. Raw evidence is in [raw/](raw/), analyses in
[analysis/](analysis/).

## Summary

- **736 attempts** (4 models × 4 conditions × 46 scenarios, one attempt
  each), no errors and no timeouts.
- **Baseline vs prompt.** Over the four models, prompt scores 0.698 and
  baseline 0.088 lower [95% CI −0.153, −0.025].
- **Skill and RAG vs prompt.** Skill (+0.027 [−0.032, +0.086]) and RAG
  (+0.028 [−0.034, +0.088]) are not distinguishable from prompt in one run.
  Only on Qwen3.5 9B are both above prompt beyond chance.
- **Qwen does not stop after a repair.** Over half of Qwen's turn-limit
  attempts had already fully succeeded.
- **Concurrency.** At half its slots, Qwen keeps every attempt well inside
  the 1800 s limit. Qwen3.5 9B also fits at full slots.

## Question

What do baseline, prompt, skill and RAG score on the 46 development scenarios
for each model (D-037), and how much concurrency can a run use without
attempts approaching the time limit?

## Method

- Inference ran on Kaggle (two T4 GPUs per session), served by
  [KaggleInferenceServer](https://github.com/Nasus20202/KaggleInferenceServer)
  (image `sha256:34de5847…`), with the `benchmark-kaggle.yaml` workflow
  sharding scenarios across GitHub runners. Benchmark code is `main` at
  `0adf206`.
- Settings: D-039 limits (50 turns, 100 tool calls, 60 s per tool call, 1800 s
  per attempt), vendor sampling (D-038), 32K context per slot, MTP drafts.
- Concurrency (D-025): agents in flight never exceed the slots the session
  reports.

| Model       | Slots per session | Shards × agents | Agents |
| ----------- | ----------------: | --------------- | -----: |
| Gemma 4 E2B |                64 | 16 × 3          |     48 |
| Gemma 4 E4B |                32 | 16 × 2          |     32 |
| Qwen3.5 4B  |                14 | 7 × 1           |      7 |
| Qwen3.5 9B  |                12 | 6 × 1           |      6 |

Each Qwen model ran in two halves of 23 scenarios each, to stay under GitHub's
6 h job limit, merged with `benchmark --merge`.

- Per-attempt timing comes from llama.cpp's per-response timings. _Queue_ is
  response time outside prefill and decode (waiting for a slot, plus network).
- Kaggle usage: about 12.5 h of the 30 h weekly GPU quota for the suite, and
  about 4 h for the sweep.

## Results

### Conditions

Macro score, with full successes out of 46 in brackets:

| Model       | Baseline   | Prompt     | Skill      | RAG        |
| ----------- | ---------- | ---------- | ---------- | ---------- |
| Gemma 4 E2B | 0.277 (10) | 0.400 (16) | 0.415 (17) | 0.429 (17) |
| Gemma 4 E4B | 0.531 (20) | 0.674 (26) | 0.694 (29) | 0.679 (30) |
| Qwen3.5 4B  | 0.750 (32) | 0.824 (36) | 0.840 (37) | 0.822 (37) |
| Qwen3.5 9B  | 0.882 (38) | 0.893 (38) | 0.952 (42) | 0.971 (43) |
| Mean        | 0.610      | 0.698      | 0.725      | 0.725      |

Difference from prompt on the same scenarios, with paired bootstrap 95% CIs
(`analyze --by-model`, [analysis/](analysis/)):

| Model       | Baseline                | Skill                   | RAG                     |
| ----------- | ----------------------- | ----------------------- | ----------------------- |
| Gemma 4 E2B | −0.124 [−0.245, −0.009] | +0.014 [−0.138, +0.163] | +0.029 [−0.105, +0.163] |
| Gemma 4 E4B | −0.143 [−0.279, −0.004] | +0.020 [−0.130, +0.167] | +0.005 [−0.121, +0.132] |
| Qwen3.5 4B  | −0.074 [−0.223, +0.072] | +0.016 [−0.120, +0.150] | −0.002 [−0.141, +0.134] |
| Qwen3.5 9B  | −0.011 [−0.109, +0.087] | +0.059 [+0.011, +0.119] | +0.078 [+0.007, +0.159] |
| Mean        | −0.088 [−0.153, −0.025] | +0.027 [−0.032, +0.086] | +0.028 [−0.034, +0.088] |

**RAG retrieval.** For each model: attempts that searched / attempts that
retrieved a scenario source page, then the macro on the latter against prompt
on the same scenarios.

- Qwen3.5 4B: 42 / 37, 0.932 against 0.840.
- Qwen3.5 9B: 34 / 28, 0.982 against 0.895.
- Gemma 4 E2B: 30 / 24, 0.375 against 0.406.
- Gemma 4 E4B: 11 / 6.

Retrieval helps the Qwen models. Gemma 4 E4B rarely searches, and Gemma 4 E2B
does not use what it finds.

### Attempts

| Model       | Attempt mean / p90 / max | Queue (mean) | Decode tok/s | Turn limit (of which full success) |
| ----------- | ------------------------ | -----------: | -----------: | ---------------------------------- |
| Gemma 4 E2B | 384 / 702 / 945 s        |          6 s |         10.1 | 0                                  |
| Gemma 4 E4B | 285 / 551 / 1112 s       |          5 s |         14.7 | 0                                  |
| Qwen3.5 4B  | 558 / 1023 / 1636 s      |         88 s |         19.6 | 70 (36)                            |
| Qwen3.5 9B  | 406 / 763 / 1267 s       |         32 s |         17.1 | 20 (12)                            |

- Qwen often keeps working after a full repair until it reaches 50 turns, as
  in the shared-settings study.
- A Gemma 4 E2B session is saturated at 48 agents. Its decode rate per
  request is the lowest of the four models.

### Concurrency sweep

Half of the scenarios (23, so 92 attempts) for each Qwen model were rerun with
more agents, from the same code and image.

| Model      | Agents / slots | Timeouts | Macro | Attempt mean / p90 / max | Decode tok/s | Longest shard |
| ---------- | -------------- | -------: | ----: | ------------------------ | -----------: | ------------: |
| Qwen3.5 9B | 6 / 12         |        0 | 0.952 | 354 / 600 / 995 s        |         17.5 |       140 min |
| Qwen3.5 9B | 12 / 12        |        0 | 0.925 | 514 / 955 / 1729 s       |         10.2 |       100 min |
| Qwen3.5 4B | 7 / 14         |        0 | 0.835 | 515 / 918 / 1636 s       |         19.7 |       159 min |
| Qwen3.5 4B | 10 / 14        |        1 | 0.768 | 530 / 1042 / 1800 s      |         21.2 |       124 min |

- **Qwen3.5 9B at full slots** decodes 40% slower per request, and its
  longest attempt came within 71 s of the limit. Wall-clock fell by 29%.
- **Qwen3.5 4B at 10 agents** decoded as fast as at 7. Its one timeout was an
  attempt that had already fully succeeded. It then generated 31.8K tokens in
  one response until the context filled; the same scenario took 247 s at
  7 agents. The timeout came from runaway generation, not contention.
- Score differences are within single-run noise (±0.19 between repeats).

The longest attempts, not throughput, set the concurrency, because the
1800 s limit includes queueing and slower decoding under load. Qwen3.5 9B
takes more agents than Qwen3.5 4B, despite having fewer slots, because its
attempts are shorter. Qwen3.5 4B is not GPU-bound at 10 agents, but its
longest attempts are already within 164 s of the limit. Chosen agents per
model: Gemma 4 E2B 48, Gemma 4 E4B 32, Qwen3.5 4B 7, and Qwen3.5 9B 6 (up to
12 if wall-clock matters).

## Limitations

- One attempt per scenario and condition. Per-model intervals are wide
  (±0.12–0.15), and in the shared-settings study the single-run macro moved
  by up to 0.19 between repeats.
- Qwen ran at half its slots and Gemma at most of them. Per-attempt
  wall-clock differs between models, but it affects outcomes only through the
  1800 s limit.

## Follow-ups

1. **Time limit independent of serving speed.** Count only prefill, decode
   and tool time against the limit, and lower the limit accordingly.
2. **Verify, then stop.** Add an instruction to the shared prompt, not a
   model-specific one, to verify the repair and then stop. This would shorten
   Qwen's attempts and reduce late regressions.
3. **Output cap per response.** In the recorded evidence (this suite, the
   sweep and the earlier studies, about 64K responses), every response except
   the runaway above stayed at or under 3,019 tokens.
   - Qwen: at most 1,885 tokens, p99.9 about 1,010.
   - Gemma: p99.9 about 1,900.

   A cap of 8,192 tokens is 2.7 times the largest normal response. It leaves
   about 4K tokens for the answer after a full 4,096-token reasoning budget,
   and ends a runaway as `token_limit` after about 7 minutes instead of 30.
   Grading uses the final cluster state, so scores are unchanged.

4. **Repeats.** Three attempts per scenario, condition and model narrow the
   pooled skill/RAG-against-prompt interval from about ±0.06 to about ±0.035,
   at about 12 h of quota per repeat. The next full suite runs three.
5. **Larger reference model.** Gemma 4 26B-A4B fits a Kaggle session (14
   slots at 32K) at about 4 h of quota per run. With Qwen3.5 9B already at
   0.89–0.97, it would mostly show the ceiling. It is outside D-037 and not
   planned now.

## Evidence

- [raw/](raw/): one merged benchmark run per model (`run-*` with attempt
  JSON, `results.json` and `run.json`) and the session stats each shard
  recorded at start (`server-*.json`). `raw/sweep/` holds the two sweep runs,
  with session stats at start and end.
- [analysis/attempts.csv](analysis/attempts.csv) and
  [sweep-attempts.csv](analysis/sweep-attempts.csv): one row per attempt with
  the timing split, produced by [extract.py](analysis/extract.py) and
  summarized by [summarize.py](analysis/summarize.py). Draft acceptance and
  the prompt-cache hit rate come from [spec.py](analysis/spec.py).
- `analysis/*-vs-prompt/` and the matching `.txt` files: `analyze --by-model`
  output for each comparison.
