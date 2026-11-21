# Training datasets

One directory per teacher, built by `finetune.data.dataset` from benchmark runs (#25, D-054). What they contain and what was found is in the [study](../../docs/research/training-dataset/README.md).

| Directory         | Teacher                                      |
| ----------------- | -------------------------------------------- |
| `deepseek-flash/` | DeepSeek V4.1 Flash through its official API |
| `qwen35-9b/`      | Qwen3.5 9B on Kaggle                         |

## Layout of a teacher directory

| Path                                      | Content                                                                                                                  |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `runs/<run>/<scenario>/prompt/NNN.json`   | every attempt of the teacher's runs on the training scenarios, kept or not, with the run's `run.json` and `results.json` |
| `manifest.yaml`                           | the kept attempts (`train`), with run, attempt, tokens and review status; scenarios without examples                     |
| `review/*.md`                             | one sheet per kept attempt: task, write commands, final report; `verdict` is `pending`, `accepted` or `rejected`         |
| `rendered/<student>.jsonl`, `freeze.json` | after the freeze: the examples rendered for each student, counts and hash                                                |

## Commands

Run from `finetune/`.

```
# Select up to 5 distinct clean attempts per scenario and write the manifest and sheets.
uv run python -m finetune.data.dataset --name <teacher> --runs <run dir>...

# Render for the four students and record counts and a hash. Refuses to run while a sheet
# is pending or rejected.
uv run python -m finetune.data.freeze --name <teacher>
```

Building again keeps a sheet's verdict. A rejected attempt is dropped and the next candidate of its scenario takes its place, with a new pending sheet.

## Rules

- Only training scenarios are collected; the validation scenarios (D-055) and the held-out set never enter a manifest (checked by `tests/test_dataset.py`).
- `runs/` is raw evidence and stays exactly as the benchmark wrote it. Change the selection, not the records.
