# Glossary

- **Ablation** — a run of a condition with part of it switched off, to see what that part contributes; here the Harness without its added mechanisms. See [research-design.md](research-design.md).

- **Adaptation condition** — one controlled version of the benchmark that differs by the model knowledge, instructions, tools or weights provided to the model. See [research-design.md](research-design.md).

- **Adapter (LoRA)** — a small set of trained weights applied on top of an unchanged base model; the form of fine-tuning used here (D-046). See [the LoRA paper](https://arxiv.org/abs/2106.09685).

- **Baseline** — the reference condition using the model with minimal task capabilities and no additional adaptation method. See [benchmark.md](benchmark.md) and [Wikipedia: Scientific control](https://en.wikipedia.org/wiki/Scientific_control).

- **Benchmark scenario** — one reproducible technical incident with a known-good state, an injected fault, a repair task and verification criteria. See [benchmark.md](benchmark.md).

- **Chunk** — one piece of a corpus document that the retrieval index stores and returns as a single result. See [the retrieval design study](research/retrieval-design/README.md).

- **Bootstrap interval** — a confidence interval obtained by resampling the scenarios many times and recomputing the statistic; used for the paired difference between two conditions (D-051). See [research-design.md](research-design.md) and [Wikipedia: Bootstrapping (statistics)](https://en.wikipedia.org/wiki/Bootstrapping_%28statistics%29).

- **Decoding throughput** — generated tokens per second, aggregated from the llama.cpp timings recorded per response. See [tech-stack.md](tech-stack.md).

- **Development scenario** — one of the 46 scenarios used to build and tune the conditions and to train the adapters; its results are reported for comparison only (D-044). See [scenarios.md](scenarios.md).

- **Draft acceptance rate** — the fraction of speculatively drafted tokens accepted by the target model, recorded per condition. See [tech-stack.md](tech-stack.md).

- **Draft model** — a smaller model that proposes tokens for speculative decoding, attached to one target model. See the [llama.cpp speculative decoding documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/speculative.md) and [Wikipedia: Speculative decoding](https://en.wikipedia.org/wiki/Speculative_decoding).

- **Embedding** — a fixed-length vector representation of text whose similarity to other vectors is used for semantic retrieval. See [tech-stack.md](tech-stack.md) and [Wikipedia: Sentence embedding](https://en.wikipedia.org/wiki/Sentence_embedding).

- **False completion** — an attempt the agent ended itself, reporting it was done, with less than a full score (D-054).

- **Fault injection** — the controlled action that changes a known-good environment into a faulty state. See [Wikipedia: Fault injection](https://en.wikipedia.org/wiki/Fault_injection).

- **Full task success** — the scenario's complete success criterion has been satisfied. See [benchmark.md](benchmark.md).

- **Harness** — the code around the model that runs the tool loop and can add tools, checks and messages of its own. As a condition: skills and retrieval combined, plus checks and tools that answer measured failures (D-049). See [benchmark.md](benchmark.md).

- **Held-out scenario** — a scenario of the final experiment whose fault does not occur in the development set; never run by an agent before the final experiment, and the basis of the official result (D-044). See [research-design.md](research-design.md).

- **Hit@k** — whether any of the top k retrieval results comes from a relevant document. See [the retrieval design study](research/retrieval-design/README.md).

- **Incident** — an observable failure in the Kubernetes environment that the model must diagnose and repair. See [Wikipedia: Incident management](https://en.wikipedia.org/wiki/Incident_management).

- **kind** — Kubernetes in Docker, used to run local Kubernetes clusters for benchmark scenarios. See the [kind documentation](https://kind.sigs.k8s.io/docs/) and [Wikipedia: Kubernetes](https://en.wikipedia.org/wiki/Kubernetes).

- **Macro score** — the mean over scenarios of each scenario's mean partial score, so every scenario counts equally whatever its number of attempts. See [benchmark.md](benchmark.md).

- **Mean reciprocal rank (MRR)** — the average of 1/rank of the first relevant result, 0 when none is returned. See [Wikipedia: Mean reciprocal rank](https://en.wikipedia.org/wiki/Mean_reciprocal_rank).

- **Model artifact** — the specific model files and revision used for inference or fine-tuning. See [tech-stack.md](tech-stack.md).

- **Model profile** — a `benchmark/model-profiles/*.env` file selecting one pinned model artifact, revision, hash and router name. See [tech-stack.md](tech-stack.md).

- **Multi-token prediction (MTP)** — a speculatively decoded mode in which a drafter proposes several tokens per step (`--spec-type draft-mtp`). See the [llama.cpp speculative decoding documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/speculative.md) and [Wikipedia: Speculative decoding](https://en.wikipedia.org/wiki/Speculative_decoding).

- **Paired difference** — the difference between two conditions computed per scenario on the same scenarios, then averaged; reported with a bootstrap interval (D-051).

- **Partial score** — a deterministic score representing how many independently verifiable scenario criteria were restored. See [benchmark.md](benchmark.md).

- **Peak context** — the largest prompt plus completion of a single model response in an attempt: the most model context the attempt occupied. See [benchmark.md](benchmark.md).

- **Prompt condition** — the benchmark condition that adds an approved system prompt to the baseline capabilities. See [research-design.md](research-design.md) and [Wikipedia: Prompt engineering](https://en.wikipedia.org/wiki/Prompt_engineering).

- **Pull-through registry cache** — a local cache that mirrors upstream container registries for the kind clusters. See [tech-stack.md](tech-stack.md) and [Wikipedia: Cache (computing)](https://en.wikipedia.org/wiki/Cache_%28computing%29).

- **Quantization** — storing model weights at reduced numeric precision to lower memory use and speed up inference; each model profile pins the exact GGUF quantization of its artifact. See [tech-stack.md](tech-stack.md) and [Wikipedia: Quantization (signal processing)](https://en.wikipedia.org/wiki/Quantization_%28signal_processing%29).

- **RAG** — retrieval-augmented generation: providing model-generated responses with context retrieved from an external document collection. See [the original RAG paper](https://arxiv.org/abs/2005.11401) and [Wikipedia: Retrieval-augmented generation](https://en.wikipedia.org/wiki/Retrieval-augmented_generation).

- **Raw shell** — direct command-line access, such as Bash, without replacing the interface with only predefined high-level tools. See the [Bash reference](https://www.gnu.org/software/bash/manual/) and [Wikipedia: Bash (Unix shell)](https://en.wikipedia.org/wiki/Bash_%28Unix_shell%29).

- **Reasoning content** — the model's thinking output, retained on assistant messages between tool calls within an attempt. See [benchmark.md](benchmark.md).

- **Reciprocal rank fusion (RRF)** — combining rankings by summing 1/(k + rank) for each result across them. See [the RRF paper](https://dl.acm.org/doi/10.1145/1571941.1572114).

- **Repeated run** — one independent execution of the same scenario and condition, used to measure stochastic variation. See [Towards Reproducible LLM Evaluation](https://arxiv.org/abs/2410.03492) and [Wikipedia: Reproducibility](https://en.wikipedia.org/wiki/Reproducibility).

- **Scenario verifier** — deterministic logic that checks observable properties of the environment after the model's attempt. See [benchmark.md](benchmark.md) and [Wikipedia: Test oracle](https://en.wikipedia.org/wiki/Test_oracle).

- **Skill** — a prepared text with a troubleshooting procedure or reference for one subsystem, which the agent loads on demand through a tool. See [benchmark.md](benchmark.md) and [the skills optimization study](research/skills-optimization/README.md).

- **Source corpus** — the approved collection of documents used for retrieval, scenario construction, validation or training-data creation. See [sources.md](sources.md) and [Wikipedia: Text corpus](https://en.wikipedia.org/wiki/Text_corpus).

- **Teacher model** — the model whose successful attempts become training examples for a smaller model; here Qwen3.5 9B, with a hosted model filling gaps (D-047).

- **Tool calling** — the model's structured request to invoke an external function or operational tool. See the [llama.cpp function-calling documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/function-calling.md).

- **Trajectory** — the full record of one attempt: system prompt, task, reasoning, tool calls, tool results and final report. Successful trajectories are the fine-tuning examples (D-047).

- **Training-data leakage** — inclusion of benchmark information in adaptation data in a way that can give the adapted model access to the evaluation task or answer. See [benchmark.md](benchmark.md) and [Wikipedia: Leakage (machine learning)](https://en.wikipedia.org/wiki/Leakage_%28machine_learning%29).

- **Validation scenario (fine-tuning)** — a development scenario kept out of the training data and used to select the training configuration (D-048).

- **Vulkan backend** — the graphics-compute backend used by llama.cpp for local model execution on compatible hardware. See the [Vulkan documentation](https://docs.vulkan.org/guide/latest/) and [Wikipedia: Vulkan](https://en.wikipedia.org/wiki/Vulkan).
