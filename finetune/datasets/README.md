# Training datasets

One directory per teacher, built by `finetune.data.dataset` from benchmark runs (D-054), and one for each synthetic variant. What they contain and what was found is in the [study](../../docs/research/training-dataset/README.md).

| Directory         | Content                                                                                                |
| ----------------- | ------------------------------------------------------------------------------------------------------ |
| `deepseek-flash/` | the whole trajectories of DeepSeek V4.1 Flash through its official API                                 |
| `qwen35-9b/`      | the whole trajectories of Qwen3.5 9B on Kaggle                                                         |
| `truncation/`     | the proposed truncation variant: cut points and the rendered cuts, per teacher                         |
| `continuation/`   | the proposed continuation variant: the Gemma attempts to continue and each teacher's continuation runs |

Both variants are derived from the teacher datasets: the truncation cuts each teacher's own trajectories, the continuation starts from one shared set of Gemma prefixes and runs with either teacher. Each variant records the dataset it was built from, by that dataset's revision, so a cut or a continuation can be traced back without copying the attempts. The ticket and the study call the two variants S1 and S2.

## Layout of a teacher directory

| Path                                      | Content                                                                                                                  |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `runs/<run>/<scenario>/prompt/NNN.json`   | every attempt of the teacher's runs on the training scenarios, kept or not, with the run's `run.json` and `results.json` |
| `manifest.yaml`                           | the kept attempts (`train`), with run, attempt, tokens and review status; scenarios without examples                     |
| `review/*.md`                             | one sheet per kept attempt: task, write commands, final report; `verdict` is `pending`, `accepted` or `rejected`         |
| `rendered/<student>.jsonl`, `freeze.json` | after the freeze: the examples rendered for each student, counts and hash                                                |

## Layout of the truncation directory

| Path                       | Content                                                                                         |
| -------------------------- | ----------------------------------------------------------------------------------------------- |
| `cuts.yaml`                | the cut point of every attempt that has a verified cut                                          |
| `freeze.json`              | the teacher dataset it was built from and its revision, the counts and what the revision covers |
| `rendered/<student>.jsonl` | the cut conversations rendered for each student; the bytes the variant freezes                  |

## Layout of the continuation directory

| Path                                              | Content                                                                                     |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `candidates.yaml`                                 | the Gemma attempts the continuation starts from, with prefix length and score               |
| `<teacher>/runs/<run>/<scenario>/prompt/NNN.json` | that teacher's continuation runs, kept or not, with the run's `run.json` and `results.json` |
| `<teacher>/manifest.yaml`                         | the kept continuations, with run, attempt and review status                                 |
| `<teacher>/review/*.md`                           | one sheet per kept continuation: task, inherited stop, the teacher's writes, its report     |
| `<teacher>/rendered/<student>.jsonl`              | the examples, with the loss masked on the inherited prefix                                  |
| `<teacher>/freeze.json`                           | the revision of manifest, candidates and renders                                            |

## Commands

Run from `finetune/`.

```
# Select up to 5 distinct clean attempts per scenario and write the manifest and sheets.
uv run python -m finetune.data.dataset --name <teacher> --runs <run dir>...

# Cut points and renders for the truncation variant; candidates for the continuation.
uv run python -m finetune.data.truncation --name <teacher>
uv run python -m finetune.data.continuations

# The continuation dataset of one teacher, from its continuation runs, and its freeze.
uv run python -m finetune.data.continuations --teacher <teacher> --runs <run dir>...
uv run python -m finetune.data.continuations --teacher <teacher> --freeze

# Render for the four students and record counts and a hash. Refuses to run while a sheet
# is pending or rejected.
uv run python -m finetune.data.freeze --name <teacher>
```

Building again keeps a sheet's verdict. A rejected attempt is dropped and the next candidate of its scenario takes its place, with a new pending sheet.

## Rules

- Only training scenarios are collected; the validation scenarios (D-055) and the held-out set never enter a manifest (checked by `tests/test_dataset.py`).
- `runs/` is raw evidence and stays exactly as the benchmark wrote it. Change the selection, not the records.
- The truncation and continuation variants are proposals; neither is part of a frozen dataset until the researcher approves it.
