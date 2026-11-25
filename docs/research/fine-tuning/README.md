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
epochs over-train; it does not select anything.

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
