# Fine-tuning technical check

## Summary

Before adapters are trained for the study (#26), this check tests whether the
approved fine-tuning protocol (D-046 to D-048, D-054) works end to end on
Kaggle's two T4 GPUs: from benchmark trajectories, through LoRA or QLoRA
training, to an adapter applied to the 4-bit GGUF file the benchmark serves.
It trains small throwaway adapters (10 trajectories, 30 steps). Their values
show that the path works and what it costs; they are not results about how
much fine-tuning helps, and no training setting is tuned on them.

- **Training data equals serving text.** Each checkpoint's chat template is
  the one llama.cpp uses when serving. For Qwen3.5 one rendered conversation
  is exactly the sequence of prompts the agent saw. For Gemma 4 it differs in
  a fifth of the turns by one turn-closing token; the adapters still help
  Gemma, so the simple rendering is kept.
- **All four models train whole trajectories on two T4s with QLoRA.** 16-bit
  LoRA fits whole trajectories for every model except Qwen3.5 9B, whose
  16-bit weights leave room only for half of them. QLoRA is not slower than
  LoRA. Training runs at roughly 170 to 270 tokens per second, about 1 to 1.7
  GPU-session hours per million training tokens.
- **Adapters trained on the checkpoint work on the served GGUF.** Every
  adapter converts to GGUF, loads next to the 4-bit file and, for the three
  models not trained on their own outputs, lowers the served model's
  perplexity on unseen scenarios. Training the GGUF itself is neither
  possible nor needed.
- **The serving stack accepts adapters.** The llama.cpp router, the
  multi-token-prediction drafter and the Kaggle inference server all run with
  an adapter loaded.
- **Precision.** The decision rule fixed before the runs selects QLoRA for
  Gemma 4 E4B and 16-bit LoRA for Qwen3.5 4B. The two precisions differ little
  on both models; the choice for #26 is open (see the PR).

## Question

1. Which checkpoint is trained for each served model?
2. Is the text a model is trained on the same text it sees when it is served?
3. Which of the four local models (D-037) can be trained on whole
   trajectories on two T4s, with 16-bit LoRA and with 4-bit QLoRA, and at
   what speed?
4. Does an adapter trained on the checkpoint load on the served 4-bit GGUF
   and improve it on unseen scenarios? Could the GGUF be trained directly?
5. Does the adapter work in the rest of the serving stack: the llama.cpp
   router, the multi-token-prediction (MTP) drafter and the Kaggle inference
   server?

## Method

### Checkpoints

Each served GGUF declares the Hugging Face model it was quantized from; that
model, at a pinned revision, is the training checkpoint
([`students.py`](../../../finetune/src/finetune/students.py)). All four are
public, and within a family the chat templates are byte-identical. The study
fine-tunes Gemma 4 E4B and Qwen3.5 4B (D-046); Gemma 4 E2B and Qwen3.5 9B are
trained here to show what fits.

| Model       | Training checkpoint                          | Revision   | Served file                                     |
| ----------- | -------------------------------------------- | ---------- | ----------------------------------------------- |
| Gemma 4 E2B | `google/gemma-4-E2B-it-qat-q4_0-unquantized` | `6befbaca` | `unsloth/gemma-4-E2B-it-qat-GGUF`, `UD-Q4_K_XL` |
| Gemma 4 E4B | `google/gemma-4-E4B-it-qat-q4_0-unquantized` | `476025a0` | `unsloth/gemma-4-E4B-it-qat-GGUF`, `UD-Q4_K_XL` |
| Qwen3.5 4B  | `Qwen/Qwen3.5-4B`                            | `851bf6e8` | `unsloth/Qwen3.5-4B-MTP-GGUF`, `Q4_K_M`         |
| Qwen3.5 9B  | `Qwen/Qwen3.5-9B`                            | `c2022362` | `unsloth/Qwen3.5-9B-MTP-GGUF`, `Q4_K_M`         |

The Gemma checkpoints are Google's quantization-aware-trained (QAT) weights
stored unquantized, the same weights the served QAT GGUF was made from.

### Rendering

A training example is one whole agent trajectory rendered as a single token
sequence with the checkpoint's chat template. When the agent runs, however,
llama.cpp builds a new prompt for every assistant turn. Training on one
sequence matches serving only if two things hold:

1. **Same template.** Each served prompt equals the Hugging Face render of the
   same messages. Checked for every assistant turn of three Qwen3.5 9B
   attempts: the agent's messages and tools were sent to the benchmark's
   llama.cpp server (`b10964`, `/apply-template`) and rendered with the
   checkpoint's template (reasoning on, as served).
2. **Append-only history.** Each turn's served prompt is a prefix of the full
   conversation's render, so that the full render contains every prompt the
   model saw. Checked on every turn of the 154 clean Qwen3.5 9B attempts of
   the October development run.

### Training

[`finetune.check`](../../../finetune/src/finetune/check.py) renders the data
and runs [`finetune.train`](../../../finetune/src/finetune/train.py) once per
model and precision, each in a fresh process on both T4s, from a clone of the
repository ([`kaggle/check/kernel.py`](../../../finetune/kaggle/check/kernel.py)).

**Stack.** Hugging Face transformers + PEFT, with bitsandbytes for 4-bit
weights (D-054). Versions are locked in
[`uv.lock`](../../../finetune/uv.lock): torch 2.14.1, transformers 5.17.0,
PEFT 0.21.3, bitsandbytes 0.50.2.

**Precision.** In both variants the base model is frozen and only the LoRA
matrices (float32) are trained; forward and backward run in float16 autocast
with dynamic loss scaling, because the T4 has no bfloat16.

- `lora` (16-bit LoRA): the frozen weights are float16.
- `qlora`: the frozen linear layers of the decoder are stored as 4-bit NF4
  with double quantization and dequantized to float16 for each matrix
  multiplication; this cuts their memory to about a quarter. The other frozen
  matrices (embeddings, output head) stay float16 and norms float32. PEFT's
  `prepare_model_for_kbit_training` casts all non-4-bit parameters to
  float32; the code casts the matrices back to float16 afterwards
  (`train.frozen_matrices_to`), since they are frozen and float32 only doubles
  their memory.

**Placement over two GPUs.** The decoder layers are split into two
contiguous blocks, one per GPU ([`placement.py`](../../../finetune/src/finetune/placement.py)).
The split balances the estimated bytes per GPU: weights plus the layer inputs
that gradient checkpointing keeps for the longest sequence. Embeddings and the
output head stay on GPU 0. This is plain model parallelism: the GPUs compute
one after the other, so the second GPU adds memory, not speed. Gemma 4 has a
per-layer embedding table (2.8B of E4B's 7.9B parameters) that is frozen and
only indexed; it stays in host memory and only the looked-up rows move to the
GPU.

**Memory saving.** Gradient checkpointing (activations recomputed in the
backward pass) and a loss computed in chunks of target positions, so that the
full-vocabulary logits of a 16K sequence are never held at once.

**Shared settings.** These are fixed for the check only and are not proposals
for #26: rank 16, alpha 16, no dropout; adapters on the language model's
attention and MLP projections (`q/k/v/o_proj`, `gate/up/down_proj`), not on
the vision or audio towers; AdamW, learning rate 1e-4 with 3 warm-up steps;
one trajectory per step; seed 0; 30 steps, i.e. three passes over the 10
training trajectories.

**Loss.** Mean cross-entropy over the target tokens: what the model generates
when serving (reasoning, tool calls, answers, turn ends). The system prompt,
tool definitions, task and tool results are context only
([`dataset.py`](../../../finetune/src/finetune/dataset.py)).

**Data.** Clean attempts (full score, ended by the agent, D-047) of the
Qwen3.5 9B prompt agent in the October development run. Training: of the
attempts outside the 8 validation scenarios (D-048), sorted by scenario,
every third one, 10 trajectories of 4.2K to 16.2K tokens. Evaluation: the 6
clean attempts on validation scenarios, 9.7K to 15.4K tokens, never trained
on. The rendered data is hashed in each session file.

**Measurements per variant.**

- _Memory sweep._ Three steps each at 4,096, 8,192, 12,288 and 16,384 tokens
  with learning rate 0 (the adapter does not change), recording peak GPU
  memory and tokens per second. Training then uses the longest length that
  fitted and skips longer trajectories rather than truncating them.
- _Training throughput_ over the 30 steps, and steps skipped by the loss
  scaler.
- _Evaluation loss._ Target-token loss on the evaluation trajectories, before
  and after training.

### Served model

The adapter is converted with llama.cpp's `convert_lora_to_gguf.py` (f16) and
applied with `--lora` to the served GGUF in `llama-perplexity`, both from the
pinned `b10964` image ([`llama.py`](../../../finetune/src/finetune/llama.py)),
on the local Vulkan GPU. Perplexity is measured on the six evaluation
trajectories with and without the adapter, in 8,192-token windows; it scores
every token in the second half of each window. Unlike the training loss it
includes tool results and other context, so it checks the direction of the
effect on the served model, not its size.

The rendered evaluation trajectories were also scored with the GGUF loaded
directly in transformers (`gguf_file=`), to see whether the 4-bit file could
be trained instead of the checkpoint.

### Rest of the serving stack

A Qwen3.5 4B LoRA adapter (same format, targets and rank as above) was loaded

- in the benchmark's llama.cpp router (`b10964`) through a preset section;
- with MTP drafting (draft length 2), generating the first turn of two
  evaluation attempts greedily (256 tokens) with and without the adapter;
- in the `llama-server` binary that
  [KaggleInferenceServer](https://github.com/Nasus20202/KaggleInferenceServer)
  builds for the T4 (llama.cpp `v0.6.0`), with MTP and reasoning on.

### Decision rule (fixed before the runs)

1. A variant whose adapter does not load on the served GGUF, or does not lower
   the served perplexity on the evaluation trajectories, is out.
2. Of the rest, the one with the longest sequence length that fits.
3. Then the fastest training throughput.
4. Ties go to 16-bit LoRA (fewer moving parts).

## Results

### Rendering

| Family  | Turns checked | Served prompt = HF render | Served prompt is a prefix of the full render |
| ------- | ------------: | ------------------------: | -------------------------------------------: |
| Qwen3.5 |            76 |                        76 |                                           76 |
| Gemma 4 |            76 |                        76 |                                           66 |

Over all 3,428 assistant turns of the 154 clean attempts:

- **Qwen3.5** is append-only in every turn. A trajectory rendered once is
  exactly the sequence of contexts the model saw.
- **Gemma 4** is not in 637 turns (19%). When an assistant message holds both
  text and a tool call, the next served prompt closes the model turn
  (`<turn|>`), while the full render keeps it open. In those turns the
  training context differs from serving by that token. The researcher chose to
  keep full renders and judge the effect by served perplexity; the fallback,
  one example per mismatching segment, would cost 3.5x the training tokens.
  Gemma's adapters lower its served perplexity (below), so the fallback is not
  needed.

The template check passes once Gemma's BOS is compared as a token: llama.cpp
adds it when tokenizing, while the template writes `<bos>` as text.

Raw: [`raw/rendering/`](raw/rendering/).

### Memory and speed on two T4s

Kaggle, 2026-11-20, two T4 with 14.6 GiB each.

| Variant           | Layers GPU 0 / 1 | 4,096            | 8,192            | 12,288          | 16,384           | Training tok/s | Skipped steps |
| ----------------- | ---------------- | ---------------- | ---------------- | --------------- | ---------------- | -------------: | ------------: |
| Gemma 4 E2B LoRA  | 0–14 / 15–34     | 4.3 / 3.8, 396   | 5.4 / 5.1, 285   | 6.8 / 6.4, 218  | 8.3 / 7.9, 178   |            240 |             0 |
| Gemma 4 E2B QLoRA | 0–6 / 7–34       | 2.6 / 2.1, 499   | 3.2 / 3.3, 337   | 4.3 / 4.5, 247  | 5.6 / 5.8, 204   |            271 |             0 |
| Gemma 4 E4B LoRA  | 0–16 / 17–41     | 7.1 / 6.2, 260   | 8.9 / 7.7, 198   | 10.8 / 9.4, 149 | 12.8 / 11.1, 127 |            167 |             0 |
| Gemma 4 E4B QLoRA | 0–11 / 12–41     | 3.8 / 3.2, 304   | 5.4 / 4.7, 221   | 7.0 / 6.3, 177  | 8.7 / 8.0, 150   |            192 |             0 |
| Qwen3.5 4B LoRA   | 0–13 / 14–31     | 5.6 / 5.3, 293   | 7.1 / 6.8, 266   | 8.5 / 8.3, 250  | 10.0 / 9.8, 239  |            240 |             3 |
| Qwen3.5 4B QLoRA  | 0–10 / 11–31     | 3.1 / 2.6, 278   | 4.4 / 4.1, 255   | 5.6 / 5.5, 244  | 6.8 / 6.9, 233   |            235 |             2 |
| Qwen3.5 9B LoRA   | 0–11 / 12–31     | 10.5 / 10.2, 219 | 12.4 / 12.3, 203 | out of memory   | out of memory    |            205 |             2 |
| Qwen3.5 9B QLoRA  | 0–7 / 8–31       | 6.1 / 4.5, 205   | 7.6 / 6.4, 197   | 9.1 / 8.5, 178  | 10.6 / 10.5, 171 |            186 |             3 |

Sequence-length cells: peak memory in GiB on GPU 0 / GPU 1 (PyTorch's
allocated maximum), then tokens per second of a full training step. Training
tok/s: over the 30 steps, at the length each variant trains with. Skipped
steps: steps whose float16 gradients overflowed and that the loss scaler
skipped.

What this shows:

- **Whole trajectories fit for every model with QLoRA.** The longest
  training trajectory is 16.2K tokens, so 16,384 is the length needed. With
  QLoRA the fullest GPU still has 4 to 9 GiB free at that length.
- **16-bit LoRA fits for all but Qwen3.5 9B.** The 9B's float16 weights alone
  take 8.7 GiB on each GPU, and 12,288 tokens no longer fit. At 8,192 tokens
  it trains on 5 of the 10 training trajectories and can evaluate on none of
  the 6 evaluation ones.
- **QLoRA is not slower.** Dequantizing NF4 weights costs time per matrix
  multiplication, but at these lengths the step is dominated by attention and
  recomputation over long sequences. Qwen3.5 4B runs at the same speed in
  both precisions; Gemma runs 13 to 15% faster with QLoRA (the cause was not
  investigated).
- **Memory grows linearly with length** (about 0.4 to 0.6 GiB per 1K tokens
  on the fuller GPU), as expected with gradient checkpointing and chunked
  loss: no full attention matrix or full logits tensor is materialized.
- **Cost.** At 170 to 270 tokens per second, one million training tokens take
  about 1.0 to 1.7 hours of a two-T4 session. Splitting over two GPUs buys
  memory only, so a session trains one model at a time.

Before the float16 cast of frozen matrices was added, QLoRA kept embeddings
and output head in float32; the Qwen3.5 9B QLoRA variant then fitted only
8,192 tokens, with 7.6 GiB of those matrices on GPU 0. The QLoRA variants were
rerun with the cast; the LoRA path does not use it.

### Evaluation loss

Target-token loss on the six evaluation trajectories, before → after 30 steps:

| Model       | LoRA          | QLoRA         |
| ----------- | ------------- | ------------- |
| Gemma 4 E2B | 1.306 → 0.918 | 1.320 → 0.945 |
| Gemma 4 E4B | 1.211 → 0.831 | 1.256 → 0.856 |
| Qwen3.5 4B  | 0.595 → 0.594 | 0.632 → 0.621 |
| Qwen3.5 9B  | not measured  | 0.527 → 0.524 |

The trajectories were written by Qwen3.5 9B, which explains the pattern:

- **Gemma** starts at twice Qwen3.5 4B's loss: another model family, with
  other habits in reasoning and tool calls. Ten trajectories teach it that
  style, and the gain carries to unseen scenarios.
- **Qwen3.5 4B** already predicts its larger sibling's text well, so there is
  little to learn from ten examples. Its adapter also covers only part of the
  model: Qwen3.5 interleaves 8 full-attention layers with 24 linear-attention
  layers, whose projections have other names and are not among the targets.
- **Qwen3.5 9B** trains on its own outputs. Its LoRA run, which kept only 5
  trajectories, memorised them (training loss 0.06 at the end).

The QLoRA rows start slightly higher than the LoRA rows because the 4-bit
quantization of the base weights itself changes the model.

### Served model

Every adapter converted to GGUF (40 to 67 MiB) and loaded with `--lora` on
its served file. Served perplexity on the six evaluation trajectories:

| Model       | Without adapter | LoRA          | QLoRA         |
| ----------- | --------------- | ------------- | ------------- |
| Gemma 4 E2B | 2.514 ± 0.029   | 2.215 ± 0.023 | 2.237 ± 0.024 |
| Gemma 4 E4B | 2.332 ± 0.025   | 2.036 ± 0.019 | 2.042 ± 0.020 |
| Qwen3.5 4B  | 8.91 ± 0.22     | 8.61 ± 0.21   | 8.63 ± 0.21   |
| Qwen3.5 9B  | 5.89 ± 0.13     | 7.40 ± 0.18   | 6.29 ± 0.15   |

- **The adapter transfers from checkpoint to served file.** For Gemma (11 to
  13% lower) and Qwen3.5 4B (3% lower) the served file improves in the same
  direction as the training loss, although the adapter was trained on a
  float16 or NF4 copy of the weights and applied to a different 4-bit
  format.
- **LoRA and QLoRA give the same served model.** Their perplexities differ by
  less than one standard error for every model.
- **The 9B adapters make the served model worse.** Trained on its own outputs,
  they fit the training trajectories and nothing else; the LoRA adapter,
  trained on half the data, more so.

Qwen3.5's higher absolute perplexity comes from the scored tool results
(Kubernetes output) and is not comparable across families; only the change
within a row matters.

### Training the GGUF directly

Not possible with this stack, and not needed:

- transformers 5.17 (with `gguf` 0.17.1) loads Qwen3.5 4B's `Q4_K_M` file
  only by dequantizing every weight to 16-bit, since it has no 4-bit GGUF
  kernel. Training it would cost what 16-bit LoRA costs.
- Gemma 4's GGUF does not load (the config parser rejects per-layer MLP
  sizes).
- On the first 2,048 tokens of the six evaluation trajectories, the
  dequantized GGUF and the checkpoint have the same target-token loss (0.518
  and 0.519). The served file is a close copy of the checkpoint, which is why
  an adapter trained on the checkpoint carries over.

Raw: [`raw/gguf/transformers-load.json`](raw/gguf/transformers-load.json).

### Rest of the serving stack

| Check                                                      | Without adapter | With adapter                 | Result                                |
| ---------------------------------------------------------- | --------------- | ---------------------------- | ------------------------------------- |
| Router (`b10964`), preset `lora = …`                       | —               | loads next to the base model | works                                 |
| MTP draft acceptance, local                                | 0.76            | 0.72 (random adapter: 0.73)  | works, acceptance within a few points |
| Kaggle `llama-server` (`v0.6.0`), two attempts: acceptance | 0.73, 0.78      | 0.69, 0.80                   | works                                 |
| Kaggle `llama-server`, tokens per second                   | 74              | 67, 73                       | works                                 |

- The MTP drafter is the base model's own prediction head, which the adapter
  does not change. Acceptance therefore can only drop where the adapted model
  diverges from the base, and it stays close.
- On Kaggle the greedy first turns differ with and without the adapter, so
  the adapter is applied.
- The Kaggle binary needs the CUDA 12 runtime and cuBLAS, which the current
  Kaggle image no longer ships; they were installed from NVIDIA's wheels.

Raw: [`raw/router/`](raw/router/), [`raw/serve/serve.json`](raw/serve/serve.json).

### Decision rule

| Model       | Step 1 (loads, lowers served perplexity) | Step 2 (length) | Step 3 (tok/s)      | Selected |
| ----------- | ---------------------------------------- | --------------- | ------------------- | -------- |
| Gemma 4 E4B | both                                     | both 16,384     | QLoRA 192, LoRA 167 | QLoRA    |
| Qwen3.5 4B  | both                                     | both 16,384     | LoRA 240, QLoRA 235 | LoRA     |
| Gemma 4 E2B | both                                     | both 16,384     | QLoRA 271, LoRA 240 | QLoRA    |
| Qwen3.5 9B  | neither                                  | —               | —                   | none     |

For Qwen3.5 4B the step-3 gap (2%) is within run-to-run noise, so the rule's
choice there is effectively a tie.

## Conclusions

- The protocol works end to end on Kaggle's two T4s for all four models:
  whole trajectories, adapters that load on the served GGUF and in every part
  of the serving stack. Nothing in the path forces truncation, a different
  stack or a different GPU.
- QLoRA is the variant that fits every model, at no cost in speed or served
  quality. 16-bit LoRA is as good where it fits.
- The training checkpoint is the right thing to train: the served GGUF is a
  faithful copy of it, and adapters transfer.
- The teacher matters more than the precision. Qwen3.5 9B's trajectories move
  Gemma a lot and Qwen3.5 4B little, and harm the 9B itself. This bears on
  the teacher question (#276, #277), not on the protocol.

## Limitations

- The adapters train on 10 trajectories for 30 steps. They show that the path
  works and what it costs, not how much fine-tuning improves the agent; no
  agent run used them.
- Served perplexity includes tool-result tokens and was measured on the local
  Vulkan GPU; it indicates direction only.
- Qwen3.5's adapters leave its 24 linear-attention layers unadapted, which
  may understate what fine-tuning can do for Qwen.
- The layer split is estimated from parameter sizes, not measured per layer;
  the fuller GPU differs by up to 1.7 GiB from the other.
- The MTP and Kaggle-server checks used a Qwen3.5 4B LoRA adapter from an
  earlier round of this check (same format, targets and rank).
- One run per variant: throughput differences of a few percent are within
  noise.

## Earlier tries

- **One T4 per variant.** The first round trained each variant on a single
  T4, with transformers + PEFT and with Unsloth. With PEFT, Qwen3.5 4B ran out
  of memory at 16,384 tokens in both precisions. Unsloth fitted 16,384 tokens
  for Qwen but refuses float16 for Gemma 4 and trains it in float32, which ran
  out of memory in the first step. This led to transformers + PEFT split over
  both T4s (D-054).
- **TPU.** A Kaggle TPU v5e-8 (8 chips with 15.7 GiB each) with TorchTPU 0.3
  was tried, one model per chip in bfloat16, because a TPU could hold whole
  trajectories without quantization. Sessions were first not granted (the
  account lacked identity verification); the run that finally got a TPU took
  45 minutes, 41 of them downloading checkpoints. Qwen3.5 9B did not fit on a
  chip (about 18 GB of bfloat16 weights), and the other models loaded but ran
  out of memory already at 4,096 tokens (Gemma 4 E4B's weights alone took 14.9
  GiB). Splitting a model over several chips was not tried; the TPU path was
  dropped.

## Evidence

| Evidence                          | Location                                                                                                                                                     |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Rendering checks                  | [`raw/rendering/`](raw/rendering/)                                                                                                                           |
| Training runs, one per variant    | [`raw/run/<variant>/`](raw/run/): `metrics.json` (settings, placement, sweep, loss, throughput), `log.txt`, `status.json`, `served.json` (served perplexity) |
| Kaggle sessions                   | [`raw/run/session-*.json`](raw/run/): commit, GPUs, SHA-256 of the rendered data (identical in all sessions)                                                 |
| GGUF loaded in transformers       | [`raw/gguf/transformers-load.json`](raw/gguf/transformers-load.json)                                                                                         |
| Router preset and MTP comparisons | [`raw/router/`](raw/router/)                                                                                                                                 |
| Kaggle inference server           | [`raw/serve/serve.json`](raw/serve/serve.json)                                                                                                               |
