# Baseline model comparison pilot

Pilot comparison of Gemma 4 E2B, Gemma 4 E4B, Qwen3.5 4B and Qwen3.5 9B on
the baseline `image-pull-failure` benchmark scenario. The purpose was local
model selection before expanding the baseline, prompt, skill, RAG and
fine-tuning experiments.

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

| Parameter          | Value                                                      |
| ------------------ | ---------------------------------------------------------- |
| Scenario           | `image-pull-failure`                                       |
| Attempts per model | 10                                                         |
| Execution          | Sequential, `parallelism=1`; `LLAMA_PARALLEL=1`            |
| Cluster            | Kind, three nodes; `kindest/node:v1.37.0` pinned by digest |
| Kubernetes         | v1.37.0                                                    |
| Context            | 32768 tokens                                               |
| GPU offload        | `LLAMA_GPU_LAYERS=999`; Vulkan `/dev/dri/renderD128`       |
| KV caches          | `f16` / `f16`                                              |
| Reasoning          | Enabled; budget 4096 tokens                                |
| Loop limits        | 12 turns, 24 tool calls, 300-second timeout                |
| Target hardware    | Local AMD Navi 10 Radeon GPU, 8 GB class VRAM              |
| Server             | llama.cpp Vulkan `server-vulkan-b10524`                    |

The quantization was not independently controlled: Gemma uses the existing QAT
Q4_0 artifact, while both Qwen artifacts use Q4_K_M. This is a practical local
model-selection pilot, not a claim that quantization effects were isolated.

## Artifacts

