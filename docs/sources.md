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

This table records the approved initial candidate source corpus. It does not freeze the exact Kubernetes minor version, documentation snapshot or upstream repository revision; those remain open for candidate qualification in #6.

| Corpus source | Used for | Status |
|---|---|---|
| [Kubernetes website repository](https://github.com/kubernetes/website) | Upstream source for the full official vanilla Kubernetes documentation corpus used directly by RAG and as source material for separate fine-tuning examples | Approved initial candidate; exact minor version and repository revision open |

The final corpus version must match the benchmark Kubernetes minor version and be frozen to an exact upstream repository revision before qualification and evaluation.

## Corpus qualification requirements

Candidate qualification for #6 should verify that the corpus:

- provides product- and version-specific knowledge that adds value beyond Gemma 4 E4B pretraining, especially version changes, configuration semantics and operational procedures;
- has sufficient breadth across Kubernetes concepts, configuration, workloads, networking, storage, security, observability, troubleshooting and release behavior to support varied incidents;
- has traceable provenance, including the upstream source, document path, version or revision, acquisition date and reproducible snapshot;
- has clear licensing and attribution requirements, with the Kubernetes website documentation recorded as CC BY 4.0 and embedded third-party material or code samples reviewed separately;
- provides stable references that can support later citations at document or section level;
- supports realistic, observable and resettable incidents that can be exercised in `kind` without proprietary infrastructure;
- is suitable for RAG retrieval from the complete corpus; and
- can support separate, human-reviewed, source-traceable troubleshooting examples for fine-tuning without using raw documentation directly or leaking final benchmark incidents.

## Model and inference

| Source | Used for | Status |
|---|---|---|
| [Google Gemma 4 model overview](https://ai.google.dev/gemma/docs/core) | Gemma 4 model sizes, capabilities, quantization and local deployment guidance | Initial technical reference |
| [Gemma 4 E4B instruction-tuned model card](https://huggingface.co/google/gemma-4-E4B-it) | Primary model checkpoint, capabilities and model metadata | Primary model reference |
| [Gemma 4 QAT Q4_0 collection](https://huggingface.co/collections/google/gemma-4-qat-q4-0) | Official QAT artifacts, including GGUF candidates for local inference | Candidate runtime artifact source |
| [Gemma 4 Technical Report](https://arxiv.org/abs/2607.02770) | Research and technical background for Gemma 4 | Background source |
| [Google DeepMind Gemma repository](https://github.com/google-deepmind/gemma) | Official Gemma implementation and model-related examples | Technical reference |
| [llama.cpp repository](https://github.com/ggml-org/llama.cpp) | Local model inference implementation and runtime source | Runtime source |
| [llama.cpp function-calling documentation](https://github.com/ggml-org/llama.cpp/blob/master/docs/function-calling.md) | Tool-calling support and runtime/template considerations | Technical reference |

## Kubernetes and cluster environment

| Source | Used for | Status |
|---|---|---|
| [Kubernetes documentation](https://kubernetes.io/docs/home/) | Technical reference for the initial incident environment | Technical reference |
| [Kubernetes concepts](https://kubernetes.io/docs/concepts/) | Kubernetes concepts and resource behaviour | Technical reference |
| [Kubernetes tasks](https://kubernetes.io/docs/tasks/) | Operational and troubleshooting procedures | Technical reference |
| [kind documentation](https://kind.sigs.k8s.io/docs/) | Local Kubernetes cluster setup and operation | Technical reference |

## Execution environment

| Source | Used for | Status |
|---|---|---|
| [Docker documentation](https://docs.docker.com/) | Container execution and sandboxing concepts | Technical reference |
| [Docker Engine security documentation](https://docs.docker.com/engine/security/) | Sandbox isolation and security boundaries | Technical reference |
| [Go documentation](https://go.dev/doc/) | Go language and tooling used by the benchmark runner | Technical reference |

## Research-method sources

| Source | Used for | Relation to this project |
|---|---|---|
| [AgentBench: Evaluating LLMs as Agents](https://arxiv.org/abs/2308.03688) | Interactive environments, multi-turn agent evaluation and task-specific environments | Directly relevant evaluation analogy; not Kubernetes-specific |
| [KubeLLM and KubeLLMBench](https://www.cs.utsa.edu/~plama/papers/KubeLLM_CameraReady.pdf) | Related Kubernetes troubleshooting agents, remediation tasks and operational metrics | Closest domain-related work; its multi-agent design is not adopted automatically |
| [ReAct: Synergizing Reasoning and Acting in Language Models](https://arxiv.org/abs/2210.03629) | Reasoning-and-action loops for models interacting with external environments | Conceptual reference for agent interaction, not a scoring protocol |
| [Retrieval-Augmented Generation for Knowledge-Intensive NLP Tasks](https://arxiv.org/abs/2005.11401) | Parametric and retrieved knowledge, provenance and retrieval-based adaptation | Foundation for the RAG condition; original task differs from this benchmark |
| [LoRA: Low-Rank Adaptation of Large Language Models](https://arxiv.org/abs/2106.09685) | Parameter-efficient model adaptation for the fine-tuning condition | Foundation for a possible training approach, not a complete training protocol |
| [Towards Reproducible LLM Evaluation: Quantifying Uncertainty in LLM Benchmark Scores](https://arxiv.org/abs/2410.03492) | Stochastic model outputs, repeated runs and uncertainty in benchmark scores | Direct support for repetition and reliability considerations |

## Methodological interpretation

These sources inform the design of the benchmark and adaptation conditions. They do not determine the final scenario format, scoring method, repetition count or implementation.

The final protocol must be approved separately and recorded in [decision-log.md](decision-log.md).

## Source status

The Corpus sources table identifies the current candidate direction. Exact corpus membership, snapshot, benchmark provenance and training-data eligibility must be decided separately during qualification and freezing.
