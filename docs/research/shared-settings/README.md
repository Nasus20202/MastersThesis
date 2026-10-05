# Shared settings for the joint optimization

## Summary

This study is stage 1 of the joint optimization (#123). Before the prompt,
skill and RAG conditions are tuned again on four models (D-037), it fixes the
settings they all share and measures a new reference for stage 2.

- **Reference.** With vendor sampling and the 50-turn limits, the prompt agent
  scored 0.658 on average over the four models. Qwen3.5 9B scored highest
  (0.929) and Gemma 4 E2B lowest (0.373).
- **File tools.** Adding `read_file`, `write_file` and `edit_file`, together
  with a hint when `kubectl edit` is called, raised the average by 0.050. The
  95% interval (−0.071 to +0.148) includes zero, but fewer `kubectl` writes
  failed (22% against 30%) and no model got worse beyond chance. The tools now
  join every condition (D-041).
- **Context.** No attempt ran out of context. The largest context used was
  30.9K of the 32K tokens per slot, so 32K stays (D-040).
- **Limits.** Qwen often uses all 50 turns, but in most of those attempts the
  repair was already done. Gemma never reaches a limit; it stops too early
  instead. The D-039 limits stay.

## Question

1. What does the prompt agent score on each model under the new shared
   settings? These are vendor sampling (D-038), 50 turns, 100 tool calls and
   1800 s per attempt (D-039), with peak context recorded (D-040). The result
   is the reference for stage 2.
2. Do the sandbox editing changes from #145 (D-041) help, and should the file
   tools join every condition?
3. Is 32K context per slot enough, and are the new limits reached?

## Method

**Setup.** The frozen prompt agent ran on the 18-scenario
[subset](../skills-optimization/subset.txt) of the development set. The four
[model profiles](../../../benchmark/model-profiles/) are Gemma 4 E2B and E4B
(QAT `UD-Q4_K_XL`) and Qwen3.5 4B and 9B (`Q4_K_M`). They were served by the
pinned llama.cpp router with multi-token prediction and 32K context per slot,
using each vendor's sampling from
[`models-preset.ini`](../../../benchmark/models-preset.ini). Every `run.json`
records the sampling values the server applied; within a model they are the
same in all runs.

**Reference.** Revision `482aca2` with overlay
[`config/reference.yaml`](config/reference.yaml) (the D-039 limits) and a
sandbox image without the `KUBE_EDITOR` hint. Each scenario ran three times
per model (54 attempts).

**Variant.** Revision `8a73cf8` with overlay
[`config/files.yaml`](config/files.yaml): the same limits plus
`file_tools: true`. In its sandbox image `kubectl edit` fails at once with a
hint to export the manifest, change it and apply it. Each scenario ran once
per model (18 attempts), as in the screening protocol agreed for #123: a
version is chosen by the mean over models, and each model is reported
separately. The hint and the file tools were changed together, so their
effects are not separated.

**GPU crashes.** The GPU lost the device six times during the runs, ending
the Qwen attempts running at those moments: 4 in the Qwen3.5 4B reference and
4 in each Qwen variant. They were rerun on the same revision, image and
overlay; one Qwen3.5 9B attempt crashed again on two reruns before it
completed. The 14 crashed records are not kept. No Gemma attempt was
affected.

**Measures.** All numbers come from `make analyze-runs --by-model` and the
attempt records:

- _macro score_: the mean score per scenario, then the mean over scenarios;
- _full success_: attempts that passed every check;
- _Δ macro_ with a 95% paired bootstrap interval over scenarios;
- attempts ended by a limit, and _peak context_ (the largest prompt plus
  completion of one response);
- file tool calls and _failed writes_: bash calls running a writing `kubectl`
  command (`apply`, `patch`, `edit`, `set`, `delete`, `scale`, …) that exited
  with an error.

## Results

### Reference

| Model       | Macro | Full success | Range of single runs | Turn limits | Mean turns | Mean duration | Failed writes | Peak context (median / max) |
| ----------- | ----: | -----------: | -------------------: | ----------: | ---------: | ------------: | ------------: | --------------------------: |
| Gemma 4 E2B | 0.373 |        18/54 |        0.324 – 0.463 |           0 |        9.0 |           99s |  42/111 (38%) |                6.0K / 11.8K |
| Gemma 4 E4B | 0.561 |        25/54 |        0.546 – 0.583 |           0 |       10.3 |          129s |  67/145 (46%) |                5.6K / 12.3K |
| Qwen3.5 4B  | 0.769 |        37/54 |        0.676 – 0.866 |          26 |       42.9 |          375s |  78/276 (28%) |               16.5K / 27.2K |
| Qwen3.5 9B  | 0.929 |        45/54 |        0.850 – 0.975 |          10 |       30.0 |          347s |  41/224 (18%) |               10.6K / 27.2K |
| Mean        | 0.658 |              |                      |             |            |               |               |                             |

The range of single runs is the macro score of each of the three repeats on its
own. It shows how much a one-run screening result can move by chance: up to
0.19 for Qwen3.5 4B and 0.14 for Gemma 4 E2B.

The two model families fail in opposite ways:

- **Gemma stops too early.** In 36 of 54 Gemma 4 E2B and 29 of 54 Gemma 4 E4B
  attempts, the agent ended on its own and reported a repair that the grader
  did not confirm. 38–46% of its `kubectl` writes failed.
- **Qwen does not stop.** Qwen3.5 4B used all 50 turns in 26 attempts, and 11
  of them had already fully succeeded (Qwen3.5 9B: 5 of 10). In the attempts
  inspected, the last turns only re-checked a workload that was already
  repaired. Attempts that hit the turn limit scored 0.560 (4B) and 0.760 (9B),
  against 0.964 and 0.967 for attempts the agent finished itself. A higher
  limit would mostly add time; teaching the agent when to stop is a prompt
  question for stage 2.

For context only: under the earlier settings (llama.cpp default sampling,
25 turns, 600 s), Gemma 4 E4B scored 0.578 and 0.595 on the same subset, and
Qwen3.5 9B scored 0.852 on one run with 9 of 18 attempts at a limit
([RAG development evaluation](../rag-development-evaluation/README.md)).
Sampling, limits and revision all differ, so the change cannot be attributed
to one setting.

### File tools and the `kubectl edit` hint

One run per scenario, compared with the reference on the same scenarios
([analysis](analysis/files-vs-reference.txt)).

| Model       | Macro | Full success | Δ macro (95% CI)        | Turn limits | Attempts using file tools | File tool calls (errors) | Failed writes |
| ----------- | ----: | -----------: | ----------------------- | ----------: | ------------------------: | -----------------------: | ------------: |
| Gemma 4 E2B | 0.435 |         7/18 | +0.062 (−0.105, +0.235) |           0 |                     13/18 |                  25 (11) |    8/35 (23%) |
| Gemma 4 E4B | 0.685 |        12/18 | +0.124 (−0.015, +0.269) |           0 |                      5/18 |                    7 (1) |   12/41 (29%) |
| Qwen3.5 4B  | 0.750 |        13/18 | −0.019 (−0.244, +0.181) |           6 |                      7/18 |                   15 (1) |   15/50 (30%) |
| Qwen3.5 9B  | 0.961 |        16/18 | +0.032 (−0.056, +0.111) |           3 |                      3/18 |                    3 (2) |    7/61 (11%) |
| Mean        | 0.708 |              | +0.050 (−0.071, +0.148) |             |                           |                          |               |

- **Score.** Three models improved and Qwen3.5 4B stayed level. Every
  per-model interval includes zero. Only Gemma 4 E4B scored outside the range
  of its single reference runs.
- **Failed writes.** Over all models, 42 of 187 writes failed (22%), against
  228 of 756 (30%) in the reference. The share fell for three models and stayed
  level for Qwen3.5 4B.
- **How the tools were used.** Mostly to write a manifest and apply it: 14 of
  23 `write_file` calls were followed by `kubectl apply` or `kubectl replace`.
  All 15 file tool errors were reads or edits of a missing file or of text that
  was not in the file; the agent then usually wrote the file or used `kubectl`.
- **`kubectl edit`.** Rare in both runs (1–7 calls per run). In the reference
  it always failed with "unable to launch the editor". With the hint, no
  attempt called it twice; the next call exported the manifest or patched the
  object.
- Attempts that used the file tools did not score consistently higher than
  those that did not. The agent picks the tools for harder repairs, so this
  comparison says nothing about cause.

### Context

No attempt ran out of context. Gemma never used more than 16.3K tokens. Qwen's
context grows with the number of turns: 5 of 108 Qwen reference attempts and
2 of 36 variant attempts went above 24K, with a maximum of 30.9K (a Qwen3.5 4B
variant attempt). 32K per slot is enough at 50 turns; a higher turn limit would
need more context for Qwen.

## Decisions

- **D-040.** Context stays at 32K per slot.
- **D-041.** The file tools and the `kubectl edit` hint join every condition.
- **D-039** limits stay unchanged.

## Limitations

- The variant ran once per scenario. Single reference runs vary by up to
  0.19 macro, so per-model differences of this size cannot be told apart from
  noise; only the direction across models is informative.
- The hint and the file tools were changed together. `kubectl edit` was rare,
  so most of any effect is likely from the file tools, but this was not tested
  separately.
- The two sandbox images were built at different times. Apart from
  `KUBE_EDITOR`, 16 Ubuntu packages differ by security patch version
  ([diff](config/sandbox-packages.diff)); `kubectl` 1.37.0, `jq` 1.8.1 and the
  package set are the same.
- Failed writes are counted from the bash command text and its exit code. The
  count misses writes inside scripts and includes dry runs.
- Attempts rerun after a GPU crash ran later, with fewer attempts in parallel
  than the original batch.
- The subset covers 18 of the 46 development scenarios; final comparisons use
  the full set.

## Evidence

| Run                    | Run ID                         | Raw evidence                                   |
| ---------------------- | ------------------------------ | ---------------------------------------------- |
| Reference, Gemma 4 E2B | `run-2026-10-04-15-17-07-022Z` | [raw/reference-e2b/](raw/reference-e2b/)       |
| Reference, Gemma 4 E4B | `run-2026-10-04-16-19-52-931Z` | [raw/reference-e4b/](raw/reference-e4b/)       |
| Reference, Qwen3.5 4B  | `run-2026-10-04-17-25-51-416Z` | [raw/reference-qwen4b/](raw/reference-qwen4b/) |
| Reference, Qwen3.5 9B  | `run-2026-10-04-20-17-48-121Z` | [raw/reference-qwen9b/](raw/reference-qwen9b/) |
| Variant, Gemma 4 E2B   | `run-2026-10-04-23-14-40-502Z` | [raw/files-e2b/](raw/files-e2b/)               |
| Variant, Gemma 4 E4B   | `run-2026-10-04-23-38-28-373Z` | [raw/files-e4b/](raw/files-e4b/)               |
| Variant, Qwen3.5 4B    | `run-2026-10-04-23-57-34-691Z` | [raw/files-qwen4b/](raw/files-qwen4b/)         |
| Variant, Qwen3.5 9B    | `run-2026-10-05-00-44-33-119Z` | [raw/files-qwen9b/](raw/files-qwen9b/)         |

Analyses: [reference by model](analysis/reference.txt) and
[variant against reference](analysis/files-vs-reference.txt), each with
`attempts.jsonl`, `summary.json` and `model_mean.json`. Sandbox images:
reference `sha256:bd3af487fa6d…`, variant `sha256:9f1479b21e10…`.
