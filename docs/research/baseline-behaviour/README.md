# Baseline incident evaluation

Descriptive evaluation of the baseline condition on five Kubernetes incidents.
Each incident was run ten times, giving 50 independent attempts in total. The
purpose was to measure repair effectiveness, partial credit, outcome stability,
execution time and tool use before adaptation.

## Method

The evaluated model was Gemma 4 E4B instruction-tuned QAT Q4_0. Every attempt
used a fresh disposable Kind cluster, the same model-visible task, the same
single Bash tool, and the same runner limits. No prompt, model or scenario
tuning was performed during measurement.

Partial score is the mean of the two equally weighted grading criteria in each
scenario. Full success requires every criterion to pass. Stability is reported
descriptively through the score distribution and full-success rate; the sample
does not support claims about general Kubernetes performance.

| Parameter         | Value                                                              |
| ----------------- | ------------------------------------------------------------------ |
| Attempts          | 10 per incident; 50 total                                          |
| Cluster           | Kind, one control-plane node, Kubernetes v1.37.0                   |
| Model repository  | `google/gemma-4-E4B-it-qat-q4_0-gguf`                              |
| Model revision    | `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`                         |
| Model file        | `gemma-4-E4B_q4_0-it.gguf`                                         |
| Quantization      | Q4_0                                                               |
| Model SHA-256     | `676c35070db6dbe52f93e9c864ee0fba4eddea94b9c875d9cb10daff453fbaee` |
| Inference runtime | llama.cpp Vulkan, `b10524`                                         |
| Loop limits       | 25 turns, 50 tool calls, 60 s per tool call, 300 s per attempt     |

## Results

| Incident                    | Score distribution       | Mean score |    Full success | Completion |
| --------------------------- | ------------------------ | ---------: | --------------: | ---------: |
| `container-crash-loop`      | 0: 9, 1: 1               |      0.100 |      1/10 (10%) |      10/10 |
| `image-pull-failure`        | 1: 10                    |      1.000 |    10/10 (100%) |      10/10 |
| `missing-rbac-binding`      | 0: 8, 0.5: 2             |      0.100 |       0/10 (0%) |      10/10 |
| `service-selector-mismatch` | 1: 10                    |      1.000 |    10/10 (100%) |      10/10 |
| `unschedulable-cpu-request` | 0: 3, 1: 7               |      0.700 |      7/10 (70%) |      10/10 |
| **All incidents**           | **0: 20, 0.5: 2, 1: 28** |  **0.580** | **28/50 (56%)** |  **50/50** |

| Incident                    | Mean loop (s) | Median loop (s) | Mean turns | Mean tool calls |  Mean tokens |
| --------------------------- | ------------: | --------------: | ---------: | --------------: | -----------: |
| `container-crash-loop`      |         138.2 |           137.3 |        7.9 |             6.9 |     16,183.2 |
| `image-pull-failure`        |         110.0 |            97.3 |        7.0 |             6.0 |     11,499.9 |
| `missing-rbac-binding`      |         163.3 |           167.8 |        7.6 |             6.6 |     20,669.5 |
| `service-selector-mismatch` |          96.6 |           107.9 |        5.9 |             4.9 |      6,456.4 |
| `unschedulable-cpu-request` |         139.4 |           136.8 |        7.0 |             6.0 |     10,566.6 |
| **All incidents**           |     **129.5** |       **128.7** |    **7.1** |         **6.1** | **13,075.1** |

The two criteria for `image-pull-failure` and `service-selector-mismatch`
passed in every attempt. `least-privilege-access` did not pass in any
`missing-rbac-binding` attempt. The two criteria for `container-crash-loop`
passed together in only one attempt. `unschedulable-cpu-request` passed both
criteria in seven attempts.

Loop duration covers model responses and tool execution, including bounded
Kubernetes polling. It is not an inference-speed or repair-time measurement.
Token totals are the recorded prompt-plus-completion totals for model
responses, including repeated context.

## Interpretation

The baseline repaired 28 of 50 attempts completely. Its performance was
incident-dependent rather than uniformly weak or strong: image-pull and
service-selector faults were solved consistently, the CPU-scheduling fault was
solved in most attempts, and the crash-loop and RBAC faults were usually not
solved.

The dominant failure in the crash-loop scenario was leaving the Deployment
with an unhealthy updated ReplicaSet, so both readiness and replica-capacity
criteria failed. In the RBAC scenario, agents often restarted or replaced
workloads without establishing the required least-privilege access; the
deployment criterion passed in only two attempts and the access criterion in
none. CPU-scheduling failures similarly left the rollout incomplete or the
replica capacity below the required state.

These results are descriptive evidence from five incidents and ten repetitions
per incident. They do not establish a model ranking, causal effects of model
size, or general performance outside these scenarios.

## Raw data

The [raw directory](raw/) contains the single merged dataset: run metadata and
one JSON record for each of the 50 logical attempts. It preserves the complete
model transcripts, tool calls, timings, token usage, termination state and
criterion-level grading evidence.
