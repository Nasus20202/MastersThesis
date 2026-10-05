# Technology stack

## Document scope

This document describes the technologies used to build and run the benchmark.

Research motivation and comparison logic are described in [research-design.md](research-design.md). Benchmark execution and scoring are described in [benchmark.md](benchmark.md). Supporting references are collected in [sources.md](sources.md).

## Benchmark runner

The benchmark runner and experiment orchestration are implemented in Go as a CLI tool.

The runner coordinates the benchmark lifecycle, invokes model execution, exposes condition-specific capabilities, collects verification results and preserves raw run data.

## Model execution

The model is served locally through `llama.cpp` in Docker.

The Vulkan backend is the initial execution target. The current reproducible implementation uses `ghcr.io/ggml-org/llama.cpp:server-vulkan-b10964` at digest `sha256:43e0e25ca654d839ebda39fd6c2f200b36e9efb3e597ba90d0aaeff1be95ca53`.

The current execution host has an AMD Ryzen 5 3600 CPU (6 cores, 12 threads), 32 GiB RAM and an AMD Radeon RX 5700 GPU with 8 GiB VRAM.

## Speculative decoding

Every model profile serves with multi-token prediction enabled (`--spec-type draft-mtp`), which changes decoding speed rather than the generated text. It applies to all conditions and can be disabled with `LLAMA_SPEC_TYPE=none`.

Qwen 3.5 artifacts carry the prediction layer inside the target file. Gemma 4 publishes its drafter as a separate model, served as GGUF by the same repository as the target. Because a drafter belongs to one target, it is attached per model through the llama.cpp router preset `benchmark/models-preset.ini`; drafters are downloaded under `benchmark/models/drafts/`.

Decoding throughput and draft acceptance are recorded per condition, because acceptance depends on the model and on how repetitive a condition's output is.

## Retrieval

The RAG condition searches the frozen corpus through SQLite indexes built by `benchmark/cmd/retrieval`: FTS5 for lexical (BM25) search and sqlite-vec for exact vector search, combined by reciprocal rank fusion in hybrid mode. SQLite is accessed through `mattn/go-sqlite3` built with the `sqlite_fts5` tag; sqlite-vec is linked through its cgo bindings.

Embeddings are produced by EmbeddingGemma 300M, served by a separate `llama-embedding` llama.cpp service (port 8081, same pinned image as the model router), so embedding does not occupy a model-router slot. The model, revision and hash are pinned in `benchmark/retrieval.env`. Indexes are gitignored, rebuilt with `make retrieval-index`, and record the corpus revision and embedding artifact they were built from. The selected configuration is described in [the retrieval design study](research/retrieval-design/README.md).

## Models

The research uses four models, each selected by a profile under `benchmark/model-profiles/` (D-037):

| Profile       | Model       | Served artifact                                                                        |
| ------------- | ----------- | -------------------------------------------------------------------------------------- |
| `gemma-4-e2b` | Gemma 4 E2B | [unsloth Gemma 4 E2B QAT GGUF](https://huggingface.co/unsloth/gemma-4-E2B-it-qat-GGUF) |
| `gemma-4-e4b` | Gemma 4 E4B | [unsloth Gemma 4 E4B QAT GGUF](https://huggingface.co/unsloth/gemma-4-E4B-it-qat-GGUF) |
| `qwen35-4b`   | Qwen3.5 4B  | [unsloth Qwen3.5 4B MTP GGUF](https://huggingface.co/unsloth/Qwen3.5-4B-MTP-GGUF)      |
| `qwen35-9b`   | Qwen3.5 9B  | [unsloth Qwen3.5 9B MTP GGUF](https://huggingface.co/unsloth/Qwen3.5-9B-MTP-GGUF)      |

Not every run must use all four. Results are reported per model and never mixed silently across models. Exact repositories, revisions and file hashes are pinned in the profiles, and the shared runtime configuration is maintained in `benchmark/config.env`. The artifacts remain provisional for final evaluation and may be superseded through a recorded decision.

## Sampling

Each model uses its vendor's recommended sampling, the same in every condition (D-038). The values are set per model in the router preset `benchmark/models-preset.ini`; parameters the vendor does not name are disabled rather than left to llama.cpp defaults.

| Model   | Temperature | Top-p | Top-k | Min-p | Presence penalty | Source                                                       |
| ------- | ----------: | ----: | ----: | ----: | ---------------: | ------------------------------------------------------------ |
| Gemma 4 |         1.0 |  0.95 |    64 |     0 |                0 | Model card "Sampling Parameters"                             |
| Qwen3.5 |         1.0 |  0.95 |    20 |     0 |              1.5 | Model card "Best Practices", thinking mode for general tasks |

Before the first attempt, a benchmark run reads the values llama-server applies to the profile's model from `/props` and stores them in `run.json`, so a run records the sampling it actually used. The router reads the preset at startup: after changing it, recreate the server with `make llama-stop llama-start`.

## Kubernetes environment

Kubernetes incidents run in local `kind` clusters.

## Local image cache and registry

Disposable `kind` clusters do not share an image store, so the benchmark runs local registries as pull-through caches, avoiding a fresh upstream pull of the same pinned images on every attempt. The caches run as containers on the `kind` Docker network and persist their data.

Registry mirrors are part of the cluster profile: the profiles under `benchmark/kind/` declare `containerdConfigPatches` mirrors for `docker.io`, `quay.io` and `ghcr.io` that point at the matching cache with the upstream as a fallback endpoint. The `default` and `calico` profiles mirror the shared caches; the specialized `registry` profile additionally mirrors the in-cluster registry that the image scenarios deploy in `prepare`.

The shared caches are local benchmark infrastructure declared in the benchmark Docker Compose stack. They are not part of any model-visible condition.

## Execution sandbox

Model commands run in a disposable Docker sandbox.

Bash is exposed to the model as the primary raw execution interface. The sandbox provides the command-line tools required by the relevant benchmark condition, including `kubectl` for Kubernetes interaction.

The sandbox has no interactive editor. `KUBE_EDITOR` points at a script that makes `kubectl edit` fail at once with a hint to export the manifest, change it and apply it (D-041). The `read_file`, `write_file` and `edit_file` tools run through the same sandbox boundary as bash.

The sandbox image is built only when it is missing; after changing its Dockerfile, remove the image so the next run rebuilds it.

The sandbox boundary should prevent access to the host environment.

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
- Kubernetes version, node image and cluster configuration,
- local registry image digest,
- llama.cpp image and revision,
- sandbox filesystem, network, privilege and resource policies,
- model repository, revision and file hash,
- drafter repository, revision and file hash,
- quantization, decoding and runtime parameters,
- retrieval and embedding artifacts,
- fine-tuning adapter and training configuration,
- harness tool versions and external-access configuration.
