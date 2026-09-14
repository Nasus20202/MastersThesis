# Sources

## Purpose

This file records the websites, documentation and papers used to support research and technical decisions.

Sources should be recorded with a stable URL and, when applicable, a version, revision, release or location. Exact revisions will be added when the corresponding benchmark or implementation choice is frozen.

## Source-recording rules

- Prefer official documentation, original papers and primary repositories.
- Record the purpose for which each source is used.
- Keep benchmark and training data traceable to their source material.
- Preserve source references used to construct or validate incidents.
- Ground-truth source references are evaluator and analysis metadata. They must not be provided to the model or used to guide retrieval. A condition may independently retrieve the same source through its approved mechanism using only model-visible information.
- Add a revision, release or retrieval date when a source is used for final evaluation.

## Corpus sources

This table records the frozen common experiment corpus.

| Corpus source                                                                                                                                                                   | Used for                                                                                                                                                                | Status                          |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------- |
| [Kubernetes website documentation subtree at the 1.37 evaluation snapshot](https://github.com/kubernetes/website/tree/ea639c1d22a60365d07b78692b1b1a2eb866bd15/content/en/docs) | Frozen corpus of English Markdown documentation pages under `content/en/docs/**/*.md`, used as source material for qualification, RAG and separate fine-tuning examples | Frozen common experiment corpus |

The common experiment corpus is the official English Kubernetes 1.37 documentation at `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`. Source excerpts and task references preserve exact file revisions, blob SHAs and line ranges.

### Kubernetes corpus licensing

The pinned Kubernetes website repository is licensed under Creative Commons Attribution 4.0 International (CC BY 4.0), as stated in its root [LICENSE](https://github.com/kubernetes/website/blob/ea639c1d22a60365d07b78692b1b1a2eb866bd15/LICENSE).

## Model and inference

| Source                                                                                                                 | Used for                                                                      | Status                            |
| ---------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- | --------------------------------- |
| [Google Gemma 4 model overview](https://ai.google.dev/gemma/docs/core)                                                 | Gemma 4 model sizes, capabilities, quantization and local deployment guidance | Initial technical reference       |
| [Gemma 4 E4B instruction-tuned model card](https://huggingface.co/google/gemma-4-E4B-it)                               | Primary model checkpoint, capabilities and model metadata                     | Primary model reference           |
| [Gemma 4 QAT Q4_0 collection](https://huggingface.co/collections/google/gemma-4-qat-q4-0)                              | Official QAT artifacts, including GGUF candidates for local inference         | Candidate runtime artifact source |
| [Gemma 4 Technical Report](https://arxiv.org/abs/2607.02770)                                                           | Research and technical background for Gemma 4                                 | Background source                 |
| [Google DeepMind Gemma repository](https://github.com/google-deepmind/gemma)                                           | Official Gemma implementation and model-related examples                      | Technical reference               |
| [llama.cpp repository](https://github.com/ggml-org/llama.cpp)                                                          | Local model inference implementation and runtime source                       | Runtime source                    |
| [llama.cpp function-calling documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/function-calling.md) | Tool-calling support and runtime/template considerations                      | Technical reference               |

## Kubernetes and cluster environment

| Source                                                       | Used for                                                 | Status              |
| ------------------------------------------------------------ | -------------------------------------------------------- | ------------------- |
| [Kubernetes documentation](https://kubernetes.io/docs/home/) | Technical reference for the initial incident environment | Technical reference |
| [Kubernetes concepts](https://kubernetes.io/docs/concepts/)  | Kubernetes concepts and resource behaviour               | Technical reference |
| [Kubernetes tasks](https://kubernetes.io/docs/tasks/)        | Operational and troubleshooting procedures               | Technical reference |
| [kind documentation](https://kind.sigs.k8s.io/docs/)         | Local Kubernetes cluster setup and operation             | Technical reference |
| [kubectl reference](https://kubernetes.io/docs/reference/kubectl/) | Command-line syntax and object inspection guidance for the kubectl skill | Technical reference |
| [Kubernetes debugging documentation](https://kubernetes.io/docs/tasks/debug/) | Evidence-first Pod and workload investigation guidance for the troubleshooting and Kubernetes skills | Technical reference |
| [Kubernetes RBAC documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) | ServiceAccount, Role, ClusterRole and binding guidance for the Kubernetes skill | Technical reference |

The project-owned skills under `skills/` are concise adaptations of these
official references. Their manifest metadata records the pinned Kubernetes
website corpus revision where applicable; they do not contain scenario names,
expected answers or evaluator-only references.

## Execution environment

| Source                                                                           | Used for                                             | Status              |
| -------------------------------------------------------------------------------- | ---------------------------------------------------- | ------------------- |
| [Docker documentation](https://docs.docker.com/)                                 | Container execution and sandboxing concepts          | Technical reference |
| [Docker Engine security documentation](https://docs.docker.com/engine/security/) | Sandbox isolation and security boundaries            | Technical reference |
| [Go documentation](https://go.dev/doc/)                                          | Go language and tooling used by the benchmark runner | Technical reference |

## Research-method sources

| Source                                                                                                                   | Used for                                                                             | Relation to this project                                                         |
| ------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------------- |
| [AgentBench: Evaluating LLMs as Agents](https://arxiv.org/abs/2308.03688)                                                | Interactive environments, multi-turn agent evaluation and task-specific environments | Directly relevant evaluation analogy; not Kubernetes-specific                    |
| [KubeLLM and KubeLLMBench](https://www.cs.utsa.edu/~plama/papers/KubeLLM_CameraReady.pdf)                                | Related Kubernetes troubleshooting agents, remediation tasks and operational metrics | Closest domain-related work; its multi-agent design is not adopted automatically |
| [ReAct: Synergizing Reasoning and Acting in Language Models](https://arxiv.org/abs/2210.03629)                           | Reasoning-and-action loops for models interacting with external environments         | Conceptual reference for agent interaction, not a scoring protocol               |
| [Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks](https://arxiv.org/abs/2005.11401)                     | Parametric and retrieved knowledge, provenance and retrieval-based adaptation        | Foundation for the RAG condition; original task differs from this benchmark      |
| [LoRA: Low-Rank Adaptation of Large Language Models](https://arxiv.org/abs/2106.09685)                                   | Parameter-efficient model adaptation for the fine-tuning condition                   | Foundation for a possible training approach, not a complete training protocol    |
| [Towards Reproducible LLM Evaluation: Quantifying Uncertainty in LLM Benchmark Scores](https://arxiv.org/abs/2410.03492) | Stochastic model outputs, repeated runs and uncertainty in benchmark scores          | Direct support for repetition and reliability considerations                     |

## Methodological interpretation

These sources inform the design of the benchmark and adaptation conditions. They do not determine the final scenario format, scoring method, repetition count or implementation.

The final protocol must be approved separately and recorded in [decision-log.md](decision-log.md).

## Source status

The corpus source, language, Kubernetes minor version and upstream revision are frozen for the common experiment conditions. Benchmark provenance and training-data eligibility remain governed by the relevant experiment decisions.
