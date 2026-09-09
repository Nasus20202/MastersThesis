# Kubernetes knowledge check

Paired evaluation of Gemma 4 E4B's Kubernetes troubleshooting knowledge with and without fixed documentation context.

## Method

Twenty-four frozen probes were run once without documentation and once with audited, fixed excerpts. Prompts and D/E/R/C criteria were unchanged. Both conditions used Gemma 4 E4B, the pinned llama.cpp runtime, reasoning enabled with a 4096-token budget, temperature 0, seed 42, maximum output 6144, no live cluster, tools, or retrieval.

The context condition supplied 112 exact excerpts from `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15` across 20 official source files. Each excerpt was checked against its recorded path, blob SHA, revision and line range. Raw API responses and timing/token metadata are preserved in [raw.jsonl](raw.jsonl) and [context-raw.jsonl](context-raw.jsonl); prompts, criteria and excerpts are in [tasks.json](tasks.json) and [context-tasks.json](context-tasks.json).

`D/E/R/C` denotes Diagnosis, Evidence, Repair and Caveats. `1` means the frozen question-specific criterion was met. Scores are not combined into a single benchmark metric.

## Results

| Scenario |     Diagnosis |      Evidence |        Repair |       Caveats |
| -------- | ------------: | ------------: | ------------: | ------------: |
| Unaided  | 17/24 (70.8%) |  8/24 (33.3%) | 20/24 (83.3%) |  4/24 (16.7%) |
| Context  | 22/24 (91.7%) | 11/24 (45.8%) | 19/24 (79.2%) | 15/24 (62.5%) |

The percentages are per criterion across the 24 probes; they are not combined into a single score.

| Probe | Unaided | Context | Main result                               |
| ----- | ------: | ------: | ----------------------------------------- |
| K01   | 1/0/1/1 | 1/0/1/1 | usage evidence incomplete                 |
| K02   | 1/0/1/0 | 1/0/1/1 | propagation caveats improved              |
| K03   | 1/0/1/0 | 1/0/1/0 | unchanged                                 |
| K04   | 1/1/1/0 | 1/1/1/0 | rollout caveats remained incomplete       |
| K05   | 1/0/1/0 | 1/0/0/1 | rollout repair remained unsafe            |
| K06   | 1/1/1/0 | 1/1/1/0 | unchanged                                 |
| K07   | 0/0/1/0 | 1/0/1/0 | EndpointSlice diagnosis improved          |
| K08   | 1/1/1/0 | 1/1/1/1 | directionality caveat improved            |
| K09   | 0/1/1/0 | 0/1/1/0 | FQDN/search error remained                |
| K10   | 1/0/1/1 | 1/0/0/1 | delayed binding identified; repair unsafe |
| K11   | 0/0/1/0 | 1/1/1/1 | attachment/access evidence improved       |
| K12   | 1/1/1/1 | 1/1/1/1 | unchanged                                 |
| K13   | 1/0/1/0 | 1/0/1/1 | RBAC scope caveat improved                |
| K14   | 1/0/0/0 | 1/1/1/1 | subresource knowledge corrected           |
| K15   | 1/0/1/0 | 1/1/1/1 | binding/aggregation reasoning improved    |
| K16   | 1/1/1/0 | 1/1/1/1 | HPA denominator caveat improved           |
| K17   | 1/0/1/1 | 1/1/1/1 | layered evidence improved                 |
| K18   | 1/1/1/0 | 1/0/1/0 | node diagnosis remained incomplete        |
| K19   | 0/0/0/0 | 1/1/0/0 | terminating-endpoint diagnosis improved   |
| K20   | 0/0/0/0 | 0/0/0/0 | `pods/exec` verb remained wrong           |
| K21   | 0/0/1/0 | 1/0/1/1 | sidecar denominator diagnosis improved    |
| K22   | 0/0/1/0 | 1/0/1/1 | access-mode diagnosis improved            |
| K23   | 1/1/0/0 | 1/0/0/0 | OOM/eviction repair remained unsafe       |
| K24   | 1/0/1/0 | 1/0/1/1 | ExternalName caveats improved             |

The corrected excerpts fixed or strengthened diagnosis and caveat knowledge for EndpointSlices, ConfigMap propagation, RBAC scope and subresources, HPA denominators, volume attachment/access modes, terminating endpoints and ExternalName Services. Persistent failures included exact `pods/exec` authorization, FQDN versus search-path semantics, complete evidence collection, and safe repairs for storage, rollout and OOM/eviction cases. Documentation presence did not guarantee faithful use: several responses still proposed invalid fields, unsafe deletion or reversed rollout/resource changes.

## Conclusions

- Overall, the results are expected for a small 4B model on hard Kubernetes troubleshooting: the baseline is useful and reasonably broad, but its precision and operational safety are not strong enough for autonomous cluster operations.
- Provided documentation improved diagnosis from 17/24 (70.8%) to 22/24 (91.7%) and caveats from 4/24 (16.7%) to 15/24 (62.5%). Evidence improved less, from 8/24 (33.3%) to 11/24 (45.8%), while repair remained high but declined from 20/24 (83.3%) to 19/24 (79.2%).
- The context most clearly helped with EndpointSlices, ConfigMap propagation, RBAC scope and subresources, HPA denominators, volume access, terminating endpoints and ExternalName Services.
- Documentation did not ensure operational reliability: errors remained in `pods/exec` authorization, FQDN/search semantics, complete evidence collection and safe repairs. Some responses still proposed invalid fields, unsafe deletion or reversed rollout/resource changes.
- The results support the pinned Kubernetes 1.37 snapshot as the common documentation-assisted qualification corpus. Its identity is now frozen for experiments; this does not freeze the remaining benchmark protocol.
