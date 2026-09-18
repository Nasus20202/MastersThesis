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

## Validation and results

Validation manifests are YAML files that reference scenario files and declare
deterministic repair cases with expected scores and full-success values. A
directory input is scanned recursively for `.yaml` and `.yml` files; unrelated
YAML files are skipped. For example:

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

The `--scenario` and `--validate` options accept files or directories. Directory
inputs are scanned recursively for YAML files. Use `--parallel N` to bound
concurrent attempts and `--repeat N` to run each scenario or validation case
more than once. Scenario runs select agents with `--agent baseline,prompt` or
repeated flags such as `--agent baseline --agent prompt`; the default `all`
selection runs both agents.

Each benchmark invocation writes one logical run under `results/<run-id>/`,
with agent attempt evidence at
`results/<run-id>/<scenario-id>/<agent-name>/<attempt>.json`. Run metadata is
stored in `results/<run-id>/run.json`. The attempt ID is zero-padded, for
example `001.json`. Grading criteria preserve command output, stderr, exit
status and duration. Failed lifecycle commands also record their phase,
command details and captured output in the attempt's `failure` object.

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
contains benchmark-specific logging and agent settings. Llama.cpp runtime and
model settings remain in `config.env` and `model-profiles/*.env`; the Makefile
loads those for the benchmark and Compose.

Start and stop the local llama-server service from the benchmark directory:

```sh
cd benchmark
docker compose --env-file config.env up -d
docker compose --env-file config.env down
```
