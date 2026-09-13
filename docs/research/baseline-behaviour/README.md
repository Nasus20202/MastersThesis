# Baseline incident evaluation

Baseline evaluation of Gemma 4 E4B on five Kubernetes troubleshooting incidents.
Each incident was repeated ten times, giving 50 attempts before adaptation.

## Method

Each attempt used a fresh disposable Kind cluster, the same model-visible task,
the same Bash tool and the same runner limits. No prompt, model or scenario
changes were made during measurement.

Partial score is the mean of two equally weighted deterministic criteria. Full
success requires both criteria to pass. All 50 attempts completed normally.

| Parameter           | Value                                                          |
| ------------------- | -------------------------------------------------------------- |
| Attempts            | 10 per incident; 50 total                                      |
| Cluster             | Kind, one control-plane node, Kubernetes v1.37.0               |
| Model               | Gemma 4 E4B instruction-tuned QAT Q4_0                         |
| Model repository    | `google/gemma-4-E4B-it-qat-q4_0-gguf`                          |
| Model revision      | `4b4a2c1d584be7264f87aac328a1bc739ce81b6c`                     |
| Runtime             | llama.cpp Vulkan, `b10524`                                     |
| Context             | 32,768 tokens                                                  |
| Reasoning           | enabled, budget 4,096                                          |
| Parallelism         | benchmark 4; llama.cpp 2                                       |
| Generation controls | temperature and max output not overridden; llama.cpp defaults  |
| Loop limits         | 25 turns, 50 tool calls, 60 s per tool call, 300 s per attempt |

## Results

| Incident                    | Score distribution       | Mean score |    Full success |
| --------------------------- | ------------------------ | ---------: | --------------: |
| `container-crash-loop`      | 0: 9, 1: 1               |      0.100 |      1/10 (10%) |
| `image-pull-failure`        | 1: 10                    |      1.000 |    10/10 (100%) |
| `missing-rbac-binding`      | 0: 8, 0.5: 2             |      0.100 |       0/10 (0%) |
| `service-selector-mismatch` | 1: 10                    |      1.000 |    10/10 (100%) |
| `unschedulable-cpu-request` | 0: 3, 1: 7               |      0.700 |      7/10 (70%) |
| **All incidents**           | **0: 20, 0.5: 2, 1: 28** |  **0.580** | **28/50 (56%)** |

| Incident                    | Median loop (s) | Mean tool calls | Mean tokens |
| --------------------------- | --------------: | --------------: | ----------: |
| `container-crash-loop`      |           137.3 |             6.9 |      16,183 |
| `image-pull-failure`        |            97.3 |             6.0 |      11,500 |
| `missing-rbac-binding`      |           167.8 |             6.6 |      20,670 |
| `service-selector-mismatch` |           107.9 |             4.9 |       6,456 |
| `unschedulable-cpu-request` |           136.8 |             6.0 |      10,567 |
| **All incidents**           |       **128.7** |         **6.1** |  **13,075** |

Loop duration includes model responses, tool execution and Kubernetes polling
and is not used for scoring. Token totals include repeated conversation context.

## Findings

- Image-pull and service-selector faults were solved in all ten attempts.
- In the crash-loop scenario, the model often found the invalid container
  arguments but deleted or restarted Pods, or applied an incorrect Deployment
  patch. Only one attempt fixed the Deployment template correctly.
- RBAC failed in all ten attempts. One partial repair restored the Deployment
  but granted broader ConfigMap permissions than allowed.
- CPU repair succeeded in seven attempts. Successful runs reduced or rolled
  back the bad CPU request; failed runs often changed replica count instead.
- 21 of 22 unsuccessful attempts ended with a response claiming or implying
  that the problem was resolved. Prompt and skill conditions should therefore
  require final-state verification.
- Partial scoring added little resolution in this pilot: 48 of 50 attempts
  scored either 0 or 1. Expanded scenarios should use more independent criteria
  if partial improvement is important to measure.

These results describe five fixed incidents and should not be interpreted as
general Kubernetes performance.

## Raw data

The [raw directory](raw/) contains run metadata and one JSON record for each
attempt, including complete transcripts, tool calls, timings, token usage,
termination state and criterion-level grading evidence.
