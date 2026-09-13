# Baseline incident evaluation

Baseline evaluation of Gemma 4 E4B on five Kubernetes troubleshooting incidents.
Each incident was repeated ten times, giving 50 attempts before adaptation.

## Method

Each attempt used a fresh disposable Kind cluster, the same model-visible task,
the same Bash tool and the same runner limits. No prompt, model or scenario
changes were made during measurement.

Partial score is the mean of two equally weighted deterministic criteria. Full
success requires both criteria to pass. All 50 attempts completed normally.

| Parameter | Value |
| --- | --- |
| Attempts | 10 per incident; 50 total |
| Cluster | Kind, one control-plane node, Kubernetes v1.37.0 |
| Model | Gemma 4 E4B instruction-tuned QAT Q4_0 |
| Model repository | `google/gemma-4-E4B-it-qat-q4_0-gguf` |
| Model revision | `4b4a2c1d584be7264f87aac328a1bc739ce81b6c` |
| Model SHA-256 | `676c35070db6dbe52f93e9c864ee0fba4eddea94b9c875d9cb10daff453fbaee` |
| Runtime | llama.cpp Vulkan, `b10524` |
| Context | 32,768 tokens |
| Reasoning | enabled, budget 4,096 |
| Parallelism | benchmark 4; llama.cpp 2 |
| Generation controls | temperature and max output not overridden; llama.cpp defaults |
| Loop limits | 25 turns, 50 tool calls, 60 s per tool call, 300 s per attempt |

## Results

| Incident | Score distribution | Mean score | Full success |
| --- | --- | ---: | ---: |
| `container-crash-loop` | 0: 9, 1: 1 | 0.100 | 1/10 (10%) |
| `image-pull-failure` | 1: 10 | 1.000 | 10/10 (100%) |
| `missing-rbac-binding` | 0: 8, 0.5: 2 | 0.100 | 0/10 (0%) |
| `service-selector-mismatch` | 1: 10 | 1.000 | 10/10 (100%) |
| `unschedulable-cpu-request` | 0: 3, 1: 7 | 0.700 | 7/10 (70%) |
| **All incidents** | **0: 20, 0.5: 2, 1: 28** | **0.580** | **28/50 (56%)** |

| Incident | Mean loop (s) | Median loop (s) | Mean turns | Mean tool calls | Mean tokens |
| --- | ---: | ---: | ---: | ---: | ---: |
| `container-crash-loop` | 138.2 | 137.3 | 7.9 | 6.9 | 16,183.2 |
| `image-pull-failure` | 110.0 | 97.3 | 7.0 | 6.0 | 11,499.9 |
| `missing-rbac-binding` | 163.3 | 167.8 | 7.6 | 6.6 | 20,669.5 |
| `service-selector-mismatch` | 96.6 | 107.9 | 5.9 | 4.9 | 6,456.4 |
| `unschedulable-cpu-request` | 139.4 | 136.8 | 7.0 | 6.0 | 10,566.6 |
| **All incidents** | **129.5** | **128.7** | **7.1** | **6.1** | **13,075.1** |

Loop duration includes model responses, tool execution and Kubernetes polling.
It is descriptive only and is not used for scoring. Token totals include repeated
conversation context.

## Findings

- Image-pull and service-selector faults were solved consistently. Both are
  direct configuration faults with a clear repair target.
- Crash-loop failures were mainly remediation failures. The model often found
  the invalid container arguments, but deleted or restarted Pods, or applied an
  incorrect Deployment patch. Only one attempt restored the persistent desired
  state correctly.
- The RBAC scenario was the clearest weakness. No attempt restored the required
  least-privilege access. One partial repair restored functionality but granted
  broader ConfigMap permissions than allowed.
- The CPU scenario was the most variable. Successful runs reduced or rolled
  back the bad CPU request; failed runs often changed replica count instead,
  leaving the faulty rollout unresolved.
- Unsuccessful traces often ended with the model assuming or declaring success
  without verifying the final state. This suggests that post-change validation
  is an important target for prompt and skill adaptations.
- Partial scoring provided little extra resolution in this pilot: 48 of 50
  attempts scored either 0 or 1. Expanded scenarios should use more independent
  criteria if partial improvement is important to measure.

The results show strong task dependence rather than uniformly good or poor
Kubernetes ability. They are descriptive evidence from five fixed incidents and
do not establish general Kubernetes performance or a model ranking.

## Raw data

The [raw directory](raw/) contains run metadata and one JSON record for each
attempt, including complete transcripts, tool calls, timings, token usage,
termination state and criterion-level grading evidence.
