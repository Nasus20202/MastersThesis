# Fine-tuning results: Qwen3.5 9B teacher

The four adapters trained on the Qwen3.5 9B teacher's dataset, and their runs
on the benchmark. Questions, method, conclusions and limitations are in the
[fine-tuning study](README.md).

## Training run

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
  environment moved to Python 3.15 with the dependency update `54fbb242`;
  there `flash-linear-attention` fails at import (torch 2.14 has no
  `torch.compile` on Python 3.15), so transformers ran the linear-attention
  layers in its reference PyTorch code. Gemma, which has no linear attention,
  was at the proof of concept's speed. The DeepSeek-teacher Qwen adapters were
  trained on Python 3.14 with the kernels
  ([results-deepseek-flash-data.md](results-deepseek-flash-data.md)).

## Published adapters

Each private repository holds `adapter_model.safetensors`,
`adapter_config.json`, `adapter-f16.gguf` (converted with the pinned llama.cpp
image) and `metrics.json`.

| Repository                                   | Revision                                   | SHA-256, `adapter-f16.gguf`                                        | SHA-256, `adapter_model.safetensors`                               |
| -------------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `Nasus20202/gemma-4-e2b-lora-qwen35-9b-data` | `c8c9c31d29476ce04c2b0f9ac303fdd70f0152cf` | `ae61fc84381f948948e981eda4c512ef5c0b0dd236db56f7d5842939796ef50e` | `8599787892b1e9b7f2c1260757019505a064061841144da88aff2a039c092b2f` |
| `Nasus20202/gemma-4-e4b-lora-qwen35-9b-data` | `b1a8f38558f550f3664b90af1e0af86d474fb4bd` | `c8611f84d2e2b0b92890c425e631dcb9a80c3cc5f6d8d4628cff4ecf51b3df10` | `9ce2fc5f806f90b16a7f8ccb033036ecf01f337f261aee0ee7508a78af51a62d` |
| `Nasus20202/qwen35-4b-lora-qwen35-9b-data`   | `45faaf766c3c0fe63a052d18470394be811a91cd` | `23cb6a3aacbb6005ce9235db5c5781fea124724cee409b912b60b892adf44e36` | `dfdf606d63efbddcf46140305163799a59d51fb97c2b3e673998aee565b5c82c` |
| `Nasus20202/qwen35-9b-lora-qwen35-9b-data`   | `355b8890eccd6874dd77817c1271f42ae4fb5adf` | `6d852c009fb1cc4342e973471eac8d51b260e062d1736119f2e5442bba258a93` | `ad862a22b4944796497560012e4c777b7275564e2b1ddb293fffeb6c7c3f410a` |

## Validation scenarios (never trained on)

Each adapter and its base model, the frozen prompt agent, 3 attempts on each of
the 8 validation scenarios (24 attempts per arm), Kaggle sessions of the pinned
KIS 1.4.0 image. The score of an attempt is the weighted share of its passed
checks; the macro score is the mean over scenarios of the scenario's mean
score. Differences are adapter minus base, with the 95% paired bootstrap interval over
scenarios of [`benchmark/cmd/analyze`](../../../benchmark/cmd/analyze)
(10,000 draws, seed 1); [`results.py`](results.py) reproduces every number here
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

## Stopping behaviour (question 2)

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

## Held-out loss and validation

The held-out loss does not predict the validation score: the Gemma adapters end
below their base models' loss, the Qwen adapters above (table above), and the
score change is positive for three adapters and the largest for a Qwen adapter.

## Trained-on scenarios (first sweep)

This run was added after the validation runs and was not part of the method
fixed before training. Each adapter ran the prompt agent once on all 46
development scenarios: the 38 trained-on scenarios and the 8 validation
scenarios. Each base model ran the same sweep, once (see the evidence below). The Qwen3.5
9B teacher has no trajectory on `readonly-rootfs`, so its adapters trained on
37 of the 38. [`results.py`](results.py) computes the tables from the raw attempts;
the intervals are those of `benchmark/cmd/analyze` (paired bootstrap over
scenarios, 10,000 draws, seed 1).

