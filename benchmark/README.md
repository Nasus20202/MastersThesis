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

verify_clean:
  - program: kubectl
    args: [rollout, status, deployment/app, --timeout=60s]

inject_fault:
  - program: ./scripts/inject-fault.sh

verify_fault:
  - program: ./scripts/verify-fault.sh

reset:
  - program: kubectl
    args: [apply, -f, manifests/app.yaml]
  - program: kubectl
    args: [rollout, status, deployment/app, --timeout=60s]
```

Command paths, manifest paths and the optional Kind config path are resolved
relative to the scenario file. Without a Kind config, Kind uses its default
single control-plane node configuration.

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
