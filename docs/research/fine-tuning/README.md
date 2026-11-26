# Fine-tuning the four models

## Summary

This study trains the LoRA adapters of the fine-tuned condition (D-053) and
compares the two teachers' datasets (D-054). The method below is fixed before
the first training run; results follow after the runs.

## Question

1. Does an adapter trained on Qwen3.5 9B trajectories or one trained on
   DeepSeek V4.1 Flash trajectories make each of the four models a better agent
   on the validation scenarios?
2. Does an adapter change the measured stopping failures (Gemma stops before
   the repair is verified, Qwen keeps checking after it is done)?
3. Do 3 epochs over-train: does the held-out loss rise in the later epochs?
4. What does training cost, in A100 hours?

## Method

**Data.** The frozen whole-trajectory datasets of the two teachers
([training datasets](../training-dataset/README.md)): 153 Qwen3.5 9B and 181
DeepSeek V4.1 Flash examples per student, on the 38 training scenarios. DeepSeek
turns without reasoning are kept as returned (D-061) and attempts whose only
issue is a forceful write stay (D-061). The synthetic variants (S1, S2) are not
trained here.

| Dataset             | Examples | Tokens, Gemma students | Tokens, Qwen students | Mean length |
| ------------------- | -------: | ---------------------: | --------------------: | ----------: |
| Qwen3.5 9B          |      153 |                  1.18M |                 1.23M |       ~7.9K |
| DeepSeek V4.1 Flash |      181 |                  1.52M |                 1.55M |       ~8.5K |

Every example fits the 16,384-token cap.

**Adapters.** 8: each of Gemma 4 E2B, Gemma 4 E4B, Qwen3.5 4B and Qwen3.5 9B
on each teacher's dataset (D-060), with one training setup (D-059). Each teacher's four adapters are one Modal run, described by `finetune/manifests/lora-<teacher>-data.yaml`:

| Value                | Setting                                                        |
| -------------------- | -------------------------------------------------------------- |
| LoRA rank / alpha    | 16 / 16, dropout 0                                             |
| Adapted projections  | attention, MLP, Qwen3.5's Gated DeltaNet inputs                |
| Epochs               | 3, one trajectory per step (459 steps Qwen data, 543 DeepSeek) |
| Optimizer            | AdamW, learning rate 1e-4, weight decay 0.01, clipping at 1.0  |
| Schedule             | linear warm-up over 5% of the steps, cosine decay to 0         |
| Precision, seed, GPU | bfloat16, seed 0, one Modal A100 40 GB, 120-minute cap         |