| Model       | Arm     | Scenarios     | Macro score | Full success | Mean turns | Turn limit | Stopped without full success | Context or server error | Mean peak context, tokens |
| ----------- | ------- | ------------- | ----------: | -----------: | ---------: | ---------: | ---------------------------: | ----------------------: | ------------------------: |
| Gemma 4 E2B | adapter | 38 trained-on |       0.667 |        24/38 |       34.4 |         16 |                            1 |                       2 |                    13,943 |
| Gemma 4 E2B | base    | 38 trained-on |       0.362 |        12/38 |        9.4 |          0 |                           26 |                       0 |                     5,977 |
| Gemma 4 E2B | adapter | 8 validation  |       0.417 |          3/8 |       40.9 |          5 |                            0 |                       0 |                    16,016 |
| Gemma 4 E2B | base    | 8 validation  |       0.125 |          1/8 |       11.6 |          0 |                            7 |                       0 |                     6,584 |
| Gemma 4 E4B | adapter | 38 trained-on |       0.822 |        31/38 |       27.0 |          8 |                            0 |                       1 |                    10,890 |
| Gemma 4 E4B | base    | 38 trained-on |       0.618 |        21/38 |       10.9 |          0 |                           17 |                       0 |                     6,139 |
| Gemma 4 E4B | adapter | 8 validation  |       0.500 |          4/8 |       37.9 |          4 |                            0 |                       0 |                    17,856 |
| Gemma 4 E4B | base    | 8 validation  |       0.646 |          4/8 |        9.9 |          0 |                            4 |                       0 |                     5,767 |
| Qwen3.5 4B  | adapter | 38 trained-on |       0.851 |        30/38 |       25.8 |          4 |                            3 |                       1 |                    11,230 |
| Qwen3.5 4B  | base    | 38 trained-on |       0.848 |        30/38 |       36.5 |         15 |                            2 |                       0 |                    14,030 |
| Qwen3.5 4B  | adapter | 8 validation  |       0.292 |          2/8 |       41.2 |          5 |                            0 |                       1 |                    20,494 |
| Qwen3.5 4B  | base    | 8 validation  |       0.688 |          5/8 |       41.1 |          3 |                            1 |                       0 |                    17,064 |
| Qwen3.5 9B  | adapter | 38 trained-on |       0.961 |        36/38 |       22.1 |          3 |                            1 |                       0 |                     9,648 |
| Qwen3.5 9B  | base    | 38 trained-on |       0.982 |        36/38 |       22.4 |          1 |                            2 |                       0 |                     9,541 |
| Qwen3.5 9B  | adapter | 8 validation  |       1.000 |          8/8 |       23.9 |          0 |                            0 |                       0 |                    13,250 |
| Qwen3.5 9B  | base    | 8 validation  |       0.812 |          6/8 |       24.9 |          1 |                            1 |                       0 |                    13,488 |

On all 46 the adapters score 0.623 (E2B), 0.766 (E4B), 0.754 (Qwen3.5 4B) and
0.967 (Qwen3.5 9B), against 0.321, 0.623, 0.820 and 0.952 for the bases. All
four context or server errors of the adapters are requests above the
32,768-token window; none is an HTTP 5xx error.

| Model                   | Scenarios                 | Adapter minus base | 95% interval     | Scenarios higher / equal / lower |
| ----------------------- | ------------------------- | -----------------: | ---------------- | -------------------------------- |
| Gemma 4 E2B             | all 46                    |             +0.303 | +0.156 to +0.449 | 17 / 26 / 3                      |
| Gemma 4 E2B             | 38 trained-on             |             +0.305 | +0.140 to +0.476 | 14 / 21 / 3                      |
| Gemma 4 E2B             | 8 validation, one attempt |             +0.292 | +0.042 to +0.625 | 3 / 5 / 0                        |
| Gemma 4 E4B             | all 46                    |             +0.143 | −0.018 to +0.301 | 14 / 26 / 6                      |
| Gemma 4 E4B             | 38 trained-on             |             +0.204 | +0.029 to +0.375 | 13 / 22 / 3                      |
| Gemma 4 E4B             | 8 validation, one attempt |             −0.146 | −0.479 to +0.167 | 1 / 4 / 3                        |
| Qwen3.5 4B              | all 46                    |             −0.066 | −0.199 to +0.066 | 6 / 30 / 10                      |
| Qwen3.5 4B              | 38 trained-on             |             +0.003 | −0.129 to +0.137 | 5 / 27 / 6                       |
| Qwen3.5 4B              | 8 validation, one attempt |             −0.396 | −0.771 to +0.000 | 1 / 3 / 4                        |
| Qwen3.5 9B              | all 46                    |             +0.015 | −0.057 to +0.087 | 4 / 40 / 2                       |
| Qwen3.5 9B              | 38 trained-on             |             −0.021 | −0.092 to +0.032 | 2 / 34 / 2                       |
| Qwen3.5 9B              | 8 validation, one attempt |             +0.188 | +0.000 to +0.438 | 2 / 6 / 0                        |
| All four models, pooled | all 46                    |             +0.099 | +0.033 to +0.164 | —                                |
| All four models, pooled | 38 trained-on             |             +0.123 | +0.053 to +0.193 | —                                |
| All four models, pooled | 8 validation, one attempt |             −0.016 | −0.167 to +0.141 | —                                |

