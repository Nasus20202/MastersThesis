# Fine-tuning results: DeepSeek V4.1 Flash teacher

The four adapters trained on the DeepSeek V4.1 Flash teacher's dataset, and
their runs on the benchmark. Questions, method, conclusions and limitations
are in the [fine-tuning study](README.md); the Qwen3.5 9B teacher's adapters
are in [results-qwen35-9b-data.md](results-qwen35-9b-data.md).

## Training runs

Two Modal runs on the frozen dataset `finetune/datasets/deepseek-flash`
(revision `be4314358093`): all 181 examples kept under the 16,384-token cap,
543 steps each (3 epochs), no errors. Held-out loss on the 16 DeepSeek V4.1
Flash validation attempts that fit the cap, as for the Qwen teacher.

| Model       | Run                                                 | Commit     | Python  |
| ----------- | --------------------------------------------------- | ---------- | ------- |
| Gemma 4 E2B | `run-20261126-194429-lora-deepseek-flash-data`      | `ce42cb50` | 3.15.2  |
| Gemma 4 E4B | `run-20261126-194429-lora-deepseek-flash-data`      | `ce42cb50` | 3.15.2  |
| Qwen3.5 4B  | `run-20261126-230134-lora-deepseek-flash-data-qwen` | `b9006f00` | 3.14.10 |
| Qwen3.5 9B  | `run-20261126-230134-lora-deepseek-flash-data-qwen` | `b9006f00` | 3.14.10 |

The first run used `finetune/manifests/lora-deepseek-flash-data.yaml` (all
four models); the second, a copy of it with only the two Qwen models, recorded
in the run's `session.json`. Training values are those of the Qwen-teacher run
(D-059).

| Model       | Base model | Epoch 1 | Epoch 2 | Epoch 3 | Tokens/s | Training, min | Peak GiB |
| ----------- | ---------: | ------: | ------: | ------: | -------: | ------------: | -------: |
| Gemma 4 E2B |      1.831 |   1.012 |   0.952 |   0.956 |    1,824 |            42 |     16.9 |
| Gemma 4 E4B |      1.663 |   0.904 |   0.857 |   0.869 |    1,385 |            55 |     24.4 |
| Qwen3.5 4B  |      1.098 |   0.776 |   0.749 |   0.759 |    4,024 |            19 |     14.9 |
| Qwen3.5 9B  |      1.022 |   0.731 |   0.711 |   0.722 |    2,712 |            29 |     26.5 |

- **Over-training (question 3).** For every model the held-out loss is lowest
  after epoch 2 and rises a little in epoch 3, and all four adapters end far
  below their base models. The held-out attempts are DeepSeek's own, so this
  loss favours the DeepSeek-teacher adapters and does not compare the two
  teachers. The adapters after epoch 3 are the ones kept (D-062).

- **Qwen speed and the Python version.** The first run also trained the two
  Qwen models; both were cancelled in epoch 3 (steps 425 and 459 of 543)
  when the `modal run` client crashed on the local machine, and the trainer
  kept no adapter of a cut-off run (each epoch's adapter has been saved since
  `c80c8c2b`). Their logs showed the cause of the slow Qwen training reported
  for the Qwen teacher: on Python 3.15, to which the dependency update
  `54fbb242` had moved the environment, `flash-linear-attention` fails at
  import (`torch.compile is not supported on Python 3.15+`) and transformers
  falls back to its reference PyTorch code.
  Installing `triton` alone did not help. The second run uses Python 3.14.10
  with the same torch, transformers, PEFT and fla versions; the Qwen students
  trained at 4,024 and 2,712 tokens/s, against 762 and 721 for the Qwen
  teacher, with a lower peak (14.9 and 26.5 GiB against 18.6 and 30.1). Both
  teachers' Gemma adapters and the Qwen-teacher Qwen adapters were trained on
  Python 3.15; only the DeepSeek-teacher Qwen adapters use the fla kernels,
  which compute the same layers with different rounding.
- **Cards.** Both runs asked for and were billed as `A100_40GB`; Modal gave the
  first run A100 40 GB cards and the second A100 80 GB cards.
- **Cost (question 4).** Modal billed $14.98 for this teacher's adapters:
  $11.65 for the first run (the two Gemma adapters and about 100 minutes of
  each cancelled Qwen variant), $0.70 for a Qwen run with `triton` on Python
  3.15, stopped after its memory sweep showed the old speed, $0.49 for a
  20-step check on Python 3.14 and $2.14 for the second run. The log times
  from container start to the saved adapter were 46 and 60 minutes (Gemma) and
  22 and 32 minutes (Qwen). By the share of the first run's container time,
  the four kept adapters cost about $6.20 (about $4.05 for Gemma, $2.14 for
  Qwen), against an estimate of about $12.40 that assumed the slow Qwen
  training.

