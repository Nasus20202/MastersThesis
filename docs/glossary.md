# Glossary

- **Adaptation condition** — one controlled version of the benchmark that differs by the model knowledge, instructions, tools or weights provided to the model. See [research-design.md](research-design.md).

- **Baseline** — the reference condition using the model with minimal task capabilities and no additional adaptation method. See [benchmark.md](benchmark.md).

- **Benchmark scenario** — one reproducible technical incident with a known-good state, an injected fault, a repair task and verification criteria. See [benchmark.md](benchmark.md).

- **Fault injection** — the controlled action that changes a known-good environment into a faulty state.

- **Full task success** — the scenario's complete success criterion has been satisfied. See [benchmark.md](benchmark.md).

- **Harness** — the operational layer that gives the model structured tools, controlled external context or additional interaction capabilities. See [tech-stack.md](tech-stack.md).

- **Incident** — an observable failure in the Kubernetes environment that the model must diagnose and repair.

- **kind** — Kubernetes in Docker, used to run local Kubernetes clusters for benchmark scenarios. See the [kind documentation](https://kind.sigs.k8s.io/docs/).

- **Model artifact** — the specific model files and revision used for inference or fine-tuning. See the [Gemma 4 E4B model card](https://huggingface.co/google/gemma-4-E4B-it).

- **Partial score** — a deterministic score representing how many independently verifiable scenario criteria were restored. See [benchmark.md](benchmark.md).

- **Prompt condition** — the benchmark condition that adds an approved system prompt to the baseline capabilities. See [research-design.md](research-design.md).

- **RAG** — retrieval-augmented generation: providing model-generated responses with context retrieved from an external document collection. See [the original RAG paper](https://arxiv.org/abs/2005.11401).

- **Raw shell** — direct command-line access, such as Bash, without replacing the interface with only predefined high-level tools. See the [Bash reference](https://www.gnu.org/software/bash/manual/).

- **Repeated run** — one independent execution of the same scenario and condition, used to measure stochastic variation. See [Towards Reproducible LLM Evaluation](https://arxiv.org/abs/2410.03492).

- **Scenario verifier** — deterministic logic that checks observable properties of the environment after the model's attempt. See [benchmark.md](benchmark.md).

- **Skill** — reusable procedural knowledge prepared for the model, such as Kubernetes inspection or troubleshooting guidance. See [research-design.md](research-design.md).

- **Source corpus** — the approved collection of documents used for retrieval, scenario construction, validation or training-data creation. See [sources.md](sources.md).

- **Tool calling** — the model's structured request to invoke an external function or operational tool. See the [llama.cpp function-calling documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/function-calling.md).

- **Training-data leakage** — inclusion of benchmark information in adaptation data in a way that can give the adapted model access to the evaluation task or answer. See [benchmark.md](benchmark.md).

- **Vulkan backend** — the graphics-compute backend used by llama.cpp for local model execution on compatible hardware. See the [Vulkan documentation](https://docs.vulkan.org/guide/latest/).
