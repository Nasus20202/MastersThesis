# Baseline model comparison pilot

Comparison of four local model artifacts on 10 attempts each of the baseline
`image-pull-failure` scenario (40 attempts total), to inform model selection.

## Method

The model artifact was the only intended experimental variable. Each model was
run sequentially for 10 attempts with a fresh scenario environment per attempt.
No attempt was manually repaired or discarded. The baseline prompt, generic
Bash-only tool, sandbox, three-node Kind topology, Kubernetes version, grading,
loop limits, context size and reasoning configuration were unchanged across
models.

The exact model-visible prompt in every attempt was:

```text
Task:
An application in the Kubernetes cluster is unhealthy. Restore it to a healthy state.

You are working in a Kubernetes troubleshooting environment. Use the bash tool to inspect and modify the environment as needed. Complete the task using only the available environment. When finished, provide a short final response.
```

| Parameter         | Value                                       |
| ----------------- | ------------------------------------------- |
| Cluster           | Three-node Kind cluster, Kubernetes v1.37.0 |
| Context           | 32,768 tokens                               |
| Reasoning         | Enabled, 4,096-token budget                 |
| Loop limits       | 12 turns, 24 tool calls, 300 seconds        |
| Hardware          | AMD Navi 10 Radeon GPU, 8 GB class VRAM     |
| Inference runtime | llama.cpp Vulkan, build b10524              |

Gemma artifacts use QAT Q4_0; Qwen artifacts use Q4_K_M. Model family, size and
quantization therefore vary together, so their individual effects cannot be
identified from this comparison.

Exact model sources, filenames, revisions and hashes are retained in the
[model profiles](../../../benchmark/model-profiles/) and [raw evidence](#raw-evidence).

## Results

The [grader](../../../benchmark/scenarios/image-pull-failure/scenario.yaml)
checks rollout completion and whether both desired and updated replica counts
equal three. The score is the mean of these two binary checks; full repair
requires both to pass. Grading runs after the agent stops and can wait up to
60 seconds for rollout completion. It does not establish the time of repair.

All scores were either 0 or 1. Full repair and agent termination are reported
separately; reaching the turn limit is not normal completion, even when the
record contains no agent error.

| Model       | Full repair | Mean score | Normal completion | Turn limit | Timeout |
| ----------- | ----------: | ---------: | ----------------: | ---------: | ------: |
| Gemma 4 E2B |        6/10 |      0.600 |                 9 |          0 |       1 |
| Gemma 4 E4B |        9/10 |      0.900 |                10 |          0 |       0 |
| Qwen3.5 4B  |       10/10 |      1.000 |                 4 |          2 |       4 |
| Qwen3.5 9B  |       10/10 |      1.000 |                 7 |          0 |       3 |

Normal completion means `agent.termination = completed`. Of those attempts,
six E2B, nine E4B, four Qwen 4B and seven Qwen 9B attempts passed both checks.
Across both Gemma models, 15 of 20 attempts passed; all 20 Qwen attempts passed.

| Model       | Mean loop time (s) | Median loop time (s) | Mean tokens | Mean turns | Mean tool calls |
| ----------- | -----------------: | -------------------: | ----------: | ---------: | --------------: |
| Gemma 4 E2B |               73.4 |                 44.2 |    12,460.6 |        6.5 |             5.6 |
| Gemma 4 E4B |               53.7 |                 53.1 |    13,957.9 |        7.7 |             6.7 |
| Qwen3.5 4B  |              156.1 |                 72.9 |    22,524.9 |        9.9 |             9.7 |
| Qwen3.5 9B  |              142.6 |                 85.5 |    20,570.2 |        8.7 |             8.0 |

These summaries include all 10 attempts per model. Loop time includes inference
and tool execution, including approximately 300 seconds for each timeout;
it excludes environment setup and grading. Mean tokens is the mean per-attempt
sum of recorded response `usage.total_tokens` (prompt plus completion tokens).
Repeated prompt context is counted again in each response; token totals are
not a direct measure of GPU work or energy use.

## Qualitative review

Most attempts began with Kubernetes resource inspection. Qwen 4B attempt
[007](results/qwen35-4b/image-pull-failure/007.json) began with `pwd && ls -la`.
The main failure patterns were:

- E2B used unresolved placeholders in attempt
  [009](results/gemma-4-e2b/image-pull-failure/009.json). Attempt
  [010](results/gemma-4-e2b/image-pull-failure/010.json) issued an invalid
  `kubectl set image` command, then timed out waiting for rollout completion.
- E4B attempt [008](results/gemma-4-e4b/image-pull-failure/008.json) restarted
  the deployment without correcting its image and scored 0.
- All seven Qwen timeouts ended in an unbounded `kubectl get pods -w` or
  `--watch` command. The subsequent grader passed both checks in each case.
  Qwen 4B attempts [004](results/qwen35-4b/image-pull-failure/004.json) and
  [007](results/qwen35-4b/image-pull-failure/007.json) exhausted the turn limit
  while continuing to inspect pods; both also passed grading.

## Interpretation and limitations

Both Qwen models repaired 10/10 attempts. Repair scores therefore do not
distinguish 9B from 4B. Qwen 9B had more normal completions (7 versus 4) and
a lower mean loop time, but a higher median loop time. E4B had nine repairs,
no forced stops and the lowest mean loop time. E2B had the lowest median loop
time, but its single timeout raised its mean above E4B's.

Ten repetitions of one fault measure variation on that fault, not performance
across Kubernetes incidents. The one-repair difference between E4B and either
Qwen model is insufficient to establish a reliable advantage. Model artifacts
also differ in quantization, and loop durations include polling and waiting,
so they cannot be interpreted as inference speed alone. No energy or monetary
cost was measured.

The observed reason to prefer Qwen 9B over 4B would be fewer forced stops,
not higher repair effectiveness. This is a candidate for further evaluation,
not an approved model change. The current approved primary model remains
Gemma 4 E4B ([D-007](../../decision-log.md)).

## Raw evidence

JSON records retain model revisions and hashes, configured runtime settings,
prompts, transcripts, tool outputs, token usage, timings, termination and
grading outcomes:

- [Gemma 4 E2B raw evidence](results/gemma-4-e2b/)
- [Gemma 4 E4B raw evidence](results/gemma-4-e4b/)
- [Qwen3.5 4B raw evidence](results/qwen35-4b/)
- [Qwen3.5 9B raw evidence](results/qwen35-9b/)
