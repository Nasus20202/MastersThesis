# Benchmark

The benchmark consists of reproducible technical incidents in a controlled environment. A scenario starts from a known-good state, injects a real fault, lets the model attempt a repair and then verifies the resulting system behaviour.

## Scenario lifecycle

A scenario follows the same high-level lifecycle:

1. prepare a known-good state,
2. inject the fault,
3. give the model the task and capabilities of its experimental condition,
4. let the model inspect and modify the environment,
5. verify the final state,
6. preserve the result and reset the environment.

A scenario is accepted only when the clean state passes, the injected fault lowers the score, an approved repair restores the expected state and the lifecycle is repeatable.

## Scoring

Scoring is deterministic and criterion-based. Independent observable outcomes contribute to the final score, so a partial repair receives partial credit.

Verification should prefer system behaviour over one exact command or configuration representation. Full task success is tracked separately from the partial score.

## Model visibility

The model sees only the task and the capabilities available in the current condition. It must not see the injected fault, expected root cause, verifier logic or ground-truth source references.

Each adaptation method is implemented as a benchmark condition so that the same scenarios, scoring and result format can be reused across baseline, prompt, skill, RAG, fine-tuning and harness evaluations.

## Sources and separation

Documentation-dependent scenarios must be traceable to the frozen source corpus.

RAG searches the complete approved corpus rather than scenario-specific excerpts. Ground-truth source references are used for validation and analysis only and must not guide retrieval.

Fine-tuning data must remain separate from benchmark incidents. Final benchmark incidents must not appear in the training data.

Sources used to construct or validate scenarios should retain enough version, revision and location information to support later citation.

## Coverage

The benchmark should cover varied technical areas and fault types rather than many variants of the same mistake. It should include general, documentation-dependent, multi-source and version-specific incidents.

Development scenarios are used while building and evaluating the methods. The final benchmark is frozen before the final experiment.

## Repeated evaluation

LLM behaviour is stochastic. Final `scenario × condition` combinations are therefore executed multiple times.

Each run preserves the deterministic criterion results, final score, full-success status, model/tool transcript and material runtime metadata needed to explain and reproduce the result.