| Model       | GGUF repository                                                                                     | GGUF file                  | Quantization |
| ----------- | --------------------------------------------------------------------------------------------------- | -------------------------- | ------------ |
| Gemma 4 E2B | [`google/gemma-4-E2B-it-qat-q4_0-gguf`](https://huggingface.co/google/gemma-4-E2B-it-qat-q4_0-gguf) | `gemma-4-E2B_q4_0-it.gguf` | Q4_0         |
| Gemma 4 E4B | [`google/gemma-4-E4B-it-qat-q4_0-gguf`](https://huggingface.co/google/gemma-4-E4B-it-qat-q4_0-gguf) | `gemma-4-E4B_q4_0-it.gguf` | Q4_0         |
| Qwen3.5 4B  | [`unsloth/Qwen3.5-4B-GGUF`](https://huggingface.co/unsloth/Qwen3.5-4B-GGUF)                         | `Qwen3.5-4B-Q4_K_M.gguf`   | Q4_K_M       |
| Qwen3.5 9B  | [`unsloth/Qwen3.5-9B-GGUF`](https://huggingface.co/unsloth/Qwen3.5-9B-GGUF)                         | `Qwen3.5-9B-Q4_K_M.gguf`   | Q4_K_M       |

The model profiles in [`benchmark/model-profiles`](../../../benchmark/model-profiles/)
contain the exact download metadata and runtime model selection. No model
files are committed. The Qwen source model repositories are `Qwen/Qwen3.5-4B`
and `Qwen/Qwen3.5-9B`.

## Results

The deterministic score is the mean of the two equally weighted repair
criteria. `Full success` is the verifier result. `Failed / partial / complete`
describes execution: a failed attempt has a recorded model-agent error, a
partial attempt completed execution without full success, and a complete
attempt completed execution with full success. This keeps model-loop failures
visible even when post-attempt grading found that the repair had already been
applied.

| Model       | Runs | Mean score | Median score | Score distribution | Full success | Failed / partial / complete | Mean loop time | Median loop time | Avg tokens | Avg output tok/s | Avg turns | Avg tool calls |
| ----------- | ---: | ---------: | -----------: | ------------------ | -----------: | --------------------------- | -------------: | ---------------: | ---------: | ---------------: | --------: | -------------: |
| Gemma 4 E2B |   10 |      0.600 |        1.000 | 0: 4, 1: 6         |   6/10 (60%) | 1 / 3 / 6                   |         73.4 s |           44.2 s |   12,460.6 |             77.1 |       6.5 |            5.6 |
| Gemma 4 E4B |   10 |      0.900 |        1.000 | 0: 1, 1: 9         |   9/10 (90%) | 0 / 1 / 9                   |         53.7 s |           53.1 s |   13,957.9 |             57.1 |       7.7 |            6.7 |
| Qwen3.5 4B  |   10 |      1.000 |        1.000 | 1: 10              | 10/10 (100%) | 4 / 0 / 6                   |        156.1 s |           72.9 s |   22,524.9 |             35.5 |       9.9 |            9.7 |
| Qwen3.5 9B  |   10 |      1.000 |        1.000 | 1: 10              | 10/10 (100%) | 3 / 0 / 7                   |        142.6 s |           85.5 s |   20,570.2 |             24.8 |       8.7 |            8.0 |

`Mean loop time` includes the fixed 300-second timeout for timed-out model
loops; the median is therefore a more representative view of ordinary runs.

Per-attempt scores:

```text
Gemma 4 E2B:  1.0, 0.0, 0.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 0.0
Gemma 4 E4B:  1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 0.0, 1.0, 1.0
Qwen3.5 4B:   1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0
Qwen3.5 9B:   1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0
```

Termination reasons were:

| Model       | Completed | Turn limit | Timeout |
| ----------- | --------: | ---------: | ------: |
| Gemma 4 E2B |         9 |          0 |       1 |
| Gemma 4 E4B |        10 |          0 |       0 |
| Qwen3.5 4B  |         4 |          2 |       4 |
| Qwen3.5 9B  |         7 |          0 |       3 |

The Qwen timeout attempts generally repaired the deployment before issuing a
long-running `kubectl get pods -w`/`--watch` command. Their deterministic
post-attempt scores were still 1.0, but the execution errors remain failures
of the complete agent run.

## Qualitative review

All 30 attempts began by discovering or inspecting Kubernetes resources, and
the transcripts consistently inspected pods and deployments. The models
identified the `ImagePullBackOff`/invalid `nginx:does-not-exist` image pattern.
The main observed behaviours were:

- Gemma 4 E2B completed nine loops and timed out once. Six attempts applied a
  valid image repair and passed both checks; three completed attempts failed to
  repair the deployment. The unsuccessful attempts included invalid or
  placeholder-style commands, and the timeout occurred after repeated
  troubleshooting commands.
- Gemma 4 E4B completed all ten loops. Nine attempts applied a valid image repair and
  passed both checks. One attempt only restarted the deployment without
  correcting the image and scored 0.
- Qwen3.5 4B repaired all ten scenarios, but often tried more commands than
  necessary, including invalid patch/replace or edit forms. Four attempts timed
  out while watching pods and two reached the turn limit after the repair.
- Qwen3.5 9B repaired all ten scenarios. It usually used `kubectl set image`,
  although some attempts deleted a ReplicaSet or pod and one used a rollout
  undo. Three attempts timed out while watching pods after the repair.
- No dominant irrelevant Linux/system-state investigation or hallucinated
  external tool/resource pattern was observed. Repeated polling/watch commands
  were the clearest tool-use weakness.

The deterministic grader independently confirmed the repaired deployment in
all Qwen attempts and nine Gemma attempts. It does not make the timeout cases
equivalent to cleanly completed executions.

## Pilot conclusion

Qwen3.5 4B and Qwen3.5 9B were more effective than both Gemma models on this
single scenario’s deterministic repair score, but both Qwen models showed
poorer termination discipline and higher token use. Gemma 4 E2B was not faster
in this run despite its smaller size and had the lowest score. Gemma 4 E4B was
substantially faster and cheaper than the Qwen models, completed every loop,
and failed only one repair.

For subsequent thesis experiments, this pilot supports selecting **Qwen3.5
9B** when repair effectiveness is the primary requirement and the additional
latency is acceptable. It ran at the controlled 32k context on the target
local hardware. Gemma 4 E4B remains the practical fallback when predictable
completion time and lower computational cost are more important. The result is
scenario-specific pilot evidence, not a general model ranking.

## Raw evidence

Each run retains the normal JSON evidence, including the model artifact and
runtime settings, exact task/prompt, transcript, Bash calls and outputs,
stdout/stderr/exit codes, token usage, llama.cpp timings, durations, turns,
tool-call counts, termination, deterministic grading and timeout/error
evidence. The raw results are retained with this research record at:

- [Gemma 4 E2B raw evidence](results/gemma-4-e2b/)
- [Gemma 4 E4B raw evidence](results/gemma-4-e4b/)
- [Qwen3.5 4B raw evidence](results/qwen35-4b/)
- [Qwen3.5 9B raw evidence](results/qwen35-9b/)
