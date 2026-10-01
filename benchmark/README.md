# Benchmark

Go benchmark runner for the thesis project. It executes reproducible Kubernetes
operational tasks in disposable local clusters and deterministically verifies
the resulting state.

## Scenarios

A scenario defines environment setup, initial-state verification and grading;
troubleshooting scenarios may also define fault injection and verification:

```yaml
id: example-incident
title: Example technical incident
tags:
  difficulty: medium
  task_type: troubleshooting
  knowledge: documentation-dependent
  area: workloads
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

Each scenario lives in its own directory; the definition must be named
`scenario.yaml` (or `scenario.yml`):

```text
scenarios/<area>/<scenario-id>/
  scenario.yaml
  validation.yaml
  manifests/
  scripts/
  kind/
  source.md
```

The runner prepares the environment, verifies the initial state, optionally
injects and verifies a fault, runs the model (or applies a declared validation
action), grades the result and deletes the cluster. `inject_fault` and
`verify_fault` are independently optional, so a scenario can be a
troubleshooting task or a constructive task such as deploying resources.
Command, manifest and Kind config paths resolve relative to the scenario file;
without a Kind config, Kind uses its default single-node configuration.

### Cluster profiles and registry cache

`cluster.kind.config` selects a profile from `benchmark/kind/`:

- `cluster.yaml` — default kindnet, most scenarios;
- `cluster-calico.yaml` — disables the default CNI and installs Calico for NetworkPolicy scenarios;
- `cluster-registry.yaml` — default plus a mirror for the in-cluster registry used by the image scenarios.

Profiles mirror `docker.io`, `quay.io` and `ghcr.io` through shared pull-through
caches that persist images across disposable clusters. Start them with
`make registry-start` (also started by `make benchmark` and
`make benchmark-validate`) and stop with `make registry-stop`. Setup, fault
handling and grading run in a pinned `benchmark-setup` container on the kind
network; the hardened model sandbox remains separate.

### Tags and filtering

Scenarios may declare free-form evaluator-only `tags`. The development corpus
uses `difficulty` (easy, medium, hard, anchor), `task_type` (troubleshooting,
constructive), `knowledge` (general, documentation-dependent, multi-source,
version-specific) and `area`. Tags are never shown to the model.

`--tag KEY=VALUE` filters a scenario or validation run; repeat or
comma-separate it. Different keys combine with AND, repeated values of one key
with OR:

```sh
go run ./cmd/benchmark --config config.yaml --scenario scenarios/ \
  --tag difficulty=medium --tag difficulty=hard --tag area=networking
```

The filter restricts the resolved scenario or validation-case set (error when
empty) and is recorded in run metadata as `tag_selector`. Makefile:
`make benchmark TAG='difficulty=hard area=networking'`.

## Validation and results

Validation files (`validation.yaml`) reference a scenario and declare
deterministic repair cases with expected scores and full-success values:

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

`--scenario` and `--validate` accept files or directories, scanned recursively
for the matching file; other YAML is ignored. `--check-corpus PATH`
(`make check-corpus`) loads every scenario, requires the ID to match its
directory and requires a paired validation file with at least one case.

`--parallel N` bounds concurrency, `--repeat N` repeats each scenario or
validation case, and `--agent baseline,prompt` selects conditions (default
`all`). Each invocation writes one run under `results/<run-id>/`, with attempts
at `results/<run-id>/<scenario-id>/<agent-name>/<attempt>.json` and metadata in
`run.json`. Grading criteria keep command output, stderr, exit status and
duration; failed lifecycle commands record their phase, command and output in
the attempt's `failure` object.

## Results browser

`cmd/browser` is a read-only terminal UI over `results/`, loading run history
and the scenario corpus and watching the results tree:

```sh
make browser
# or: cd benchmark && go run ./cmd/browser --results results --scenarios scenarios
```

`--results`/`--scenarios` default to `results`/`scenarios`; `--no-watch`
disables live updates; the Makefile passes `BROWSER_RESULTS`/`BROWSER_SCENARIOS`.
Four tabs (**Runs**, **Agents**, **Tasks**, **Totals**) aggregate the whole
history; at 112+ columns the list sits beside a dashboard. `/` filters by
scenario tags.

## Development

```sh
cd benchmark
go test -tags=sqlite_fts5 ./...
go fmt ./...
go vet ./...
go run ./cmd/benchmark --config config.yaml --scenario scenarios/image-pull-failure/scenario.yaml
```

`--config` may be repeated to layer YAML settings; `config.yaml` holds logging,
agent and container settings while llama.cpp runtime/model settings stay in
`config.env` and `model-profiles/*.env`. JSON Schemas in `benchmark/schemas/`
power editor validation via `.vscode/settings.json` (Red Hat YAML extension),
but the Go runner is the source of truth. Use `make llama-start` /
`make llama-stop` for the local llama-server.