The values were checked on 2026-11-25 against
[Google's Gemma 4 LoRA guide](https://ai.google.dev/gemma/docs/core/huggingface_text_finetune_qlora)
(rank 16, alpha 32, learning rate 2e-4, weight decay 0.01, clipping 1.0, 5%
warm-up, cosine, batch 16),
[Unsloth's Qwen3.5 guide](https://unsloth.ai/docs/models/qwen3.5/fine-tune)
(rank 16, alpha 16) and
[Unsloth's LoRA hyperparameter guide](https://unsloth.ai/docs/get-started/fine-tuning-llms-guide/lora-hyperparameters-guide)
(learning rate 2e-4 to 5e-6, 1–3 epochs, alpha equal to or twice the rank).
Qwen publishes no LoRA guide for Qwen3.5. The learning rate is lower than the
guides' because each step is one trajectory instead of a batch of 16.

**Held-out loss.** Mean cross-entropy over the assistant tokens, with and
without the adapter, before training, after each epoch and after training
(D-060), overall and for each attempt. The attempts are the same for every
adapter: the clean DeepSeek V4.1 Flash attempts of a run with 3 attempts on
each of the 8 validation scenarios, which are never trained on
(`benchmark/model-profiles/deepseek-flash.env`, listed in
`finetune/datasets/validation/deepseek-flash/manifest.yaml`). All 24 scored full marks; 16 fit the
16,384-token cap, and longer ones are dropped as in training, which leaves
`hostpath-fsgroup` out. One teacher's attempts serve both adapters as the same
reference: for the Qwen-trained adapter they also show whether what it learned
carries over to another teacher's solutions. The loss shows whether later
epochs over-train; it does not select anything. The same run on Qwen3.5 9B (Kaggle,
`finetune/datasets/validation/qwen35-9b/manifest.yaml`, the failed setups
rerun one at a time) gave 19 clean attempts, 2 or 3 per scenario; they were
collected after the Qwen-teacher training run started and are not part of it.

**Validation.** Each adapter and each base model runs the frozen prompt agent
on the 8 validation scenarios, 3 attempts each. Reported per model: score, the
stopping measures of the [stopping-rule study](../stopping-rule/README.md)
(attempts ending at the turn limit, final reports on an unverified repair,
turns after the repair is verified), and the held-out loss. The researcher
then decides which teacher's adapters go to the final experiment (D-060). An
adapter below its base model still goes (D-062).

**Serving and publishing.** llama.cpp loads the adapter as a GGUF next to the
unchanged base model file (D-062). Each adapter is converted to GGUF locally
and uploaded with its safetensors to a private Hugging Face repository; the
revision and SHA-256 are recorded here (D-062).

**Cost estimate before the runs.** About 5.6 A100 hours ($14) for the 8
adapters, up to about $21: the proof of concept measured Qwen3.5 4B (about
3,600 training tokens/s) and Gemma 4 E4B (1,260); the speeds of Qwen3.5 9B
and Gemma 4 E2B are guesses.

## Results

### Qwen3.5 9B teacher

One Modal run, `run-20261126-002517-lora-qwen35-9b-data`, from commit
`f1230bee` with `finetune/manifests/lora-qwen35-9b-data.yaml`: 4 variants, all
153 examples kept, 459 steps each, no errors. Modal assigned A100 80 GB
cards to the 40 GB request.

Held-out loss on the 16 DeepSeek V4.1 Flash validation attempts that fit the
cap (mean cross-entropy over the assistant tokens; lower is better):

| Model       | Base model | Epoch 1 | Epoch 2 | Epoch 3 | Tokens/s | Training, min | Peak GiB |
| ----------- | ---------: | ------: | ------: | ------: | -------: | ------------: | -------: |
| Gemma 4 E2B |      1.831 |   1.472 |   1.499 |   1.551 |    1,928 |            31 |     16.5 |
| Gemma 4 E4B |      1.663 |   1.376 |   1.405 |   1.466 |    1,425 |            41 |     23.8 |
| Qwen3.5 4B  |      1.098 |   1.096 |   1.127 |   1.150 |      762 |            81 |     18.6 |
| Qwen3.5 9B  |      1.022 |   1.060 |   1.088 |   1.113 |      721 |            85 |     30.1 |

- **Over-training (question 3).** For every model the held-out loss is lowest
  after epoch 1 and rises in epochs 2 and 3. The Gemma adapters stay below
  their base model; the Qwen adapters end above it on DeepSeek's solutions.
  The adapters after epoch 3 are the ones kept (D-062).
- **Cost (question 4).** Modal billed $10.21 for `train_variant` ($9.49 GPU,
  $0.44 CPU, $0.29 memory; the image build added $0.01), above the estimate of
  about $6.50 (up to $9). The log times, from container start to the saved
  adapter, were 35, 47, 89 and 95 minutes (4.4 hours); the bill covers 4.52
  A100 40 GB hours, which gives $2.26 per hour for the 2 cores and 8 GiB the
  run requested. The run asked for and was billed as `A100_40GB`, although the
  logged devices have 79.3 GiB (A100 80 GB), so the cost uses the 40 GB rate.
  Modal's list prices on 2026-11-26 ([modal.com/pricing](https://modal.com/pricing)):
  A100 40 GB $0.000583 per second ($2.10/h), CPU $0.0000131 per core-second,
  memory $0.00000222 per GiB-second, billed at the larger of the request and
  the use.
- **Qwen speed.** The Qwen students trained at 762 and 721 tokens/s; the proof
  of concept's A100 runs of Qwen3.5 4B recorded 2,161 and 2,380
  ([raw](../fine-tuning-proof-of-concept/raw/)). The run's
  environment (Python 3.15, CUDA 13.2, since the dependency update `54fbb242`)
  locks no `triton`, which the flash-linear-attention kernels need, so the
  linear-attention layers most likely ran the reference PyTorch code. Gemma,
  which has no linear attention, was at the proof of concept's speed.

### Published adapters

Each private repository holds `adapter_model.safetensors`,
`adapter_config.json`, `adapter-f16.gguf` (converted with the pinned llama.cpp
image) and `metrics.json`.

| Repository                                   | Revision                                   | SHA-256, `adapter-f16.gguf`                                        | SHA-256, `adapter_model.safetensors`                               |
| -------------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `Nasus20202/gemma-4-e2b-lora-qwen35-9b-data` | `c8c9c31d29476ce04c2b0f9ac303fdd70f0152cf` | `ae61fc84381f948948e981eda4c512ef5c0b0dd236db56f7d5842939796ef50e` | `8599787892b1e9b7f2c1260757019505a064061841144da88aff2a039c092b2f` |
| `Nasus20202/gemma-4-e4b-lora-qwen35-9b-data` | `b1a8f38558f550f3664b90af1e0af86d474fb4bd` | `c8611f84d2e2b0b92890c425e631dcb9a80c3cc5f6d8d4628cff4ecf51b3df10` | `9ce2fc5f806f90b16a7f8ccb033036ecf01f337f261aee0ee7508a78af51a62d` |
| `Nasus20202/qwen35-4b-lora-qwen35-9b-data`   | `45faaf766c3c0fe63a052d18470394be811a91cd` | `23cb6a3aacbb6005ce9235db5c5781fea124724cee409b912b60b892adf44e36` | `dfdf606d63efbddcf46140305163799a59d51fb97c2b3e673998aee565b5c82c` |
| `Nasus20202/qwen35-9b-lora-qwen35-9b-data`   | `355b8890eccd6874dd77817c1271f42ae4fb5adf` | `6d852c009fb1cc4342e973471eac8d51b260e062d1736119f2e5442bba258a93` | `ad862a22b4944796497560012e4c777b7275564e2b1ddb293fffeb6c7c3f410a` |

## Evidence

- [raw/run-20261126-002517-lora-qwen35-9b-data/](raw/run-20261126-002517-lora-qwen35-9b-data/):
  `session.json` (commit, data revision, settings) and each variant's
  `metrics.json` (training log, sweep, held-out losses per epoch and per
  attempt) and `log.txt`.
