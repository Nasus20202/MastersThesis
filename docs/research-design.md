# Research design

## Goal

Compare ways of adapting **Gemma 4 E4B** to diagnose and repair technical incidents that require product-specific knowledge in a local environment.

## Comparison

The primary comparison uses one fixed model and the same benchmark wherever possible:

- **baseline** — model knowledge and raw shell access,
- **prompt** — baseline with troubleshooting guidance,
- **skill** — baseline with concise prepared procedural knowledge,
- **RAG** — retrieval from a frozen technical documentation corpus,
- **fine-tuning** — LoRA trained on separate source-traceable troubleshooting examples.

Each method must be integrated as a benchmark condition before it is evaluated.

## Environment

- Gemma 4 E4B is served by **llama.cpp in Docker** with Vulkan.
- The benchmark runner and experiment orchestration are implemented in **Go**.
- Technical incidents run in a reproducible **kind** environment.
- Model commands run through raw Bash in an isolated disposable Docker sandbox without host access.
- Python is limited to fine-tuning work where the training ecosystem requires it.

Model artifacts, container versions and material runtime settings are recorded for reproducibility.

## Evaluation

Scenario scoring is deterministic and criterion-based. Partial repairs receive partial credit; complete task success is tracked separately.

Each final `scenario × condition` combination is executed multiple times. Repeated runs are used to measure both effectiveness and stability.

Primary and supporting measures include:

- partial score and full task success,
- variation across repeated runs,
- execution time,
- token and tool use where available,
- computational cost, with fine-tuning cost reported separately from inference cost.

Raw experimental results are preserved. Negative and inconclusive results are valid outcomes.

## Development and final evaluation

Development scenarios may be used to improve prompts, skills, retrieval and training choices. The final benchmark and experiment protocol are frozen before final runs and are not changed in response to final results.

## Sources

Research and technical decisions should record the sources they rely on. Prefer stable documentation versions, releases, revisions or paper identifiers and preserve enough location information to cite the source later.

Benchmark and training data must remain traceable to source material. Final benchmark incidents must not be included in fine-tuning data.
