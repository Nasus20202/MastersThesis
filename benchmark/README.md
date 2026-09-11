# Benchmark

This directory contains the Go benchmark runner for the thesis project. The
runner is intended to execute reproducible technical incident scenarios in
disposable local Kubernetes clusters and observe whether the workload can be
restored after a fault is injected.

## Scenarios

Scenarios define the workload setup, clean-state verification, fault injection,
fault verification and reset steps. A complete generic scenario looks like
this:

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

reset:
  - program: kubectl
    args: [apply, -f, manifests/app.yaml]
  - program: kubectl
    args: [rollout, status, deployment/app, --timeout=60s]
```

The runner performs these steps in order:

1. prepare;
2. verify the clean state;
3. inject the fault;
4. verify the fault;
5. leave a boundary for future model repair, or apply a declared repair in validation mode;
6. grade the repaired state;
7. reset the scenario.

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
more than once.

Each run writes machine-readable evidence under `results/<run-id>/`, including
run metadata and one JSON file per attempt. Grading criteria preserve command
stdout, stderr, exit status and duration. Failed lifecycle commands also record
their phase, command details and captured output in the attempt's `failure`
object.

## Development

Run Go commands from the benchmark module directory:

```sh
cd benchmark
go test ./...
go test ./... -cover
go fmt ./...
go vet ./...
go run ./cmd/benchmark --scenario scenarios/image-pull-failure/scenario.yaml
```

Start and stop the local llama-server service from the benchmark directory:

```sh
cd benchmark
docker compose --env-file config.env up -d
docker compose --env-file config.env down
```
