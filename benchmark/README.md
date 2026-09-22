# Benchmark

This directory contains the Go benchmark runner for the thesis project. The
runner is intended to execute reproducible Kubernetes operational tasks in
disposable local clusters and deterministically verify the resulting state.

## Scenarios

Scenarios define the environment setup, initial-state verification and grading.
Troubleshooting scenarios may additionally define fault injection and fault
verification. A generic troubleshooting scenario looks like this:

```yaml
id: example-incident
title: Example technical incident
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

Each scenario lives in its own directory and its definition must be named
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

The runner performs these steps in order:

1. prepare;
2. verify the initial state;
3. optionally inject and verify a fault;
4. run the model, or apply a declared action in validation mode;
5. grade the resulting state;
6. delete the disposable cluster.

`inject_fault` and `verify_fault` are independently optional. A scenario may
use either, both, or neither depending on how its pre-model state is prepared
and verified. Scenarios without them can represent constructive tasks such as
deploying or configuring Kubernetes resources.

Command paths, manifest paths and the optional Kind config path are resolved
relative to the scenario file. Without a Kind config, Kind uses its default
single control-plane node configuration.

### Cluster profiles and registry cache

`cluster.kind.config` selects a cluster profile from `benchmark/kind/`:

- `cluster.yaml` — the default kindnet profile, used by most scenarios;
- `cluster-calico.yaml` — disables the default CNI and installs Calico for scenarios that need NetworkPolicy enforcement;
- `cluster-registry.yaml` — the default profile plus a mirror for the in-cluster registry that the image scenarios deploy in `prepare`.

The profiles declare `containerdConfigPatches` mirrors for `docker.io`,
`quay.io` and `ghcr.io` that point at shared pull-through registry caches, with
the upstream kept as a fallback endpoint. The caches run as containers on the
`kind` Docker network and persist their data, so pinned images are pulled once
and reused across disposable clusters. Start and stop them with
`make registry-start` and `make registry-stop`; `make benchmark` and
`make benchmark-validate` start them automatically.

Scenario setup, fault handling and grading run in a pinned `benchmark-setup`
container on the shared kind network, not in the model sandbox. The model
sandbox remains the hardened agent boundary.

## Validation and results

Validation manifests are YAML files named `validation.yaml` (or
`validation.yml`) that reference scenario files and declare deterministic
repair cases with expected scores and full-success values. For example:

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

The `--scenario` and `--validate` options accept files or directories. A
directory is scanned recursively for `scenario.yaml`/`.yml` during scenario
runs and `validation.yaml`/`.yml` during validation runs; other YAML files,
such as manifests, are ignored. A discovered file that fails to load is
reported as an error. Explicit file paths may use any name.

`--check-corpus PATH` validates the whole corpus: it loads every scenario,
requires the scenario ID to match its directory name, and requires a paired
validation file with at least one case. A missing or unexpectedly named
`scenario.yaml` or `validation.yaml` fails the check. Run it with
`make check-corpus`.

Use `--parallel N` to bound concurrent attempts and `--repeat N` to run each
scenario or validation case more than once. Scenario runs select agents with
`--agent baseline,prompt` or repeated flags such as
`--agent baseline --agent prompt`; the default `all` selection runs both
agents.

Each benchmark invocation writes one logical run under `results/<run-id>/`,
with agent attempt evidence at
`results/<run-id>/<scenario-id>/<agent-name>/<attempt>.json`. Run metadata is
stored in `results/<run-id>/run.json`. The attempt ID is zero-padded, for
example `001.json`. Grading criteria preserve command output, stderr, exit
status and duration. Failed lifecycle commands also record their phase,
command details and captured output in the attempt's `failure` object.

## Results browser

`cmd/browser` is a read-only terminal UI over the runs under `results/`. It
loads `results/<run-id>/` together with the scenario corpus for task titles and
watches the results tree, so a running benchmark appears as it writes.

```sh
make browser
# or
cd benchmark
go run ./cmd/browser --results results --scenarios scenarios
```

`--results` and `--scenarios` default to `results` and `scenarios`; `--no-watch`
disables live updates. The Makefile passes `BROWSER_RESULTS` and
`BROWSER_SCENARIOS` when set.

The browser has three tabs — **Runs**, **Agents** and **Tasks** — each
aggregated over the whole history. On terminals at least 112 columns wide the
selected list is shown beside a dashboard; narrower terminals show one pane.

## Development

Run Go commands from the benchmark module directory:

```sh
cd benchmark
go test ./...
go test ./... -cover
go fmt ./...
go vet ./...
go run ./cmd/benchmark \
  --config config.yaml \
  --scenario scenarios/image-pull-failure/scenario.yaml
```

Benchmark configuration files use YAML and may be layered by repeating
`--config`; later files override earlier files. The default `config.yaml`
contains logging, agent and container settings. Llama.cpp runtime and model
settings remain in `config.env` and `model-profiles/*.env`; the Makefile loads
those for the benchmark and Compose.

JSON Schemas for `scenario.yaml`, `validation.yaml` and `config.yaml` live in
`benchmark/schemas/`. VS Code applies them through `.vscode/settings.json` when
the Red Hat YAML extension (`redhat.vscode-yaml`, recommended in
`.vscode/extensions.json`) is installed. The schemas are an editor aid; the Go
runner remains the source of truth.

Start and stop the local llama-server service from the benchmark directory:

```sh
make llama-start
make llama-stop
```
