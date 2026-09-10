# Benchmark

This directory contains the Go benchmark runner for the thesis project. The
runner is intended to execute reproducible technical incident scenarios in
disposable local Kubernetes clusters and observe whether the workload can be
restored after a fault is injected.

## Scenarios

Scenarios define the workload setup, clean-state verification, fault injection,
fault verification, weighted repair criteria and reset steps. A complete generic
scenario looks like this:

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

Each command may also define an optional `env` mapping. Command paths, manifest
paths and the optional Kind config path are resolved relative to the scenario
file. Without a Kind config, Kind uses its default single control-plane node
configuration.

The runner executes the lifecycle in this order:

1. prepare;
2. verify the clean state;
3. inject and verify the fault;
4. leave a boundary for future agent repair execution;
5. run every grading check independently;
6. reset the scenario.

Each grading check passes when its command exits with status `0`. The result
contains the criterion ID, weight, pass/fail value, stdout, stderr, exit code
and duration in seconds. The aggregate score is the passed weight divided by
total weight; `full_success` is true only when every criterion passes.

The command writes a pretty-printed JSON result to stdout and logs to stderr:

```json
{
  "grading": {
    "criteria": [
      {
        "id": "workload-ready",
        "weight": 1,
        "passed": true,
        "stdout": "healthy\n",
        "stderr": "",
        "exit_code": 0,
        "duration": 0.012345
      }
    ],
    "score": 1,
    "full_success": true
  }
}
```

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
