# Research design

## Goal

Compare ways of adapting **Gemma 4 E4B** to diagnose and repair technical incidents that require product-specific knowledge in a local environment.

## Conditions

- baseline — model knowledge and raw shell access,
- prompt — troubleshooting guidance,
- skill — concise prepared procedural knowledge,
- RAG — retrieval from a frozen technical documentation corpus,
- fine-tuning — LoRA trained on separate source-traceable troubleshooting examples,
- harness — richer tools and external knowledge access.

## Measures

Primary outcome: whether the incident is repaired. Secondary measures include execution cost, stability and tool use.

The model, benchmark, source corpus and final experiment settings are frozen before final runs.