Score per scenario, one attempt each, base → adapter:

| Scenario                                | Gemma 4 E2B | Gemma 4 E4B |  Qwen3.5 4B |  Qwen3.5 9B |
| --------------------------------------- | ----------: | ----------: | ----------: | ----------: |
| automount-token-disabled                | 0.00 → 0.00 | 0.00 → 1.00 | 1.00 → 0.00 | 1.00 → 1.00 |
| basic-deploy-service                    | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| bind-precreated-pv                      | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| configmap-update-needs-reload           | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| daemonset-missing-toleration            | 0.00 → 1.00 | 0.00 → 0.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| deprecated-api-group                    | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| gateway-httproute-basic                 | 0.25 → 0.00 | 0.25 → 0.25 | 1.00 → 0.50 | 1.00 → 1.00 |
| gateway-route-misconfig                 | 0.00 → 1.00 | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| headless-service-dns                    | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| image-pull-failure                      | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| immutable-configmap                     | 0.33 → 1.00 | 0.33 → 1.00 | 0.33 → 1.00 | 1.00 → 1.00 |
| ingress-path-routing                    | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| job-backoff-exhausted                   | 1.00 → 0.00 | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| limitrange-defaults                     | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| liveness-restart-loop                   | 0.00 → 1.00 | 1.00 → 0.00 | 0.00 → 1.00 | 1.00 → 1.00 |
| networkpolicy-blocked                   | 0.00 → 0.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| networkpolicy-cross-namespace           | 0.00 → 0.00 | 1.00 → 0.00 | 1.00 → 0.00 | 1.00 → 0.00 |
| networkpolicy-egress-dns                | 0.00 → 0.00 | 0.00 → 0.00 | 0.80 → 0.00 | 0.80 → 1.00 |
| node-selector-no-match                  | 0.00 → 1.00 | 0.00 → 1.00 | 0.00 → 1.00 | 1.00 → 1.00 |
| nonroot-securitycontext                 | 0.00 → 0.00 | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| oom-memory-limit                        | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| overprivileged-serviceaccount           | 0.50 → 0.50 | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| private-registry-pull-secret            | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| psa-restricted-namespace                | 0.33 → 0.33 | 0.67 → 1.00 | 0.33 → 1.00 | 1.00 → 1.00 |
| pvc-pending                             | 0.00 → 1.00 | 1.00 → 0.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| qos-guaranteed-requests                 | 0.33 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| readiness-probe-breaks-endpoints        | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| readonly-rootfs (no teacher data)       | 0.00 → 0.00 | 0.00 → 0.00 | 0.00 → 0.00 | 1.00 → 0.50 |
| resourcequota-blocked                   | 1.00 → 1.00 | 0.75 → 1.00 | 0.25 → 1.00 | 1.00 → 1.00 |
| rollback-bad-release                    | 0.00 → 0.00 | 0.50 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| secret-ref-mismatch                     | 0.00 → 1.00 | 0.00 → 1.00 | 1.00 → 0.67 | 1.00 → 1.00 |
| service-targetport-mismatch             | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| sidecar-shared-volume                   | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| startup-probe-slow-app                  | 0.00 → 1.00 | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| statefulset-headless-service            | 0.00 → 0.00 | 0.00 → 1.00 | 0.50 → 0.50 | 1.00 → 1.00 |
| statefulset-volumeclaim                 | 0.00 → 0.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| stuck-terminating-pod                   | 0.00 → 1.00 | 1.00 → 1.00 | 1.00 → 0.67 | 1.00 → 1.00 |
| taint-toleration-placement              | 1.00 → 0.50 | 1.00 → 1.00 | 1.00 → 1.00 | 0.50 → 1.00 |
| antiaffinity-unschedulable (validation) | 0.00 → 1.00 | 0.00 → 0.00 | 0.00 → 0.00 | 1.00 → 1.00 |
| configmap-missing-key (validation)      | 1.00 → 1.00 | 1.00 → 1.00 | 0.50 → 1.00 | 0.50 → 1.00 |
| container-crash-loop (validation)       | 0.00 → 0.00 | 1.00 → 0.00 | 1.00 → 0.00 | 1.00 → 1.00 |
| dns-name-resolution (validation)        | 0.00 → 0.00 | 1.00 → 1.00 | 1.00 → 1.00 | 1.00 → 1.00 |
| hostpath-fsgroup (validation)           | 0.00 → 0.33 | 0.33 → 0.00 | 1.00 → 0.00 | 0.00 → 1.00 |
| missing-rbac-binding (validation)       | 0.00 → 0.00 | 1.00 → 1.00 | 1.00 → 0.00 | 1.00 → 1.00 |
| stale-image-ifnotpresent (validation)   | 0.00 → 0.00 | 0.33 → 1.00 | 1.00 → 0.33 | 1.00 → 1.00 |
| unschedulable-cpu-request (validation)  | 0.00 → 1.00 | 0.50 → 0.00 | 0.00 → 0.00 | 1.00 → 1.00 |

