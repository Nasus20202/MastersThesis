# Technology stack

## Document scope

This document describes the technologies used to build and run the benchmark.

Research motivation and comparison logic are described in [research-design.md](research-design.md). Benchmark execution and scoring are described in [benchmark.md](benchmark.md). Supporting references will be collected in [sources.md](sources.md).

The stack contains both current choices and provisional candidates. Exact versions, revisions and artifacts are tracked in the [decision log](decision-log.md) and should be recorded before reproducible evaluation.

## Benchmark runner

The benchmark runner and experiment orchestration will be implemented in Go.

The runner is responsible for coordinating the benchmark lifecycle, invoking model execution, exposing condition-specific capabilities, collecting verification results and preserving raw run data.

The initial implementation should remain a sensible CLI tool with a small and understandable structure.

## Model execution

The model will be served locally through `llama.cpp` in Docker.

The Vulkan backend is the initial execution target. The current reproducible implementation uses `ghcr.io/ggml-org/llama.cpp:server-vulkan-b10524` at digest `sha256:1f834a5e0248d83c4cc459591a53250499592742f3c33f8830054e9ccb1c4c5b`. Its runtime parameters are maintained in `benchmark/config.env`; they remain provisional for final evaluation.

The current execution host has an AMD Ryzen 5 3600 CPU (6 cores, 12 threads), 32 GiB RAM and an AMD Radeon RX 5700 GPU with 8 GiB VRAM.

The model-serving interface and material runtime settings must be recorded for reproducibility.

## Primary model

Gemma 4 E4B is the primary model for the current experiment.

The conceptual instruction-tuned checkpoint is [google/gemma-4-E4B-it](https://huggingface.co/google/gemma-4-E4B-it).

The current reproducible implementation uses `gemma-4-E4B_q4_0-it.gguf` from [google/gemma-4-E4B-it-qat-q4_0-gguf](https://huggingface.co/google/gemma-4-E4B-it-qat-q4_0-gguf) at repository revision `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`. The downloaded file has SHA-256 `676c35070db6dbe52f93e9c864ee0fba4eddea94b9c875d9cb10daff453fbaee`. This artifact remains provisional for final evaluation and may be superseded through a recorded decision.

The selected artifact and runtime configuration are maintained in `benchmark/config.env`.

## Conditional alternatives

If Gemma 4 E4B cannot reliably use the required execution or tool-calling interface after the model template, llama.cpp support and the Go integration have been checked, the project may evaluate an alternative model, such as Qwen 3.5 4B, Qwen 3.5 9B or Gemma 4 12B, as a separately approved fallback.

A fallback does not change Gemma 4 E4B as the primary model. A model change must be recorded as a research decision, and results from another model must not be mixed silently with the primary comparison.

Alternative model artifacts will be recorded if the fallback plan is activated.

## Kubernetes environment

Kubernetes incidents will run in local `kind` clusters.

`kind` provides the Kubernetes environment used by benchmark scenarios. The cluster configuration, Kubernetes version, node image and workload definitions must be recorded for reproducibility.

The selected Kubernetes version and cluster configuration will be recorded before evaluation.

## Execution sandbox

Model commands will run in a disposable Docker sandbox.

Bash will be exposed to the model as the primary raw execution interface. The sandbox will provide the command-line tools required by the relevant benchmark condition, including `kubectl` for Kubernetes interaction.

The sandbox boundary should prevent access to the host environment. The selected filesystem, network, privilege and resource policies will be recorded before evaluation.

## Adaptation-specific components

The adaptation methods may add the following components:

- prompt condition: an approved system prompt,
- skill condition: prepared procedural skill content,
- RAG condition: an embedding and retrieval pipeline over the frozen common experiment corpus recorded in [sources.md](sources.md),
- fine-tuning condition: a separately trained model adapter and its training tooling,
- harness condition: predefined operational tools and controlled external context access.

These components are benchmark conditions, not separate primary execution stacks. They should reuse the common runner, environment and result format wherever possible.

## Version and provenance records

Before final evaluation, record:

- Go version and runner revision,
- Docker version and image digests,
- kind version,
- Kubernetes version and node image,
- llama.cpp image and revision,
- model repository, revision and file hash,
- quantization and runtime parameters,
- retrieval and embedding artifacts,
- fine-tuning adapter and training configuration,
- harness tool versions and external-access configuration.

Relevant sources and stable references should be recorded in [sources.md](sources.md). Decisions and later changes should be recorded in [decision-log.md](decision-log.md).
