# Fine-tuning the four models

## Summary

This study trains the LoRA adapters of the fine-tuned condition (D-053) and
compares the two teachers' datasets (D-054). The method below was fixed before
the first training run. Four adapters trained on the Qwen3.5 9B teacher's
trajectories were trained, published and run on the benchmark (ticket #27); the
DeepSeek V4.1 Flash teacher's adapters are not trained yet, so question 1
compares no teachers yet.

- **On scenarios never trained on, the adapters do not yet show a gain that
  holds across models.** Pooled over the four models the adapter scores 0.056
  higher than its base (95% interval −0.043 to +0.149); Qwen3.5 4B gains
  0.174 (+0.042 to +0.306), the other three do not separate from their bases.
- **The adapters change when the agent stops.** The Gemma bases stop without a
  full repair in 19 of 24 (E2B) and 12 of 24 (E4B) attempts and never reach the
  turn limit; their adapters stop that way in 2 and 3 and reach the turn limit in
  14 and 8, with 3 to 4 times as many turns. The Qwen bases already keep working,
  and their turn counts do not change.
- **The held-out loss rises after epoch 1 for every model**, so 3 epochs
  over-train by that measure (question 3). Training took 4.4 hours and cost
  $10.21 (question 4).

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

### Validation scenarios (never trained on)

Each adapter and its base model, the frozen prompt agent, 3 attempts on each of
the 8 validation scenarios (24 attempts per arm), Kaggle sessions of the pinned
KIS 1.4.0 image. The score of an attempt is the weighted share of its passed
checks; the macro score is the mean over scenarios of the scenario's mean
score. Differences are adapter minus base, with the 95% paired bootstrap interval over
scenarios of [`benchmark/cmd/analyze`](../../../benchmark/cmd/analyze)
(10,000 draws, seed 1); [`pooled.py`](pooled.py) reproduces every number here
from the raw attempts and pools the four models.

| Model       | Arm     | Macro score | Full success | Mean turns | Turn limit | Stopped without full success | Context or server error | Mean peak context, tokens |
| ----------- | ------- | ----------: | -----------: | ---------: | ---------: | ---------------------------: | ----------------------: | ------------------------: |
| Gemma 4 E2B | adapter |       0.340 |         7/24 |       42.0 |         14 |                            2 |                       1 |                    18,764 |
| Gemma 4 E2B | base    |       0.271 |         5/24 |       11.1 |          0 |                           19 |                       0 |                     7,026 |
| Gemma 4 E4B | adapter |       0.611 |        14/24 |       35.7 |          8 |                            3 |                       1 |                    17,492 |
| Gemma 4 E4B | base    |       0.576 |        12/24 |       12.0 |          0 |                           12 |                       0 |                     7,055 |
| Qwen3.5 4B  | adapter |       0.861 |        20/24 |       36.5 |          9 |                            0 |                       0 |                    17,482 |
| Qwen3.5 4B  | base    |       0.688 |        15/24 |       41.6 |          9 |                            3 |                       0 |                    18,360 |
| Qwen3.5 9B  | adapter |       0.806 |        19/24 |       30.2 |          3 |                            0 |                       3 |                    14,909 |
| Qwen3.5 9B  | base    |       0.861 |        20/24 |       30.1 |          4 |                            1 |                       0 |                    13,600 |

"Stopped without full success" counts attempts that ended with the agent's
own final report and less than full marks: a proxy for the stopping study's
final reports on an unverified repair, which is not computed here. "Context or
server error" counts attempts ended by the context window (32,768 tokens) or by
an error answer of the inference endpoint.

| Model                   | Adapter minus base | 95% interval     |
| ----------------------- | -----------------: | ---------------- |
| Gemma 4 E2B             |             +0.069 | −0.188 to +0.306 |
| Gemma 4 E4B             |             +0.035 | −0.222 to +0.264 |
| Qwen3.5 4B              |             +0.174 | +0.042 to +0.306 |
| Qwen3.5 9B              |             −0.056 | −0.153 to +0.028 |
| All four models, pooled |             +0.056 | −0.043 to +0.149 |

Mean score per scenario, base → adapter:

| Scenario                   | Gemma 4 E2B | Gemma 4 E4B |  Qwen3.5 4B |  Qwen3.5 9B |
| -------------------------- | ----------: | ----------: | ----------: | ----------: |
| antiaffinity-unschedulable | 0.00 → 0.67 | 1.00 → 1.00 | 0.67 → 0.67 | 1.00 → 1.00 |
| configmap-missing-key      | 0.33 → 0.50 | 0.50 → 1.00 | 0.50 → 1.00 | 1.00 → 1.00 |
| container-crash-loop       | 0.17 → 0.00 | 0.00 → 0.33 | 1.00 → 1.00 | 1.00 → 1.00 |
| dns-name-resolution        | 0.00 → 0.33 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| hostpath-fsgroup           | 0.00 → 0.00 | 0.67 → 0.00 | 0.44 → 0.67 | 0.33 → 0.33 |
| missing-rbac-binding       | 0.22 → 0.33 | 0.44 → 0.67 | 0.67 → 1.00 | 0.89 → 1.00 |
| stale-image-ifnotpresent   | 0.44 → 0.56 | 0.33 → 0.56 | 0.56 → 0.56 | 1.00 → 0.78 |
| unschedulable-cpu-request  | 1.00 → 0.33 | 0.67 → 0.33 | 0.67 → 1.00 | 0.67 → 0.33 |

- **Gain (question 1).** Three of the four adapters score above their bases, but only
  Qwen3.5 4B's interval excludes zero, and with 8 scenarios and 3 attempts each
  one scenario moves a model's macro score by up to 0.125. The first
  single-attempt run of the E4B pair read −0.146; with 3 attempts it reads
  +0.035.
- **Server errors in the Qwen3.5 9B adapter's result.** Two of its three
  `unschedulable-cpu-request` attempts ended on HTTP 525 answers of the inference
  endpoint (at turns 3 and 39), not on anything the model did; the third scored
  1.0. Without the two, the adapter's macro score is 0.889, the difference is
  +0.028 (−0.069 to +0.139) and the pooled difference is +0.076 (−0.021 to
  +0.172). The tables above keep the attempts as recorded.
- **Where the adapters gain and lose.** The adapters gain most on
  `configmap-missing-key` and `missing-rbac-binding` (up in all four models or
  equal) and lose on `hostpath-fsgroup` for E4B (0.67 → 0.00) and on
  `unschedulable-cpu-request` for E2B and E4B. Differences of one scenario's three
  attempts are not reliable.

### Stopping behaviour (question 2)

- **Gemma stops before the repair is complete, the adapter does not.** The E2B
  and E4B bases ended 19 and 12 of 24 attempts with a final report and less than
  full marks, and none at the turn limit, after 11.1 and 12.0 turns on average.
  Their adapters ended 2 and 3 attempts that way and 14 and 8 at the turn limit,
  after 42.0 and 35.7 turns. The extra turns are not free: failed kubectl
  changes rose (E4B: 53 of 105 against 23 of 63; E2B: 59 of 132 against 12 of
  41), the peak context grew from about 7,000 to about 18,000 tokens, and the
  adapters ended one attempt each on the 32,768-token window or an error of the inference endpoint (E2B, E4B).
- **Qwen kept working before and after.** The Qwen bases already used 30 to 42
  turns and reached the turn limit in 9 and 4 attempts; their adapters used 36.5
  and 30.2 turns and reached it in 9 and 3. The stopping change is a Gemma
  effect.
- **The score gain is not only persistence.** Qwen3.5 4B's gain of 0.174 came with
  no change in turns, and on E2B the adapter reaches full marks in 7 attempts
  against 5 while using 3.8 times the turns.

### Held-out loss and validation

The held-out loss does not predict the validation score: the Gemma adapters end
below their base models' loss, the Qwen adapters above (table above), and the
score change is positive for three adapters and the largest for a Qwen adapter.

## Evidence

- [raw/run-20261126-002517-lora-qwen35-9b-data/](raw/run-20261126-002517-lora-qwen35-9b-data/):
  `session.json` (commit, data revision, settings) and each variant's
  `metrics.json` (training log, sweep, held-out losses per epoch and per
  attempt) and `log.txt`.