- **The Gemma adapters gain on the scenarios they were trained on; the Qwen
  adapters do not, because their bases already solve those scenarios.** On the 38
  trained-on scenarios Gemma 4 E2B gains 0.305 and Gemma 4 E4B 0.204, and both
  intervals exclude zero. Qwen3.5 4B (+0.003) and 9B (−0.021) do not separate
  from their bases, which already score 0.848 and 0.982 there. Pooled over the
  four models the adapters gain 0.123 (+0.053 to +0.193) on the trained-on
  scenarios and lose 0.016 (−0.167 to +0.141) on the validation scenarios in
  the sweep's single attempt.
- **For both Gemma models the trained-on gain is several times the validation
  gain.** Over 3 attempts the validation gains are 0.069 (E2B) and 0.035 (E4B),
  against 0.305 and 0.204 on the trained-on scenarios. The two scenario sets are
  about equally hard for the E4B base (0.618 and 0.646 in the sweep), so the
  Gemma adapters learned mostly the training scenarios.
- **Where E4B gains and loses.** It gains on 13 trained-on scenarios, 9 of them
  from 0. It loses on 3 (`liveness-restart-loop`, `networkpolicy-cross-namespace`,
  `pvc-pending`); on each the base made a full repair in 7 to 10 turns, while the
  adapter used all 50 turns and scored 0.
- **One attempt per validation scenario is too few to judge an adapter.** The
  sweep's single attempts on the 8 validation scenarios give 0.417, 0.500, 0.292
  and 1.000, against 0.340, 0.611, 0.861 and 0.806 over 3 attempts. Qwen3.5 4B
  failed 4 scenarios it had solved in all 3 earlier attempts, with the same
  model, adapter and limits; in `missing-rbac-binding` it spent 44 of its 50
  turns listing resources before it reached the RBAC objects. Qwen3.5 9B solved
  `unschedulable-cpu-request` in the single attempt, where two of its three
  validation attempts had ended on HTTP 525 errors.
- **The stopping change holds on the trained-on scenarios, and for Qwen it
  runs the other way.** The E2B and E4B bases ended 26 and 17 of 38 with their
  own report and less than full marks, the adapters 1 and 0; the adapters
  reached the turn limit 16 and 8 times, the bases never, and the E4B adapter
  used 2.5 times the turns. Qwen3.5 4B's base reached the turn limit in 15 of 38
  and its adapter in 4 (36.5 against 25.8 turns); Qwen3.5 9B's base in 1 and its
  adapter in 3.
- **`readonly-rootfs`, not in the Qwen teacher's data,** stays at 0 for every
  model except Qwen3.5 9B (0.50).

## Draft acceptance

