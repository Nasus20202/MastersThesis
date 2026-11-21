# Glossary

- **Ablation** — a run of a condition with part of it switched off, to see what that part contributes; here the Harness without its added mechanisms. See [research-design.md](research-design.md).

- **Adaptation condition** — one controlled version of the benchmark that differs by the model knowledge, instructions, tools or weights provided to the model. See [research-design.md](research-design.md).

- **Adapter (LoRA)** — a small set of trained weights applied on top of an unchanged base model; the form of fine-tuning used here (D-053). See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#lora) and [the LoRA paper](https://arxiv.org/abs/2106.09685).

- **Alpha (LoRA)** — the scale of a LoRA adapter's change, applied as alpha / rank. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#lora).

- **Attention** — the layer step in which each token weighs earlier tokens by comparing its query with their keys and takes the weighted sum of their values. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#what-the-model-is-made-of) and [Vaswani et al., 2017](https://arxiv.org/abs/1706.03762).

- **Attention head** — one of several parallel slices of an attention layer, each with its own queries, keys and values. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#projections-by-name).

- **Baseline** — the reference condition using the model with minimal task capabilities and no additional adaptation method. See [benchmark.md](benchmark.md) and [Wikipedia: Scientific control](https://en.wikipedia.org/wiki/Scientific_control).

- **Benchmark scenario** — one reproducible technical incident with a known-good state, an injected fault, a repair task and verification criteria. See [benchmark.md](benchmark.md).

- **bfloat16** — a 16-bit number format with float32's range and less precision; training runs in it (D-055). See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#number-formats-and-memory).

- **Bootstrap interval** — a confidence interval obtained by resampling the scenarios many times and recomputing the statistic; used for the paired difference between two conditions (D-049). See [research-design.md](research-design.md) and [Wikipedia: Bootstrapping (statistics)](https://en.wikipedia.org/wiki/Bootstrapping_%28statistics%29).

- **Chat template** — a model's own text format for a conversation, with special markers per turn; training renders trajectories with the served model's template. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#from-an-attempt-to-a-training-example).

- **Chunk** — one piece of a corpus document that the retrieval index stores and returns as a single result. See [the retrieval design study](research/retrieval-design/README.md).

- **Decoding throughput** — generated tokens per second, aggregated from the llama.cpp timings recorded per response. See [tech-stack.md](tech-stack.md).

- **Development scenario** — one of the 46 scenarios used to build and tune the conditions and to train the adapters; its results are reported for comparison only (D-044). See [scenarios.md](scenarios.md).

- **Draft acceptance rate** — the fraction of speculatively drafted tokens accepted by the target model, recorded per condition. See [tech-stack.md](tech-stack.md).

- **Draft model** — a smaller model that proposes tokens for speculative decoding, attached to one target model. See the [llama.cpp speculative decoding documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/speculative.md) and [Wikipedia: Speculative decoding](https://en.wikipedia.org/wiki/Speculative_decoding).

- **Dropout (LoRA)** — setting a random fraction of the adapter's inputs to zero at each training step to reduce memorising. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#lora).

- **Embedding** — a fixed-length vector representation of text whose similarity to other vectors is used for semantic retrieval. See [tech-stack.md](tech-stack.md) and [Wikipedia: Sentence embedding](https://en.wikipedia.org/wiki/Sentence_embedding).

- **Epoch** — one pass over all training examples; with one trajectory per step, steps = epochs × examples. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#training).

- **Evaluation loss** — the loss on held-back examples, measured before and after training. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#training).

- **False completion** — an attempt the agent ended itself, reporting it was done, with less than a full score. See the [stopping-rule study](research/stopping-rule/README.md).

- **Fault injection** — the controlled action that changes a known-good environment into a faulty state. See [Wikipedia: Fault injection](https://en.wikipedia.org/wiki/Fault_injection).

- **Full task success** — the scenario's complete success criterion has been satisfied. See [benchmark.md](benchmark.md).

- **Gated attention** — Qwen3.5's attention layer: `q_proj` also makes a gate that scales each head's result before `o_proj`. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#projections-by-name).

- **Gated DeltaNet** — the linear-attention layer of Qwen3.5 (three of every four layers): a fixed-size memory updated per token, so cost grows linearly with length; each token reads its memory with a query and writes the difference to its value (the delta rule). See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#projections-by-name) and [Yang et al., 2024](https://arxiv.org/abs/2412.06464).

- **Gradient checkpointing** — keeping only each layer's input from the forward pass and recomputing the rest in the backward pass, trading compute for memory. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#number-formats-and-memory) and [Chen et al., 2016](https://arxiv.org/abs/1604.06174).

- **Harness** — the code around the model that runs the tool loop and can add tools, checks and messages of its own. As a condition: skills and retrieval combined, plus checks and tools that answer measured failures (D-048). See [benchmark.md](benchmark.md).

- **Held-out scenario** — a scenario of the final experiment whose fault does not occur in the development set; never run by an agent before the final experiment, and the basis of the official result (D-044). See [research-design.md](research-design.md).

- **Hit@k** — whether any of the top k retrieval results comes from a relevant document. See [the retrieval design study](research/retrieval-design/README.md).

- **Incident** — an observable failure in the Kubernetes environment that the model must diagnose and repair. See [Wikipedia: Incident management](https://en.wikipedia.org/wiki/Incident_management).

- **kind** — Kubernetes in Docker, used to run local Kubernetes clusters for benchmark scenarios. See the [kind documentation](https://kind.sigs.k8s.io/docs/) and [Wikipedia: Kubernetes](https://en.wikipedia.org/wiki/Kubernetes).

- **Learning rate** — the size of each weight change in training. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#training).

- **Loss (cross-entropy)** — the mean negative log probability the model gives the correct next token, over the target tokens; lower is better. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#training).

- **Macro score** — the mean over scenarios of each scenario's mean partial score, so every scenario counts equally whatever its number of attempts. See [benchmark.md](benchmark.md).

- **Mean reciprocal rank (MRR)** — the average of 1/rank of the first relevant result, 0 when none is returned. See [Wikipedia: Mean reciprocal rank](https://en.wikipedia.org/wiki/Mean_reciprocal_rank).

- **MLP** — the per-token step of each layer: `down_proj(act(gate_proj(x)) × up_proj(x))`; holds most of a layer's weights. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#projections-by-name).

- **Model artifact** — the specific model files and revision used for inference or fine-tuning. See [tech-stack.md](tech-stack.md).

- **Model profile** — a `benchmark/model-profiles/*.env` file selecting one pinned model artifact, revision, hash and router name. See [tech-stack.md](tech-stack.md).

- **Multi-token prediction (MTP)** — a speculatively decoded mode in which a drafter proposes several tokens per step (`--spec-type draft-mtp`). See the [llama.cpp speculative decoding documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/speculative.md) and [Wikipedia: Speculative decoding](https://en.wikipedia.org/wiki/Speculative_decoding).

- **NF4** — the 4-bit format QLoRA stores frozen weights in: blocks of 64 weights with one scale each, 16 levels spaced for normally distributed values. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#qlora).

- **Paired difference** — the difference between two conditions computed per scenario on the same scenarios, then averaged; reported with a bootstrap interval (D-049).

- **Partial score** — a deterministic score representing how many independently verifiable scenario criteria were restored. See [benchmark.md](benchmark.md).

- **Peak context** — the largest prompt plus completion of a single model response in an attempt: the most model context the attempt occupied. See [benchmark.md](benchmark.md).

- **Perplexity** — the exponential of the loss; measured on the served model with and without the adapter. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#training).

- **Projection** — a weight matrix that maps a token's vector to another vector; each layer type has its own named projections (`q_proj` … `o_proj` in attention, `in_proj_qkv`, `in_proj_z`, `in_proj_a`, `in_proj_b`, `out_proj` in Gated DeltaNet, `gate_proj`, `up_proj`, `down_proj` in the MLP). LoRA adapts projections by name. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#projections-by-name).

- **Prompt condition** — the benchmark condition that adds an approved system prompt to the baseline capabilities. See [research-design.md](research-design.md) and [Wikipedia: Prompt engineering](https://en.wikipedia.org/wiki/Prompt_engineering).

- **Pull-through registry cache** — a local cache that mirrors upstream container registries for the kind clusters. See [tech-stack.md](tech-stack.md) and [Wikipedia: Cache (computing)](https://en.wikipedia.org/wiki/Cache_%28computing%29).

- **QLoRA** — LoRA on a base model stored in 4 bits (NF4); less memory, slower, and the adapter trains against quantized weights. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#qlora) and [Dettmers et al., 2023](https://arxiv.org/abs/2305.14314).

- **Quantization** — storing model weights at reduced numeric precision to lower memory use and speed up inference; each model profile pins the exact GGUF quantization of its artifact. See [tech-stack.md](tech-stack.md) and [Wikipedia: Quantization (signal processing)](https://en.wikipedia.org/wiki/Quantization_%28signal_processing%29).

- **Query, key and value** — the three vectors attention makes per token (`q_proj`, `k_proj`, `v_proj`): what the token looks for, what it offers, and what it passes on. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#projections-by-name).

- **RAG** — retrieval-augmented generation: providing model-generated responses with context retrieved from an external document collection. See [the original RAG paper](https://arxiv.org/abs/2005.11401) and [Wikipedia: Retrieval-augmented generation](https://en.wikipedia.org/wiki/Retrieval-augmented_generation).

- **Rank (LoRA)** — the inner size of a LoRA adapter's two matrices: how many directions of change each projection can learn. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#lora).

- **Raw shell** — direct command-line access, such as Bash, without replacing the interface with only predefined high-level tools. See the [Bash reference](https://www.gnu.org/software/bash/manual/) and [Wikipedia: Bash (Unix shell)](https://en.wikipedia.org/wiki/Bash_%28Unix_shell%29).

- **Reasoning content** — the model's thinking output, retained on assistant messages between tool calls within an attempt. See [benchmark.md](benchmark.md).

- **Reciprocal rank fusion (RRF)** — combining rankings by summing 1/(k + rank) for each result across them. See [the RRF paper](https://dl.acm.org/doi/10.1145/1571941.1572114).

- **Repeated run** — one independent execution of the same scenario and condition, used to measure stochastic variation. See [Towards Reproducible LLM Evaluation](https://arxiv.org/abs/2410.03492) and [Wikipedia: Reproducibility](https://en.wikipedia.org/wiki/Reproducibility).

- **Scenario verifier** — deterministic logic that checks observable properties of the environment after the model's attempt. See [benchmark.md](benchmark.md) and [Wikipedia: Test oracle](https://en.wikipedia.org/wiki/Test_oracle).

- **Seed** — the number that fixes a run's random choices (example order, adapter start values). See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#training).

- **Sequence-length cap** — the longest trajectory trained on, in tokens; longer ones are dropped, never cut. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#from-an-attempt-to-a-training-example).

- **Skill** — a prepared text with a troubleshooting procedure or reference for one subsystem, which the agent loads on demand through a tool. See [benchmark.md](benchmark.md) and [the skills optimization study](research/skills-optimization/README.md).

- **Sliding-window attention** — attention over only the last few hundred tokens, interleaved in Gemma 4 with global layers that see the whole sequence. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#what-the-model-is-made-of).

- **Source corpus** — the approved collection of documents used for retrieval, scenario construction, validation or training-data creation. See [sources.md](sources.md) and [Wikipedia: Text corpus](https://en.wikipedia.org/wiki/Text_corpus).

- **Target modules** — the projections that get a LoRA adapter. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#lora).

- **Target tokens** — the tokens the loss is computed on: what the model writes when served (reasoning, tool calls, final report), not the prompt, task or tool results. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#from-an-attempt-to-a-training-example).

- **Teacher forcing** — training a model to predict each next token of a given text (here a teacher's trajectory) from the text before it, so every position is trained in one forward pass. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#how-the-model-learns-from-a-trajectory).

- **Teacher model** — the model whose successful attempts become training examples for a smaller model; here Qwen3.5 9B and DeepSeek V4.1 Flash (D-054).

- **Token** — the unit a language model reads and writes (a word, part of a word or a special marker); lengths, memory and cost are counted in tokens. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#from-an-attempt-to-a-training-example).

- **Tool calling** — the model's structured request to invoke an external function or operational tool. See the [llama.cpp function-calling documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/function-calling.md).

- **Training-data leakage** — inclusion of benchmark information in adaptation data in a way that can give the adapted model access to the evaluation task or answer. See [benchmark.md](benchmark.md) and [Wikipedia: Leakage (machine learning)](https://en.wikipedia.org/wiki/Leakage_%28machine_learning%29).

- **Trajectory** — the full record of one attempt: system prompt, task, reasoning, tool calls, tool results and final report. Successful trajectories are the fine-tuning examples (D-054).

- **Validation scenario (fine-tuning)** — a development scenario kept out of the training data and used to select the training configuration (D-055).

- **Vulkan backend** — the graphics-compute backend used by llama.cpp for local model execution on compatible hardware. See the [Vulkan documentation](https://docs.vulkan.org/guide/latest/) and [Wikipedia: Vulkan](https://en.wikipedia.org/wiki/Vulkan).

- **Warm-up** — raising the learning rate from small to full over the first training steps. See [fine-tuning concepts](research/fine-tuning-proof-of-concept/concepts.md#training).
