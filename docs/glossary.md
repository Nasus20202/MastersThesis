# Glossary

- **Adaptation condition** — one controlled version of the benchmark that differs by the model knowledge, instructions, tools or weights provided to the model. See [research-design.md](research-design.md).

- **Baseline** — the reference condition using the model with minimal task capabilities and no additional adaptation method. See [benchmark.md](benchmark.md) and [Wikipedia: Scientific control](https://en.wikipedia.org/wiki/Scientific_control).

- **Benchmark scenario** — one reproducible technical incident with a known-good state, an injected fault, a repair task and verification criteria. See [benchmark.md](benchmark.md).

- **Decoding throughput** — generated tokens per second, aggregated from the llama.cpp timings recorded per response. See [tech-stack.md](tech-stack.md).

- **Draft acceptance rate** — the fraction of speculatively drafted tokens accepted by the target model, recorded per condition. See [tech-stack.md](tech-stack.md).

- **Draft model** — a smaller model that proposes tokens for speculative decoding, attached to one target model. See the [llama.cpp speculative decoding documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/speculative.md) and [Wikipedia: Speculative decoding](https://en.wikipedia.org/wiki/Speculative_decoding).

- **Fault injection** — the controlled action that changes a known-good environment into a faulty state. See [Wikipedia: Fault injection](https://en.wikipedia.org/wiki/Fault_injection).

- **Full task success** — the scenario's complete success criterion has been satisfied. See [benchmark.md](benchmark.md).

- **Harness** — the operational layer that gives the model structured tools, controlled external context or additional interaction capabilities. See [tech-stack.md](tech-stack.md) and [Wikipedia: Test harness](https://en.wikipedia.org/wiki/Test_harness).

- **Incident** — an observable failure in the Kubernetes environment that the model must diagnose and repair. See [Wikipedia: Incident management](https://en.wikipedia.org/wiki/Incident_management).

- **kind** — Kubernetes in Docker, used to run local Kubernetes clusters for benchmark scenarios. See the [kind documentation](https://kind.sigs.k8s.io/docs/) and [Wikipedia: Kubernetes](https://en.wikipedia.org/wiki/Kubernetes).

- **Model artifact** — the specific model files and revision used for inference or fine-tuning. See the [Gemma 4 E4B model card](https://huggingface.co/google/gemma-4-E4B-it).

- **Model profile** — a `benchmark/model-profiles/*.env` file selecting one pinned model artifact, revision, hash and router name. See [tech-stack.md](tech-stack.md).

- **Multi-token prediction (MTP)** — a speculatively decoded mode in which a drafter proposes several tokens per step (`--spec-type draft-mtp`). See the [llama.cpp speculative decoding documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/speculative.md) and [Wikipedia: Speculative decoding](https://en.wikipedia.org/wiki/Speculative_decoding).

- **Partial score** — a deterministic score representing how many independently verifiable scenario criteria were restored. See [benchmark.md](benchmark.md).

- **Prompt condition** — the benchmark condition that adds an approved system prompt to the baseline capabilities. See [research-design.md](research-design.md) and [Wikipedia: Prompt engineering](https://en.wikipedia.org/wiki/Prompt_engineering).

- **Pull-through registry cache** — a local cache that mirrors upstream container registries for the kind clusters. See [tech-stack.md](tech-stack.md) and [Wikipedia: Cache (computing)](https://en.wikipedia.org/wiki/Cache_%28computing%29).

- **Quantization** — storing model weights at reduced numeric precision to lower memory use and speed up inference; each model profile pins the exact GGUF quantization of its artifact. See [tech-stack.md](tech-stack.md) and [Wikipedia: Quantization (signal processing)](https://en.wikipedia.org/wiki/Quantization_%28signal_processing%29).

- **RAG** — retrieval-augmented generation: providing model-generated responses with context retrieved from an external document collection. See [the original RAG paper](https://arxiv.org/abs/2005.11401) and [Wikipedia: Retrieval-augmented generation](https://en.wikipedia.org/wiki/Retrieval-augmented_generation).

- **Raw shell** — direct command-line access, such as Bash, without replacing the interface with only predefined high-level tools. See the [Bash reference](https://www.gnu.org/software/bash/manual/) and [Wikipedia: Bash (Unix shell)](https://en.wikipedia.org/wiki/Bash_%28Unix_shell%29).

- **Reasoning content** — the model's thinking output, retained on assistant messages between tool calls within an attempt. See [benchmark.md](benchmark.md).

- **Repeated run** — one independent execution of the same scenario and condition, used to measure stochastic variation. See [Towards Reproducible LLM Evaluation](https://arxiv.org/abs/2410.03492) and [Wikipedia: Reproducibility](https://en.wikipedia.org/wiki/Reproducibility).

- **Scenario verifier** — deterministic logic that checks observable properties of the environment after the model's attempt. See [benchmark.md](benchmark.md) and [Wikipedia: Test oracle](https://en.wikipedia.org/wiki/Test_oracle).

- **Skill** — reusable procedural knowledge prepared for the model, such as Kubernetes inspection or troubleshooting guidance. See [research-design.md](research-design.md) and [Wikipedia: Skill](https://en.wikipedia.org/wiki/Skill).

- **Source corpus** — the approved collection of documents used for retrieval, scenario construction, validation or training-data creation. See [sources.md](sources.md) and [Wikipedia: Text corpus](https://en.wikipedia.org/wiki/Text_corpus).

- **Tool calling** — the model's structured request to invoke an external function or operational tool. See the [llama.cpp function-calling documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/function-calling.md).

- **Training-data leakage** — inclusion of benchmark information in adaptation data in a way that can give the adapted model access to the evaluation task or answer. See [benchmark.md](benchmark.md) and [Wikipedia: Leakage (machine learning)](https://en.wikipedia.org/wiki/Leakage_%28machine_learning%29).

- **Vulkan backend** — the graphics-compute backend used by llama.cpp for local model execution on compatible hardware. See the [Vulkan documentation](https://docs.vulkan.org/guide/latest/) and [Wikipedia: Vulkan](https://en.wikipedia.org/wiki/Vulkan).
