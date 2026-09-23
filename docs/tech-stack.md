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

The Vulkan backend is the initial execution target. The current reproducible implementation uses `ghcr.io/ggml-org/llama.cpp:server-vulkan-b10964` at digest `sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53`.

The current execution host has an AMD Ryzen 5 3600 CPU (6 cores, 12 threads), 32 GiB RAM and an AMD Radeon RX 5700 GPU with 8 GiB VRAM.

The model-serving interface and material runtime settings must be recorded for reproducibility.

## Speculative decoding

Every model profile serves with multi-token prediction enabled (`--spec-type draft-mtp`), which changes decoding speed rather than the generated text. It applies to all conditions and can be disabled with `LLAMA_SPEC_TYPE=none`.

Qwen 3.5 artifacts carry the prediction layer inside the target file. Gemma 4 publishes its drafter as a separate model, served as GGUF by the same repository as the target. Because a drafter belongs to one target, it is attached per model through the llama.cpp router preset `benchmark/models-preset.ini`; drafters are downloaded under `benchmark/models/drafts/`.

Decoding throughput and draft acceptance are recorded per condition, because acceptance depends on the model and on how repetitive a condition's output is.

## Primary model

Gemma 4 E4B is the primary model for the current experiment.

The conceptual instruction-tuned checkpoint is [google/gemma-4-E4B-it](https://huggingface.co/google/gemma-4-E4B-it). The served artifact is the [unsloth Gemma 4 E4B QAT GGUF](https://huggingface.co/unsloth/gemma-4-E4B-it-qat-GGUF), a quantization of the Google QAT unquantized weights. Exact repositories, revisions and file hashes are pinned in the model profiles, and the runtime configuration is maintained in `benchmark/config.env`. This artifact remains provisional for final evaluation and may be superseded through a recorded decision.

## Conditional alternatives

If Gemma 4 E4B cannot reliably use the required execution or tool-calling interface after the model template, llama.cpp support and the Go integration have been checked, the project may evaluate an alternative model, such as Qwen 3.5 4B, Qwen 3.5 9B or Gemma 4 12B, as a separately approved fallback.

A fallback does not change Gemma 4 E4B as the primary model. A model change must be recorded as a research decision, and results from another model must not be mixed silently with the primary comparison.

Alternative model artifacts will be recorded if the fallback plan is activated.

## Kubernetes environment

Kubernetes incidents will run in local `kind` clusters.

`kind` provides the Kubernetes environment used by benchmark scenarios. The cluster configuration, Kubernetes version, node image and workload definitions must be recorded for reproducibility.

The selected Kubernetes version and cluster configuration will be recorded before evaluation.

## Local image cache and registry

Disposable `kind` clusters do not share an image store, so the benchmark runs local registries as pull-through caches, avoiding a fresh upstream pull of the same pinned images on every attempt. The caches run as containers on the `kind` Docker network and persist their data.

Registry mirrors are part of the cluster profile: the profiles under `benchmark/kind/` declare `containerdConfigPatches` mirrors for `docker.io`, `quay.io` and `ghcr.io` that point at the matching cache with the upstream as a fallback endpoint. The runner no longer writes registry configuration into the nodes at runtime. The `default` and `calico` profiles mirror the shared caches; the specialized `registry` profile additionally mirrors the in-cluster registry that the image scenarios deploy in `prepare`.

The shared caches are local benchmark infrastructure declared in the benchmark Docker Compose stack. They are not part of any model-visible condition.

## Execution sandbox

Model commands will run in a disposable Docker sandbox.

Bash will be exposed to the model as the primary raw execution interface. The sandbox will provide the command-line tools required by the relevant benchmark condition, including `kubectl` for Kubernetes interaction.

The sandbox boundary should prevent access to the host environment. The selected filesystem, network, privilege and resource policies will be recorded before evaluation.

## Evaluator setup environment

Scenario setup, fault handling and grading run in a pinned `benchmark-setup` container instead of on the host. It reuses the sandbox Docker integration with different flags: it joins the shared `kind` network, mounts the repository read-only and the internal kubeconfig directory read-write, and runs unhardened with outbound network access. It pins `kubectl`, `helm` and `skopeo`, and the registry-seeding scripts use `skopeo` rather than a Docker daemon. This is trusted evaluator tooling, not the model boundary.

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
- local registry image digest,
- llama.cpp image and revision,
- model repository, revision and file hash,
- drafter repository, revision and file hash,
- quantization, decoding and runtime parameters,
- retrieval and embedding artifacts,
- fine-tuning adapter and training configuration,
- harness tool versions and external-access configuration.

Relevant sources and stable references should be recorded in [sources.md](sources.md). Decisions and later changes should be recorded in [decision-log.md](decision-log.md).
