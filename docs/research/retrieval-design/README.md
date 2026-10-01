# Retrieval design

Selection of the retrieval configuration for the RAG condition: search mode,
chunking and top-k over the frozen Kubernetes documentation corpus (D-017). The
question is which configuration most often returns a documentation file that a
knowledge probe cites, when the query is the kind of short search an agent
writes.

## Method

The complete corpus (1,717 Markdown files) is indexed in one SQLite file per
chunking. Both indexes use the same corpus snapshot and embedding model.

| Parameter        | Value                                                                                                          |
| ---------------- | -------------------------------------------------------------------------------------------------------------- |
| Lexical search   | SQLite FTS5, `porter` tokenizer, BM25; query terms quoted and joined with `OR`                                 |
| Semantic search  | sqlite-vec exact cosine search over EmbeddingGemma 300M (QAT, Q8_0) embeddings, 768 dimensions                 |
| Hybrid search    | reciprocal rank fusion (k = 60) of the lexical and semantic top 50                                             |
| Embedding prompt | model-card prompts: `title: {title} \| text: {chunk}` and `task: search result \| query: {query}`              |
| Section chunks   | one chunk per Markdown heading section, oversized sections split at paragraphs; at most 1.5 KiB; 21,757 chunks |
| Window chunks    | 1.5 KiB windows with 256 B overlap; 14,221 chunks                                                              |
| Top-k            | 3 and 5                                                                                                        |
| Result cap       | 8 KiB of model-visible text, matching the Bash output cap (D-027)                                              |

The 3 × 2 × 2 grid gives 12 configurations.

**Queries.** The 24 knowledge-check probes
([context-tasks.json](../kubernetes-knowledge-check/context-tasks.json)) were
used in two forms: the full probe text (_as-is_), and three short search queries
per probe written by Gemma 4 E4B (_rewrite_; temperature 0, seed 42, reasoning
enabled). The rewrites approximate the queries an agent would issue, so they are
the primary form. They were generated once, reviewed for format and frozen
before evaluation in [queries.json](queries.json), which also records the
instruction, model artifact and raw responses. Benchmark scenarios and their
source references were not used.

**Relevance.** A result is relevant when its chunk comes from a corpus file the
probe cites (source, criteria or provided context). A probe's _primary_ files
are the sources it was written from.

**Metrics.** Per query: hit@k (any relevant file in the top k), reciprocal rank
of the first relevant result, and hit@k on primary files. Rewrite scores are
averaged over a probe's three queries, then over probes. Results longer than the
8 KiB cap are counted as truncated.

**Selection rule** (fixed before the evaluation): highest mean rewrite hit@k;
configurations within 1/24 (one probe) of the best are tied, and ties are broken
by rewrite MRR, then smaller k, then the simpler mode (lexical < semantic <
hybrid), then sections before windows. The evaluation also reports each
configuration's rewrite hit@k difference from the selected one, with a 95%
paired percentile bootstrap interval over probes (10,000 resamples, seed 1).

## Results

| Configuration          | hit@k rewrite | MRR rewrite | primary rewrite | hit@k as-is | MRR as-is | primary as-is |
| ---------------------- | ------------: | ----------: | --------------: | ----------: | --------: | ------------: |
| lexical/sections/k=3   |         0.625 |       0.502 |           0.417 |       0.750 |     0.618 |         0.625 |
| lexical/sections/k=5   |         0.750 |       0.529 |           0.514 |       0.833 |     0.639 |         0.750 |
| semantic/sections/k=3  |         0.667 |       0.528 |           0.472 |       0.750 |     0.625 |         0.667 |
| semantic/sections/k=5  |         0.764 |       0.551 |           0.583 |       0.917 |     0.667 |         0.708 |
| hybrid/sections/k=3    |         0.681 |       0.567 |           0.403 |       0.875 |     0.743 |         0.708 |
| hybrid/sections/k=5    |         0.736 |       0.580 |           0.472 |       0.958 |     0.762 |         0.792 |
| lexical/windows/k=3    |         0.639 |       0.495 |           0.472 |       0.833 |     0.715 |         0.667 |
| lexical/windows/k=5    |         0.722 |       0.513 |           0.514 |       0.875 |     0.724 |         0.708 |
| semantic/windows/k=3   |         0.708 |       0.521 |           0.500 |       0.875 |     0.736 |         0.708 |
| semantic/windows/k=5   |         0.778 |       0.537 |           0.542 |       0.958 |     0.757 |         0.750 |
| hybrid/windows/k=3     |         0.694 |       0.567 |           0.528 |       0.833 |     0.715 |         0.708 |
| **hybrid/windows/k=5** |     **0.792** |   **0.589** |           0.556 |       0.958 |     0.742 |         0.750 |

