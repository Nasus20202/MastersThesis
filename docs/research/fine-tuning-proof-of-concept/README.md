# Fine-tuning proof of concept

Before the fine-tuning runs of #26, this proof of concept checks that the
training code in [`finetune/`](../../../finetune/) works on Modal for both
model families, measures what one GPU type costs per training token, and
checks that a trained adapter carries over to the served 4-bit model (#24).
It tests the GPU choice of D-055 (one A100 40 GB per run). Raw evidence is in
[raw/](raw/).

## Summary

- **The pipeline runs end to end for both families.** Qwen3.5 4B and
  Gemma 4 E4B trained on whole trajectories up to 16,384 tokens in bfloat16
  LoRA, saved an adapter and logged the evaluation loss before and after, on
  both GPUs. No run ran out of memory; the longest step needed 15.2 GiB
  (Qwen) and 24.5 GiB (Gemma).
- **The A100 40 GB stays.** Training cost $0.45 per million tokens on the
  A100 against $0.47 on the L40S, the mean of the two models. The 5% gap is
  under the 15% margin fixed before the runs, so D-055 holds. The A100 is
  1.2–1.5× faster for Qwen and about as fast for Gemma.
- **The GPU does not change the result.** Memory and evaluation losses agree
  between the two GPUs to 0.001.
- **The adapter carries over to the served model.** Adapters trained on the
  bfloat16 checkpoint and converted to GGUF change the served 4-bit model in
  the same direction as the evaluation loss: Gemma's perplexity on the
  evaluation trajectories fell from 2.33 to 2.03; Qwen's rose from 8.91 to
  9.20, within about one standard error.
- **Thirty steps teach Gemma the format but not Qwen.** Gemma's evaluation
  loss fell from 1.21 to 0.83; Qwen's rose slightly, from 0.596 to 0.606. The
  training data are Qwen3.5 9B trajectories, which Qwen3.5 4B already
  predicts well. This is a smoke test, not a measure of how much fine-tuning
  helps.
- **The causal-conv1d kernel speeds Qwen up by 10–14%.** A fifth run, with
  Qwen3.5's convolution kernel from a prebuilt wheel, trained at 4,119
  instead of 3,609 tokens per second (steady steps), with the same memory and
  losses.
- **The POC cost about $1.60** for five runs of 7–8 minutes each.

## Question

1. Which Modal GPU, the A100 40 GB or the L40S 48 GB, trains our models at the
   lower cost per million training tokens?
2. Does the pipeline run end to end on Modal for both model families: whole
   trajectories (16,384 tokens) in bfloat16 LoRA, evaluation loss before and
   after, a saved adapter?
3. Does a POC adapter, trained on the checkpoint in bfloat16, lower the
   perplexity of the served 4-bit GGUF on the evaluation trajectories?

The plan fixed before the runs compared the A100 with the H100 80 GB. After the
A100 runs showed that the models use at most 25 GiB, the researcher replaced
the H100 with the L40S: the question became whether a cheaper GPU is enough,
not whether a faster one pays off.

## Method

**Runs.** Four: Qwen3.5 4B and Gemma 4 E4B, one model per family, each on an
A100 40 GB and an L40S, one Modal container and GPU per run
([`finetune.modal`](../../../finetune/src/finetune/modal.py)). Both families
are needed because they differ where speed is decided: Qwen3.5 has
linear-attention layers run by Triton kernels (flash-linear-attention),
Gemma 4 has full-attention layers, a 262K vocabulary and a per-layer
embedding table. E2B and 9B are expected to scale with their size. A fifth
run repeats Qwen3.5 4B on the A100 with Qwen3.5's convolution kernel
installed.

```sh
uv run modal run -m finetune.modal --manifest manifests/poc.yaml \
  --variants <qwen35-4b-lora|gemma-4-e4b-lora> --gpu <A100-40GB|L40S>
```

| Setting     | Value                                                                                                                        |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Checkpoints | `Qwen/Qwen3.5-4B`, `google/gemma-4-E4B-it-qat-q4_0-unquantized`, revisions in [`config.yaml`](../../../finetune/config.yaml) |
| Training    | 16-bit LoRA, bfloat16 weights, rank 16, alpha 16, no dropout, attention and MLP projections                                  |
| Optimiser   | AdamW, learning rate 1e-4, 3 warm-up steps, one trajectory per step, 30 steps, seed 0                                        |
| Memory      | gradient checkpointing, loss computed in chunks                                                                              |
| Software    | Python 3.14.10, torch 2.14.1 (CUDA 13.0), transformers 5.17.0, peft 0.21.3, flash-linear-attention 0.6.0, triton 3.8.0       |

These training values only make the runs comparable; they are not proposals
for #26.

**Data.** Clean attempts (full score, ended by the agent) of the Qwen3.5 9B
prompt agent in the [October development run](../kaggle-development-suite/README.md),
listed in [`finetune/manifests/poc.yaml`](../../../finetune/manifests/poc.yaml).
Training: of the attempts outside the 8 validation scenarios, sorted by
scenario, every third one: 10 trajectories of 4.7K to 16.2K tokens (Qwen
tokenizer). Evaluation: the 6 clean attempts on validation scenarios, 10.1K
to 15.3K tokens, never trained on. Each model trains on the same rendered
data on both GPUs; the hashes are in `session.json`.

**Measures** (`metrics.json` of each run):

- _Sweep_: peak GPU memory and tokens per second of a training step at
  4,096, 8,192, 12,288 and 16,384 tokens (three steps each, learning rate 0).
- _Training throughput_: training tokens per second over the 30 steps.
- _Evaluation loss_ on the evaluation trajectories, before and after
  training: mean cross-entropy over the assistant tokens.

**Cost per million training tokens** = price per hour / (training tokens per
second × 3,600) × 10⁶. The price per hour is Modal's GPU price plus 4 CPU
cores and 32 GiB of memory, $0.45 per hour (Modal's price list on
2026-11-20: CPU $0.0000131 per core-second, memory $0.00000222 per
GiB-second). Runs 3–4 requested 2 cores and 16 GiB; the same host price is
used for every run so that only the GPU differs.

| GPU        | GPU $/h | Host $/h | Total $/h |
| ---------- | ------: | -------: | --------: |
| A100 40 GB |    2.10 |     0.45 |      2.55 |
| L40S 48 GB |    1.95 |     0.45 |      2.40 |

**Decision rule** (fixed before the runs):

1. A GPU on which either model fails to train 16,384-token trajectories is out.
2. Of the rest, the one with the lower mean cost per million training tokens
   over the two models wins.
3. If they differ by less than 15%, the A100 40 GB stays (D-055): one run per
   GPU and model cannot show a smaller difference.

**Served model.** Each adapter of the A100 runs is converted to GGUF with
llama.cpp and applied with `--lora` to the served 4-bit file
([`serving/llama.py`](../../../finetune/src/finetune/serving/llama.py)), in
the llama.cpp image pinned in `config.yaml` (CPU build), on the local
machine. Measure: perplexity on the six evaluation trajectories in
8,192-token windows, with and without the adapter. llama.cpp scores every
token in the second half of each window, tool results included, so the value
shows the direction of the effect, not its size. The GPU does not change the
adapter (see the losses below), so one adapter per model is checked.

## Results

### Cost per training token

Training throughput over the 30 steps and the cost it implies.

| Model       | GPU  | Training tokens/s | Steady tokens/s | $/M tokens |
| ----------- | ---- | ----------------: | --------------: | ---------: |
| Qwen3.5 4B  | A100 |             2,161 |           3,609 |       0.33 |
| Qwen3.5 4B  | L40S |             1,752 |           2,597 |       0.38 |
| Gemma 4 E4B | A100 |             1,260 |           1,260 |       0.56 |
| Gemma 4 E4B | L40S |             1,192 |           1,192 |       0.56 |
| **Mean**    | A100 |                   |                 |   **0.45** |
| **Mean**    | L40S |                   |                 |   **0.47** |

$/M tokens uses the training tokens/s column (the rule's measure). Steady
tokens/s leaves out the three Qwen steps (1, 6 and 9) that took 13–26 s
because Triton compiled kernels for a new sequence length; Gemma has no such
steps. Both GPUs trained 16,384 tokens, so neither is out (rule 1). The
L40S costs 5.5% more per token (rule 2), under the 15% margin, so the A100
stays (rule 3).

The difference between the GPUs depends on the family. Qwen's linear
attention keeps its speed nearly flat with length, and the A100 is 1.2× to
1.5× faster in the sweep. Gemma's full attention grows with the square of the
length; its speed halves from 4K to 16K tokens on both GPUs, and the two are
within 7% of each other.

| Model       | GPU  | 4,096 | 8,192 | 12,288 | 16,384 |
| ----------- | ---- | ----: | ----: | -----: | -----: |
| Qwen3.5 4B  | A100 | 3,861 | 3,804 |  3,751 |  3,671 |
| Qwen3.5 4B  | L40S | 3,219 | 2,706 |  2,570 |  2,486 |
| Gemma 4 E4B | A100 | 1,804 | 1,434 |  1,189 |    997 |
| Gemma 4 E4B | L40S | 1,875 | 1,355 |  1,116 |    958 |

Sweep: training tokens per second at each sequence length, mean of three steps.

### End to end

| Model       | GPU  | Load s | Weights GiB | Peak GiB 4K → 16K | Training peak GiB | Host peak GiB | Eval loss before → after |
| ----------- | ---- | -----: | ----------: | ----------------- | ----------------: | ------------: | ------------------------ |
| Qwen3.5 4B  | A100 |     52 |         7.9 | 9.9 → 15.2        |              15.1 |          12.5 | 0.5956 → 0.6055          |
| Qwen3.5 4B  | L40S |     10 |         7.9 | 9.9 → 15.2        |              15.1 |          12.3 | 0.5956 → 0.6060          |
| Gemma 4 E4B | A100 |     78 |        14.9 | 17.4 → 24.5       |              24.3 |          18.8 | 1.2119 → 0.8332          |
| Gemma 4 E4B | L40S |     16 |        14.9 | 17.4 → 24.5       |              24.3 |          18.4 | 1.2118 → 0.8327          |

GPU memory is PyTorch's peak allocation; host memory is the process's peak
resident memory. Load time includes the checkpoint download on the A100 runs;
the L40S runs read it from the Modal volume the first runs filled.

All four runs kept every trajectory (10 training, 6 evaluation) under the
16,384-token cap, trained 30 finite steps and saved an adapter
(21.2M trainable parameters for Qwen, 34.9M for Gemma). Memory grows
linearly with length, about 0.45 GiB (Qwen) and 0.6 GiB (Gemma) per 1K
tokens, so 16K tokens use less than two thirds of either GPU. The longest
trajectories of #26 could be up to 32K tokens; at this rate Gemma would need
about 34 GiB, which still fits a 40 GB A100. That is an extrapolation, not a
measurement.

Gemma's evaluation loss fell by a third; Qwen's rose by 0.01. The training
trajectories come from Qwen3.5 9B, whose format and wording Qwen3.5 4B
already predicts (loss 0.60 against Gemma's 1.21), so 30 steps at this
learning rate have little to teach it.

### Convolution kernel

Run 5 repeats run 1 (Qwen3.5 4B on the A100) with the `causal-conv1d` CUDA
kernel; everything else is the same. PyPI has only the kernel's source, and
compiling it in Modal's image builder failed with the log cut off before the
error, so run 5 installs a prebuilt wheel for torch 2.14.1 and CUDA 13.0 from
Astral's GPU index
([build-causal-conv1d](https://github.com/astral-sh-build/build-causal-conv1d)). Runs 1–4
used transformers' PyTorch fallback for Qwen3.5's short convolution, and
their logs warn about it; run 5's log does not.

| Measure                     | Fallback (run 1) |  Kernel (run 5) | Change |
| --------------------------- | ---------------: | --------------: | -----: |
| Training tokens/s, 30 steps |            2,161 |           2,380 |   +10% |
| Steady tokens/s             |            3,609 |           4,119 |   +14% |
| Sweep tokens/s at 16,384    |            3,671 |           4,105 |   +12% |
| Peak GiB at 16,384          |             15.2 |            14.9 |    −2% |
| Eval loss before → after    |  0.5956 → 0.6055 | 0.5956 → 0.6056 |        |
| $/M tokens                  |             0.33 |            0.30 |    −9% |

The gain is 11–12% at every sweep length (4,332, 4,272, 4,211 and 4,105
tokens/s from 4K to 16K tokens). The L40S was not rerun with the kernel. For
the GPU choice to change, the kernel would have to make the L40S 15% cheaper
than the A100; a gain of this size on both GPUs cannot.

### Served model

| Model       | Adapter      | Perplexity without | Perplexity with |
| ----------- | ------------ | -----------------: | --------------: |
| Gemma 4 E4B | run 2 (A100) |      2.332 ± 0.025 |   2.033 ± 0.019 |
| Qwen3.5 4B  | run 1 (A100) |      8.908 ± 0.217 |   9.203 ± 0.230 |

Perplexity of the served 4-bit GGUF (`gemma-4-E4B-it-qat-UD-Q4_K_XL`,
`Qwen3.5-4B-Q4_K_M`) on the six evaluation trajectories, 8,192-token
windows; ± is llama.cpp's standard error.

Both adapters converted without errors and loaded into the served model.
The served perplexity moves the way the evaluation loss on the bfloat16
checkpoint moved: down by 13% for Gemma, well beyond the standard error, and
up by 3% for Qwen, within it. What training
changes in the checkpoint therefore reaches the 4-bit model. The values are
not comparable between the families: the tokenizers and chat templates
differ, and the windows include tool results.

## Conclusions

- The training code is ready for #26 on Modal: both families, whole
  trajectories, evaluation loss and adapter, with no change between GPUs.
- The A100 40 GB of D-055 is confirmed by the rule fixed before the runs. No
  change of decision is proposed.
- The served path works: an adapter trained on the bfloat16 checkpoint
  changes the served 4-bit model in the expected direction, so the
  fine-tuned condition can use the benchmark's existing llama.cpp serving.
- Proposal for the runs of #26: request 2 CPU cores and 8 GiB of host memory
  per run instead of 4 and 32 GiB. One core was busy; host memory peaked at
  18.8 GiB (Gemma), and Modal bills the larger of request and use, so the
  request is not a limit.

## Limitations

- One run per GPU and model; differences of a few percent are noise, which
  the 15% margin of the rule reflects.
- The H100 was not measured; the result says only that the L40S is not
  cheaper than the A100.
- 30 steps on 10 trajectories measure speed and memory, not how much
  fine-tuning helps; no agent run uses these adapters.
- Runs 1–4 ran Qwen's short convolution in PyTorch's fallback; the kernel
  was measured in one run, on the A100 only. The fallback is the same on
  both GPUs, so the GPU comparison holds; with the kernel, Qwen trains about
  10% faster than the cost table shows.
- Modal picks the host CPU; it is not recorded.
- Served perplexity covers the two POC models only; Gemma 4 E2B and
  Qwen3.5 9B share their family's template and conversion path. Perplexity
  counts tool results as well as the model's own tokens.

## Evidence

| Run                   | Model              | GPU       | Commit     | Raw                                                  |
| --------------------- | ------------------ | --------- | ---------- | ---------------------------------------------------- |
| `run-20261120-184233` | Qwen3.5 4B         | A100      | `23781a2a` | [raw/run-20261120-184233/](raw/run-20261120-184233/) |
| `run-20261120-190435` | Gemma 4 E4B        | A100      | `23781a2a` | [raw/run-20261120-190435/](raw/run-20261120-190435/) |
| `run-20261120-191247` | both               | L40S      | `0745571c` | [raw/run-20261120-191247/](raw/run-20261120-191247/) |
| `run-20261120-222907` | Qwen3.5 4B, kernel | A100      | `7bd7e559` | [raw/run-20261120-222907/](raw/run-20261120-222907/) |
| served perplexity     | both               | local CPU | —          | [raw/served/](raw/served/)                           |

Each run's commit is tagged `finetune-poc/<run>`. Each run directory holds
`session.json` (commit, manifest, data hashes) and
per variant `metrics.json` and `log.txt`. Adapters stay on the Modal volume
`finetune-results`.
