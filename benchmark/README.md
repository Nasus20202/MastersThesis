# Benchmark

Go runner for the thesis benchmark. It runs reproducible Kubernetes tasks in
disposable local `kind` clusters and verifies the resulting state
deterministically.

## Scenarios

Each scenario lives in its own directory, and its definition is named
`scenario.yaml` (or `scenario.yml`):

```text
scenarios/<area>/<scenario-id>/
  scenario.yaml
  validation.yaml
  manifests/
  scripts/
  kind/
```

```yaml
id: example-incident
title: Example technical incident
tags:
  difficulty: medium
  task_type: troubleshooting
  knowledge: documentation-dependent
  area: workloads
sources:
  - path: content/en/docs/concepts/workloads/controllers/deployment.md
    sections:
      - Rolling Back a Deployment
task: Restore the application to its healthy state.

cluster:
  kind:
    config: kind/cluster.yaml

prepare:
  - program: kubectl
    args: [apply, -f, manifests/app.yaml]
    env:
      EXAMPLE_MODE: strict

verify_clean:
  - program: kubectl
    args: [rollout, status, deployment/app, --timeout=60s]

inject_fault:
  - program: ./scripts/inject-fault.sh

verify_fault:
  - program: ./scripts/verify-fault.sh

grading:
  - id: workload-ready
    weight: 1
    check:
      program: ./scripts/check-restored.sh
```

The runner prepares the environment, verifies the initial state, optionally
injects and verifies a fault, runs the model (or applies a declared validation
repair), grades the result and deletes the cluster. `inject_fault` and
`verify_fault` are independently optional, so a scenario can be a
troubleshooting task or a constructive one such as deploying resources.
Command, manifest and Kind config paths resolve relative to the scenario file.

`sources` lists the evaluator-only documentation pages a scenario is grounded
in: each `path` is relative to the corpus repository root and `sections` are
heading paths joined with `>`. They are never shown to the model;
`--check-corpus` requires at least one, and `analyze` uses them to score RAG
retrieval.

### Cluster profiles and registry cache

`cluster.kind.config` selects a profile from `benchmark/kind/`; without one,
Kind uses its default single-node cluster:

- `cluster.yaml` — default kindnet, most scenarios;
- `cluster-calico.yaml` — Calico instead of the default CNI, for NetworkPolicy scenarios;
- `cluster-registry.yaml` — default plus a mirror for the in-cluster registry used by the image scenarios.

All profiles pull `docker.io`, `quay.io` and `ghcr.io` images through shared
pull-through caches that persist across clusters. `make benchmark` and
`make benchmark-validate` start them; `make registry-start` and
`make registry-stop` manage them directly. Setup, fault handling and grading
run in a pinned `benchmark-setup` container on the kind network, separate from
the hardened model sandbox.

### Tags and filtering

Scenarios may declare free-form, evaluator-only `tags`; they are never shown to
the model. The development corpus uses `difficulty` (easy, medium, hard,
anchor), `task_type` (troubleshooting, constructive), `knowledge` (general,
documentation-dependent, multi-source, version-specific) and `area`.

`--tag KEY=VALUE` filters a scenario or validation run and may be repeated or
comma-separated. Different keys combine with AND, values of one key with OR:

```sh
go run ./cmd/benchmark --config config.yaml --scenario scenarios/ \
  --tag difficulty=medium --tag difficulty=hard --tag area=networking
```

An empty selection is an error, and the filter is recorded in the run metadata
as `tag_selector`. With the Makefile:
`make benchmark TAG='difficulty=hard area=networking'`.

## Validation and results

A validation file (`validation.yaml`) references a scenario and declares
deterministic repair cases with their expected score and full success:

```yaml
scenarios:
  - scenario_file: scenario.yaml
    cases:
      - id: broken
        expected_score: 0
        expected_full_success: false
      - id: repaired
        repair:
          - program: kubectl
            args: [set, image, deployment/app, app=nginx:1.31.5]
        expected_score: 1
        expected_full_success: true
```

`--scenario` and `--validate` accept files or directories; directories are
scanned recursively for the matching file name and other YAML is ignored.
`--check-corpus PATH` (`make check-corpus`) loads every scenario, requires its
ID to match its directory and requires a paired validation file with at least
one case.

`--parallel N` bounds concurrency, `--repeat N` repeats each scenario or
validation case, and `--agent baseline,prompt` selects conditions (default
`all`). Each invocation writes one run to `results/<run-id>/`: metadata in
`run.json`, the summary in `results.json` and each attempt in
`<scenario-id>/<agent-name>/<attempt>.json`. Grading criteria keep stdout,
stderr, exit status and duration; a failed lifecycle command records its
phase, command and output in the attempt's `failure` object.

## Results browser

`cmd/browser` is a read-only terminal UI over `results/`. It loads the run
history and the scenario corpus and follows new results as they are written:

```sh
make browser
# or: cd benchmark && go run ./cmd/browser --results results --scenarios scenarios
```

`--results` and `--scenarios` default to `results` and `scenarios` (Makefile:
`BROWSER_RESULTS`, `BROWSER_SCENARIOS`); `--no-watch` disables live updates.
Four tabs (**Runs**, **Agents**, **Tasks**, **Totals**) aggregate the whole
history, and at 112 columns or more each list sits beside a dashboard. `/`
filters by model and scenario tags.

## Run analysis

`cmd/analyze` scores finished runs evaluator-side. Per condition it reports
the macro score, full successes, mean turns and tokens, failed `kubectl`
changes, use of `kubectl edit` and `kubectl rollout status`, and per-criterion
pass rates; for the RAG condition, also search use and whether each search
returned one of the scenario's source pages. `--reference` compares on the same
scenarios with a paired bootstrap interval for the macro difference.

```sh
make analyze-runs ANALYZE_ARGS='--condition RUN/rag --reference RUN/prompt --out DIR'
```

`--subset FILE` restricts the analysis to the scenario paths listed in a file,
and `--out` writes `summary.json` and one line per attempt to `attempts.jsonl`.

## Development

```sh
cd benchmark
go test -tags=sqlite_fts5 ./...
go fmt ./...
go vet ./...
go run ./cmd/benchmark --config config.yaml --scenario scenarios/image-pull-failure/scenario.yaml
```

`--config` may be repeated to layer YAML settings. `config.yaml` holds logging,
agent and container settings; llama.cpp runtime and model settings stay in
`config.env` and `model-profiles/*.env`. `make llama-start` and
`make llama-stop` manage the local llama-server. JSON Schemas in
`benchmark/schemas/` power editor validation through `.vscode/settings.json`
(Red Hat YAML extension); the Go runner remains the source of truth.