No result exceeded the 8 KiB cap.

The rule selects **hybrid/windows/k=5**. Four configurations are within 1/24 of
its rewrite hit@k (semantic/windows/k=5, semantic/sections/k=5 and
lexical/sections/k=5, the last exactly at the boundary); hybrid/windows/k=5
wins the tie on MRR.

Rewrite hit@k difference from the selected configuration:

| Configuration         | Difference |           95% CI |
| --------------------- | ---------: | ---------------: |
| semantic/windows/k=5  |     −0.014 | [−0.083, +0.056] |
| semantic/sections/k=5 |     −0.028 | [−0.111, +0.069] |
| lexical/sections/k=5  |     −0.042 | [−0.125, +0.042] |
| hybrid/sections/k=5   |     −0.056 | [−0.139, +0.028] |
| lexical/windows/k=5   |     −0.069 | [−0.153, +0.014] |
| semantic/windows/k=3  |     −0.083 | [−0.153, −0.014] |
| hybrid/windows/k=3    |     −0.097 | [−0.167, −0.042] |
| hybrid/sections/k=3   |     −0.111 | [−0.208, −0.028] |
| semantic/sections/k=3 |     −0.125 | [−0.236, −0.014] |
| lexical/windows/k=3   |     −0.153 | [−0.264, −0.056] |
| lexical/sections/k=3  |     −0.167 | [−0.292, −0.056] |

Every k=3 configuration is measurably worse than the selected one. No other k=5
configuration is: their intervals include zero.

Per probe, K10, K11, K12 and K20 were found by all three rewrites in every k=5
configuration. K19 was the hardest probe: at most one rewrite hit in any
configuration, and it is the only as-is probe the selected configuration
missed. The selected configuration missed all three K07 rewrites, but returned
neighbouring pages (the EndpointSlice API reference, the glossary and
`service.md`) rather than the cited `debug-service.md` and
`endpoint-slices.md`.

## Conclusions

- With rewritten queries, the selected configuration returns a cited file in
  the top 5 for 79% of queries and a primary source for 56%.
- k=5 is better than k=3 in every mode and chunking, and five 1.5 KiB chunks
  fit the 8 KiB cap.
- The probe set cannot distinguish the modes or chunkings at k=5. Hybrid/windows
  was selected by the pre-registered tie-break, not by a demonstrated advantage.
  Hybrid ranking gave the highest rewrite MRR at both k, which is weak evidence
  that fusion improves ordering.
- The full probe text beats its short rewrites by 0.08–0.22 hit@k at k=5, so the quality of the agent's own queries limits retrieval at
  least as much as the configuration does.

## Limitations

- 24 probes and 72 rewrites from a single generation: differences below about
  0.1 hit@k cannot be resolved.
- Relevance is file-level and limited to cited files, so on-topic neighbouring
  pages count as misses.
- One embedding model, one chunk size and one rewrite model were tested.
- Query K03-3 contains a stray backtick left by the generation output parser;
  queries were frozen without manual edits.
- Retrieval quality is measured without the agent; whether retrieved context
  improves repairs is measured by the benchmark.

## Evidence

- [queries.json](queries.json) — frozen query set with generation settings and raw model responses.
- [metadata.json](metadata.json) — run provenance: repository revision, query and index hashes, corpus revision and embedding artifact.
- [summary.json](summary.json) — per-configuration and per-probe scores, the selection and bootstrap intervals.
- [raw.jsonl](raw.jsonl) — top 5 results of every query in every mode and chunking.

Reproduce with `make corpus-download download-models retrieval-index retrieval-evaluate`.
Index builds are deterministic: rebuilding the sections index reproduced the
same SHA-256.
