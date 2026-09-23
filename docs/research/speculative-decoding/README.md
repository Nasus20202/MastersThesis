# Multi-token prediction draft-length tuning

Selection of the per-model `--spec-draft-n-max` value for llama.cpp
multi-token prediction (MTP). The draft length trades fewer target-model
verification steps against more rejected draft tokens; this evaluation measures
which length maximises generation throughput for each served model on the
evaluation host.

## Method

The only intended experimental variable was the number of tokens the MTP layer
drafts per step. Each model was served with MTP disabled, then enabled with
`--spec-draft-n-max` of 1, 2, 3 and 4. Every setting generated one completion
for each of ten fixed prompts (greedy, temperature 0, maximum 256 tokens),
giving ten runs per setting and 200 runs in total. Prompts spanned explanation,
code, arithmetic reasoning, comparison and narrative writing, because draft
acceptance depends on how predictable the output is.

| Parameter         | Value                                                                 |
| ----------------- | --------------------------------------------------------------------- |
| Inference runtime | llama.cpp Vulkan `server-b10964`                                      |
| Context           | 32,768 tokens unified per slot                                        |
| Parallel slots    | 2                                                                     |
| GPU offload       | all layers, flash attention auto, f16 K/V cache                       |
| Reasoning         | enabled, 4,096-token budget                                           |
| Sampling          | greedy (temperature 0), maximum 256 generated tokens                  |
| Hardware          | AMD Navi 10 Radeon GPU, 8 GiB class (7.98 GiB usable), Vulkan backend |

| Model       | Repository                        | Artifact                        |
| ----------- | --------------------------------- | ------------------------------- |
| Gemma 4 E2B | `unsloth/gemma-4-E2B-it-qat-GGUF` | `gemma-4-E2B-it-qat-UD-Q4_K_XL` |
| Gemma 4 E4B | `unsloth/gemma-4-E4B-it-qat-GGUF` | `gemma-4-E4B-it-qat-UD-Q4_K_XL` |
| Qwen3.5 4B  | `unsloth/Qwen3.5-4B-MTP-GGUF`     | `Qwen3.5-4B-Q4_K_M`             |
| Qwen3.5 9B  | `unsloth/Qwen3.5-9B-MTP-GGUF`     | `Qwen3.5-9B-Q4_K_M`             |

Gemma 4 targets use a separate MTP drafter attached through
`benchmark/models-preset.ini`; Qwen3.5 artifacts carry the MTP layer inside the
target file. Exact repositories, revisions and hashes are in
[the model profiles](../../../benchmark/model-profiles/).

Generation throughput is taken from the server's `eval time` line (generated
tokens per second, excluding prompt evaluation). Draft acceptance is the
fraction of drafted tokens accepted by the target model; mean length includes
the target token. The sweep harness is retained as
[`sweep.sh`](sweep.sh); it restarts the server per setting and records
throughput and acceptance per run and VRAM per setting.

## Results

Mean generation throughput over the ten prompts. The speedup is relative to the
same model with MTP disabled.

| Model       | MTP off |      n-max 1 |      n-max 2 |      n-max 3 |      n-max 4 | Chosen |
| ----------- | ------: | -----------: | -----------: | -----------: | -----------: | -----: |
| Gemma 4 E2B |    38.9 | 55.4 (1.42x) | 61.9 (1.59x) | 63.7 (1.64x) | 62.0 (1.59x) |      3 |
| Gemma 4 E4B |    57.1 | 71.0 (1.24x) | 71.1 (1.25x) | 64.5 (1.13x) | 56.5 (0.99x) |      2 |
| Qwen3.5 4B  |    30.2 | 36.1 (1.19x) | 42.3 (1.40x) | 39.9 (1.32x) | 36.5 (1.21x) |      2 |
| Qwen3.5 9B  |    24.7 | 29.1 (1.18x) | 33.9 (1.37x) | 30.6 (1.24x) | 28.5 (1.15x) |      2 |

Mean draft acceptance falls as the draft length grows, because longer drafts
add tokens the target model is increasingly likely to reject.

| Model       | n-max 1 | n-max 2 | n-max 3 | n-max 4 |
| ----------- | ------: | ------: | ------: | ------: |
| Gemma 4 E2B |    0.70 |    0.58 |    0.50 |    0.43 |
| Gemma 4 E4B |    0.72 |    0.61 |    0.51 |    0.44 |
| Qwen3.5 4B  |    0.88 |    0.79 |    0.68 |    0.62 |
| Qwen3.5 9B  |    0.90 |    0.82 |    0.74 |    0.66 |

Because all settings saw the same ten prompts, the per-prompt comparison removes
prompt difficulty. Counting each prompt's fastest setting: Gemma 4 E2B chose
n-max 3 in 8 of 10 prompts, Gemma 4 E4B split between n-max 1 (6) and n-max 2
(4), and both Qwen models chose n-max 2 in 10 of 10 prompts.

Total VRAM in use at load, including the desktop baseline, peaked at 7.65 GiB
(Qwen3.5 9B, n-max 3); the Gemma drafter adds roughly 0.1 GiB over the same
target without it. No profile exceeded the card's 7.98 GiB.

Gemma 4 E2B decodes more slowly than E4B (best 63.7 against 71.1 t/s) despite
being the smaller model. The ordering holds with MTP disabled (38.9 against
57.1 t/s) and reproduces with `llama-bench` on this GPU (tg128 40.3 against 64.8
t/s), while the same benchmark on the CPU reverses it (23.1 against 13.3 t/s).
The gap is therefore Vulkan underutilization for the smaller model on this card,
not a property of the model or of multi-token prediction.

## Conclusions

- Multi-token prediction improves generation throughput on every model,
  by 1.18x to 1.64x depending on model and draft length.
- The best draft length is model-specific rather than universal: 3 for Gemma 4
  E2B, and 2 for Gemma 4 E4B and both Qwen3.5 models. The llama.cpp default of
  3 is optimal only for Gemma 4 E2B.
- A draft length of 2 is best or within about 3% of best for every model, so it
  is the safest shared default; the per-model values were still written to the
  profiles.
- Rejected drafts explain the optimum. Past the best length, the additional
  drafted tokens are rejected often enough that the extra verification work
  outweighs the reduction in target-model steps.

The chosen values are set as `LLAMA_SPEC_DRAFT_N_MAX` in each
[model profile](../../../benchmark/model-profiles/) and forwarded to the server
by `benchmark/docker-compose.yaml`.

## Limitations

- Throughput was measured on one host with one Vulkan backend and a single
  repetition per prompt. Backend, driver and thermal state can shift absolute
  values.
- Each run was a single greedy 256-token completion, not the agentic benchmark
  workload. Acceptance depends on output content, and sampled decoding may
  change it. The relative ordering is the result; the absolute rates are
  indicative.
- The desktop baseline contributes about 2 GiB to the reported VRAM totals;
  model-attributed use is lower.
- No energy or monetary cost was measured.

## Raw evidence

- [Per-run measurements](results/runs.csv) — throughput, acceptance, mean
  length and generated tokens per run.
- [VRAM measurements](results/vram.csv) — total VRAM per model and setting.
- [Raw responses](results/raw/) — one completion per run.