## Published adapters

Each private repository holds `adapter_model.safetensors`,
`adapter_config.json`, `adapter-f16.gguf` (converted with the pinned llama.cpp
image) and `metrics.json`.

| Repository                                        | Revision                                   | SHA-256, `adapter-f16.gguf`                                        | SHA-256, `adapter_model.safetensors`                               |
| ------------------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| `Nasus20202/gemma-4-e2b-lora-deepseek-flash-data` | `9e27cd3c47bc2b9fbb58a6712771c146f92a451b` | `888f6fd1d8f4aa8af48b1cbe4c3356dbee9504c7f41bf705163c5d66f8ec7308` | `a601f370bb48093391bf927a4e4d70384a830c0a5dfeef1dd81ff6d5b22cd8d7` |
| `Nasus20202/gemma-4-e4b-lora-deepseek-flash-data` | `c483cf29905643f65cca80b8b3720f766ac37a13` | `2ecbeac5ceb3929992366f56df943c98e2456c698df0fa501212b7dbd2b6f1b6` | `41c6afa57d6958de21150da615800aaa308b31cff02c19826907d7fc21e9ef64` |
| `Nasus20202/qwen35-4b-lora-deepseek-flash-data`   | `d6c288eef8d060920f10c6c73e11539d24687906` | `388ecf86537959ccd22e8a8a755cdebefa6957365254f0df3ee43ee22061aee0` | `31e2df4e881c871e26a89e1a9d7d03e97533e5b830d45452ef010d11d3112620` |
| `Nasus20202/qwen35-9b-lora-deepseek-flash-data`   | `82b2fc861e86175fec923363bcf3b175ca7016bf` | `df64d5004550769713e4318e1af356b84a11fd2ecc2a56f1eeeb46ba8059cb68` | `9138a7852e0a1159489488dc2838632406b3c3f01eda79cb3cb147fc97673122` |

## Validation scenarios (never trained on)

