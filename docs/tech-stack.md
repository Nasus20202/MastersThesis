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

The Vulkan backend is the initial execution target. The selected llama.cpp image, revision and runtime parameters will be recorded before evaluation.

The model-serving interface and material runtime settings must be recorded for reproducibility.

## Initial model candidate

The initial model candidate is Gemma 4 E4B.

The model is a starting point for the experiment rather than a permanent project requirement. The conceptual instruction-tuned checkpoint is [google/gemma-4-E4B-it](https://huggingface.co/google/gemma-4-E4B-it).

An official QAT GGUF candidate for local llama.cpp execution is [google/gemma-4-E4B-it-qat-q4_0-gguf](https://huggingface.co/google/gemma-4-E4B-it-qat-q4_0-gguf). This is a provisional runtime candidate, not yet a frozen experiment artifact.

The exact model revision, file hash, quantization and runtime configuration will be selected and recorded later.

## Alternative models

If the initial model cannot reliably use the required execution or tool-calling interface, the project may evaluate an alternative model, such as Gemma 4 12B or Qwen 3.5 9B.

A tool-calling failure should first be checked against the model template, llama.cpp support and the Go integration. A model change must be recorded as a research decision and should not be mixed silently with results from another model.

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
- RAG condition: an embedding and retrieval pipeline over the approved Kubernetes documentation corpus,
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