The share of the multi-token-prediction head's drafted tokens that the model
accepted, over all responses of the arm (`draft_n_accepted` over `draft_n` of
llama.cpp's per-response timings):

| Model       | Validation, base | Validation, adapter | Sweep, base | Sweep, adapter |
| ----------- | ---------------: | ------------------: | ----------: | -------------: |
| Gemma 4 E2B |            0.405 |               0.397 |       0.425 |          0.438 |
| Gemma 4 E4B |            0.535 |               0.508 |       0.578 |          0.527 |
| Qwen3.5 4B  |            0.615 |               0.612 |       0.639 |          0.632 |
| Qwen3.5 9B  |            0.678 |               0.655 |       0.694 |          0.683 |

On the validation scenarios each adapter accepts fewer drafts than its base, by
0.003 to 0.051 (the sweeps: by 0.007 to 0.051, except Gemma 4 E2B, 0.013 more). A likely cause is that the drafts come from the base model's
prediction head, which the adapter was not trained on. The loss is too small
to change the cost of a run.

## Evidence

- [raw/run-20261126-002517-lora-qwen35-9b-data/](raw/run-20261126-002517-lora-qwen35-9b-data/):
  `session.json` (commit, data revision, settings) and each variant's
  `metrics.json` (training log, sweep, held-out losses per epoch and per
  attempt) and `log.txt`.
- Benchmark runs of the "Benchmark on Kaggle" workflow (`benchmark-kaggle.yaml`,
  prompt agent, main at `e26dda31`), each under `raw/benchmark/<run id>/` with the
  attempts, `run.json`, `results.json` and the Kaggle session statistics in
  `server/`. [`results.py`](results.py) reads them.

| Workflow run | Model       | Arm     | Scenarios × attempts                   | Run                                                                         |
| ------------ | ----------- | ------- | -------------------------------------- | --------------------------------------------------------------------------- |
| 45787279251  | Gemma 4 E4B | adapter | 8 validation × 3                       | [run-2026-11-26-10-21-49-018Z](raw/benchmark/run-2026-11-26-10-21-49-018Z/) |
| 45787284343  | Gemma 4 E4B | base    | 8 validation × 3                       | [run-2026-11-26-10-38-02-231Z](raw/benchmark/run-2026-11-26-10-38-02-231Z/) |
| 45788550579  | Gemma 4 E2B | base    | 8 validation × 3                       | [run-2026-11-26-10-55-41-090Z](raw/benchmark/run-2026-11-26-10-55-41-090Z/) |
| 45788554710  | Gemma 4 E2B | adapter | 8 validation × 3                       | [run-2026-11-26-11-23-19-500Z](raw/benchmark/run-2026-11-26-11-23-19-500Z/) |
| 45788558731  | Qwen3.5 4B  | base    | 8 validation × 3                       | [run-2026-11-26-12-44-57-945Z](raw/benchmark/run-2026-11-26-12-44-57-945Z/) |
| 45788563760  | Qwen3.5 4B  | adapter | 8 validation × 3                       | [run-2026-11-26-13-54-10-398Z](raw/benchmark/run-2026-11-26-13-54-10-398Z/) |
| 45788568733  | Qwen3.5 9B  | base    | 7 validation × 3 (one shard timed out) | [run-2026-11-26-14-54-48-134Z](raw/benchmark/run-2026-11-26-14-54-48-134Z/) |
| 45844415761  | Qwen3.5 9B  | base    | `antiaffinity-unschedulable` × 3       | [run-2026-11-26-16-39-32-870Z](raw/benchmark/run-2026-11-26-16-39-32-870Z/) |
| 45788574096  | Qwen3.5 9B  | adapter | 8 validation × 3                       | [run-2026-11-26-15-50-36-767Z](raw/benchmark/run-2026-11-26-15-50-36-767Z/) |
| 45774728575  | Gemma 4 E4B | adapter | 46 × 1                                 | [run-2026-11-26-08-04-10-096Z](raw/benchmark/run-2026-11-26-08-04-10-096Z/) |
| 45782058987  | Gemma 4 E4B | base    | 46 × 1                                 | [run-2026-11-26-08-39-29-198Z](raw/benchmark/run-2026-11-26-08-39-29-198Z/) |
| 45782063459  | Gemma 4 E2B | adapter | 46 × 1                                 | [run-2026-11-26-09-10-02-271Z](raw/benchmark/run-2026-11-26-09-10-02-271Z/) |
| 45857870095  | Qwen3.5 4B  | adapter | 46 × 1                                 | [run-2026-11-26-17-55-55-486Z](raw/benchmark/run-2026-11-26-17-55-55-486Z/) |
| 45872770169  | Qwen3.5 9B  | adapter | 46 × 1                                 | [run-2026-11-26-18-54-30-234Z](raw/benchmark/run-2026-11-26-18-54-30-234Z/) |
| 46090791919  | Gemma 4 E2B | base    | 46 × 1                                 | [run-2026-11-28-03-02-06-835Z](raw/benchmark/run-2026-11-28-03-02-06-835Z/) |
| 46103788997  | Qwen3.5 4B  | base    | 46 × 1                                 | [run-2026-11-28-08-01-55-125Z](raw/benchmark/run-2026-11-28-08-01-55-125Z/) |
| 46103791587  | Qwen3.5 9B  | base    | 46 × 1                                 | [run-2026-11-28-08-50-19-962Z](raw/benchmark/run-2026-11-28-08-50-19-962Z/) |

The three base sweeps of 2026-11-28 ran from branch `docs/26-deepseek-teacher-results`
at `d215150f`, after the first dispatch of four Qwen sweeps (16 shards) failed in
seconds because a Qwen session has 14 slots; those four runs hold no attempts and
were repeated with 12 shards. They also serve as the base of the DeepSeek
teacher's adapters.

A workflow run marked failed or cancelled can still hold every attempt: a
failed shard is usually one attempt ended by the context window, and a
cancelled start step is its 15-minute timeout. The attempt counts above are
those in the merged results.
