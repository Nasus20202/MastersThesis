# Fine-tune Module — Agent Instructions

Python conventions for `finetune/`. See the repository root
[AGENTS.md](../AGENTS.md) for the research workflow, approval gate, and
documentation policy, which also apply here.

## Paid compute

- Never start a Modal run (`modal run`, `modal deploy`, `modal shell`) or any
  other paid service without the researcher's explicit approval for that
  specific run. State the variants, GPU type, timeout and estimated cost first.
  Approval of one run does not cover the next one, a retry or a smoke test.
- Keep `modal.timeout_minutes` in [config.yaml](config.yaml) as the cost cap;
  raise it only with approval.

## Configuration

- [config.yaml](config.yaml) is the only place for settings: GPU type,
  memory, timeouts, volume names, tool versions, image tags and digests,
  checkpoints and their revisions, training values. Code reads them through
  `finetune.config`; never hard-code a version, hash or resource size.
- Every version or image under Renovate's control has a
  `# renovate: datasource=... depName=...` comment on the line above it.
- Checkpoint revisions and training values are research inputs: change them
  only with the researcher's approval.
- Python dependencies are pinned in `pyproject.toml` and `uv.lock`. Get
  explicit approval before adding one.

## Architecture

- `src/finetune/` is the library, one subpackage per stage: `data`
  (attempts, rendering, rows, manifests), `runs` (run inputs and
  provenance), `training` (one training job, split by concern: `lora`,
  `loss`, `trainer`, `sweep`, `evaluate`, `gpu`, `run`) and
  `serving` (templates check, GGUF and perplexity). `config` reads
  `config.yaml`.
- Each subpackage keeps its dataclasses in `models.py` and its logic in the
  other modules. Keep modules short and about one thing; split a module
  that grows past a few hundred lines.
- [src/finetune/modal.py](src/finetune/modal.py) is the only Modal entrypoint:
  data is prepared locally, each container trains one variant in-process.
  Ship code to containers from the checkout (`add_local_python_source`), not
  by cloning or shelling out.
- Prefer library calls to subprocesses. The only remaining ones are `git`
  for provenance and `docker` for the pinned llama.cpp image.

## Code style

- Prefer readable, simple solutions over clever ones.
- Records are frozen dataclasses with typed fields: settings, jobs, rows,
  results, metrics. Do not pass `dict`s or `argparse.Namespace` between
  functions. A `dict` is only for a real mapping with uniform values (name ->
  version, file -> hash) or for an external format that requires it (chat
  messages, tool definitions, HTTP bodies, transformers' `device_map`).
  JSON files are written from dataclasses with `asdict`.
- Every module has a short docstring that says what it does and why.
- Add comments only when they give value (a "why", an invariant, a gotcha).
- Research inputs (validation scenarios, checkpoints, training values) live in
  `config.yaml`, never as constants in code.
- `make finetune-format` and `make finetune-lint` must pass before committing.

## Logging

- Every module that reports progress has `log = logging.getLogger(__name__)`.
  Never `print`.
- Every entrypoint (`main`, the Modal functions) calls `finetune.logs.setup()`
  once, before anything logs: one format, library progress bars off.
- A training run writes its own log to `<out>/log.txt` (`logs.LogFile`), next
  to `metrics.json`, so the log is kept with the results it explains.
- Log the start of each slow step and its outcome with the numbers that
  matter (tokens, seconds, peak memory, loss). Log recovered failures (out of
  memory) as warnings. Log an exception with `log.exception` where it is
  recorded, then let it propagate.
- Results go to files written from dataclasses (JSON), not to the log; the
  log says where they are.

## Testing

- `make finetune-test` runs on the CPU with tiny random models; no test
  downloads checkpoints or calls Modal.
- Every test must exercise some logic and fail under a small change to the
  code it covers. Do not assert a configuration value against itself.