Each adapter and its base model, the frozen prompt agent, 3 attempts on each of
the 8 validation scenarios (24 attempts per arm), Kaggle sessions of the pinned
KIS 1.4.0 image, with the same bases as the Qwen3.5 9B teacher's adapters
([results-qwen35-9b-data.md](results-qwen35-9b-data.md#validation-scenarios-never-trained-on)).
The score, the macro score and the paired bootstrap are defined there;
`python3 results.py --teacher deepseek-flash-data` reproduces every number here
from the raw attempts. The adapters' revisions in the Kaggle sessions are those of
the table above.

| Model       | Arm     | Macro score | Full success | Mean turns | Turn limit | Stopped without full success | Context or server error | Timed out | Mean peak context, tokens | Tokens per turn |
| ----------- | ------- | ----------: | -----------: | ---------: | ---------: | ---------------------------: | ----------------------: | --------: | ------------------------: | --------------: |
| Gemma 4 E2B | adapter |       0.340 |         6/24 |       19.2 |          0 |                           15 |                       3 |         0 |                    18,820 |             405 |
| Gemma 4 E2B | base    |       0.271 |         5/24 |       11.1 |          0 |                           19 |                       0 |         0 |                     7,026 |             337 |
| Gemma 4 E4B | adapter |       0.569 |        13/24 |       18.8 |          0 |                            6 |                       5 |         0 |                    20,399 |             472 |
| Gemma 4 E4B | base    |       0.576 |        12/24 |       12.0 |          0 |                           12 |                       0 |         0 |                     7,055 |             282 |
| Qwen3.5 4B  | adapter |       0.667 |        15/24 |       21.5 |          0 |                            0 |                       8 |         2 |                    21,542 |             511 |
| Qwen3.5 4B  | base    |       0.688 |        15/24 |       41.6 |          9 |                            3 |                       0 |         0 |                    18,360 |             192 |
| Qwen3.5 9B  | adapter |       0.778 |        18/24 |       16.0 |          0 |                            2 |                       3 |         1 |                    15,383 |             407 |
| Qwen3.5 9B  | base    |       0.861 |        20/24 |       30.1 |          4 |                            1 |                       0 |         0 |                    13,600 |             178 |

"Context or server error" counts attempts ended by the context window (32,768
tokens): 12 where the window cut the answer off (peak context 32,766 to 32,768)
and 7 where the server refused the request as too long (HTTP 400,
`exceed_context_size_error`). None of the 96 adapter attempts ended on an HTTP 5xx
answer. "Timed out" counts attempts ended by `context deadline exceeded` (3: two
Qwen3.5 4B attempts on `container-crash-loop`, scoring 0 and 1.0, and one Qwen3.5
9B attempt on `hostpath-fsgroup`, scoring 0); `results.py` reports them
separately from the other endings. "Tokens per turn" is the generated tokens over
the agent's turns.

| Model                   | Adapter minus base | 95% interval     |
| ----------------------- | -----------------: | ---------------- |
| Gemma 4 E2B             |             +0.069 | −0.146 to +0.292 |
| Gemma 4 E4B             |             −0.007 | −0.347 to +0.375 |
| Qwen3.5 4B              |             −0.021 | −0.292 to +0.222 |
| Qwen3.5 9B              |             −0.083 | −0.222 to +0.028 |
| All four models, pooled |             −0.010 | −0.137 to +0.116 |

Mean score per scenario, base → adapter:

| Scenario                   | Gemma 4 E2B | Gemma 4 E4B |  Qwen3.5 4B |  Qwen3.5 9B |
| -------------------------- | ----------: | ----------: | ----------: | ----------: |
| antiaffinity-unschedulable | 0.00 → 0.67 | 1.00 → 0.67 | 0.67 → 1.00 | 1.00 → 1.00 |
| configmap-missing-key      | 0.33 → 0.67 | 0.50 → 1.00 | 0.50 → 1.00 | 1.00 → 1.00 |
| container-crash-loop       | 0.17 → 0.33 | 0.00 → 1.00 | 1.00 → 0.67 | 1.00 → 1.00 |
| dns-name-resolution        | 0.00 → 0.00 | 1.00 → 0.33 | 1.00 → 1.00 | 1.00 → 1.00 |
| hostpath-fsgroup           | 0.00 → 0.00 | 0.67 → 0.00 | 0.44 → 0.33 | 0.33 → 0.33 |
| missing-rbac-binding       | 0.22 → 0.22 | 0.44 → 0.67 | 0.67 → 1.00 | 0.89 → 1.00 |
| stale-image-ifnotpresent   | 0.44 → 0.33 | 0.33 → 0.22 | 0.56 → 0.33 | 1.00 → 0.56 |
| unschedulable-cpu-request  | 1.00 → 0.50 | 0.67 → 0.67 | 0.67 → 0.00 | 0.67 → 0.33 |

- **Gain (question 1).** The adapters do not separate from their bases: one is
  above (E2B), three are below, the pooled difference is −0.010 and every
  interval includes zero. With 8 scenarios, one scenario moves a model's macro
  score by up to 0.125.
- **Where the adapters gain and lose.** `configmap-missing-key` is at or above
  the base for all four models and `missing-rbac-binding` for three (E2B equal).
  `stale-image-ifnotpresent` is lower for all four, and `hostpath-fsgroup` for
  E4B (0.67 → 0.00) and Qwen3.5 4B. Differences of one scenario's three attempts
  are not reliable.
- **The context window costs score.** 19 of 96 adapter attempts ended on the
  window (against 3 of 96 for the Qwen teacher's adapters; two more of theirs
  ended on HTTP 525 errors), 17 of them on `hostpath-fsgroup`,
  `stale-image-ifnotpresent` and `unschedulable-cpu-request`, the three scenarios
  with the largest drops. These attempts scored 0, except four
  `stale-image-ifnotpresent` attempts at 0.33.

## Stopping behaviour (question 2)

- **No adapter reaches the turn limit, and the Gemma adapters stop early less
  often.** The E2B and E4B bases ended 19 and 12 of 24 attempts with a final
  report and less than full marks; their adapters 15 and 6. The Gemma adapters work
  longer than their bases (19.2 and 18.8 turns against 11.1 and 12.0) but about
  half as long as the Qwen teacher's adapters (42.0 and 35.7), which reached the
  turn limit in 14 and 8 attempts.
- **The Qwen adapters work about half as long as their bases.** Qwen3.5 4B and
  9B used 21.5 and 16.0 turns against 41.6 and 30.1, and reached the turn limit
  in 0 attempts against 9 and 4. The Qwen teacher's adapters stayed near the
  bases' turn counts (36.5 and 30.2) and turn-limit endings (9 and 3).
- **The adapters write about twice as much per turn.** They generate 405 to 511
  tokens per turn, against 178 to 337 for the bases and 164 to 214 for the Qwen
  teacher's adapters. Fewer turns therefore still fill the window (mean peak
  context 15,000 to 21,500 tokens), which is how the attempts above ended. The
  cause was not examined.

## Held-out loss and validation

The held-out loss does not predict the validation score. These adapters end far
below their bases and the Qwen teacher's adapters on DeepSeek's solutions
(0.72 to 0.96, against 1.11 to 1.55 for the Qwen teacher's adapters), yet score
no higher than their bases and, on the point estimates, equal to or below the
Qwen teacher's adapters (next section).

## Against the Qwen3.5 9B teacher's adapters

Same scenarios and attempts per arm; difference is DeepSeek-teacher adapter minus
Qwen-teacher adapter, with the paired bootstrap of the tables above. This
comparison was not part of the method fixed before training.

| Model                   | DeepSeek minus Qwen teacher | 95% interval     |
| ----------------------- | --------------------------: | ---------------- |
| Gemma 4 E2B             |                      −0.000 | −0.146 to +0.139 |
| Gemma 4 E4B             |                      −0.042 | −0.292 to +0.250 |
| Qwen3.5 4B              |                      −0.194 | −0.472 to +0.042 |
| Qwen3.5 9B              |                      −0.028 | −0.083 to +0.000 |
| All four models, pooled |                      −0.066 | −0.168 to +0.031 |

- **No teacher is shown better.** The DeepSeek teacher's adapters score equal or
  lower for every model, but every interval includes zero.
- **Two differences besides the teacher.** The Qwen3.5 4B and 9B adapters of the
  two teachers were trained on different Python versions (3.15 without, 3.14
  with the `flash-linear-attention` kernels, see the training runs above), and
  the DeepSeek teacher's adapters lost more attempts to the context window.
  Qwen3.5 4B, the model with the largest difference, lost 8 of 24.

## Draft acceptance

The share of the multi-token-prediction head's drafted tokens that the model
accepted, over all responses of the arm (as in the Qwen teacher's results):

| Model       | Validation, base | Validation, adapter |
| ----------- | ---------------: | ------------------: |
| Gemma 4 E2B |            0.405 |               0.295 |
| Gemma 4 E4B |            0.535 |               0.400 |
| Qwen3.5 4B  |            0.615 |               0.412 |
| Qwen3.5 9B  |            0.678 |               0.456 |

Every adapter accepts fewer drafts than its base, by 0.110 to 0.222, against
0.003 to 0.051 for the Qwen teacher's adapters. The cause was not examined.

## Evidence

- [raw/run-20261126-194429-lora-deepseek-flash-data/](raw/run-20261126-194429-lora-deepseek-flash-data/):
  `session.json` and the Gemma variants' `metrics.json` and `log.txt`.
- [raw/run-20261126-230134-lora-deepseek-flash-data-qwen/](raw/run-20261126-230134-lora-deepseek-flash-data-qwen/):
  the same for the Qwen variants.
- Benchmark runs of the "Benchmark on Kaggle" workflow (`benchmark-kaggle.yaml`,
  prompt agent, 8 validation scenarios × 3, main at `85fc9a2c`), each under
  `raw/benchmark/<run id>/` as for the Qwen teacher. The bases are the runs listed in
  [results-qwen35-9b-data.md](results-qwen35-9b-data.md#evidence).
  [`results.py`](results.py) reads them.

| Workflow run | Model       | Arm     | Run                                                                         |
| ------------ | ----------- | ------- | --------------------------------------------------------------------------- |
| 45915343111  | Gemma 4 E4B | adapter | [run-2026-11-27-00-57-17-891Z](raw/benchmark/run-2026-11-27-00-57-17-891Z/) |
| 45915354059  | Qwen3.5 4B  | adapter | [run-2026-11-27-02-15-34-812Z](raw/benchmark/run-2026-11-27-02-15-34-812Z/) |
| 45915364020  | Gemma 4 E2B | adapter | [run-2026-11-27-02-51-29-827Z](raw/benchmark/run-2026-11-27-02-51-29-827Z/) |
| 45915374704  | Qwen3.5 9B  | adapter | [run-2026-11-27-04-00-17-183Z](raw/benchmark/run-2026-11-27-04-00-17-183Z/) |

All four workflow runs are marked failed and each holds all 24 attempts. No
trained-on sweep was run for this teacher's adapters.
