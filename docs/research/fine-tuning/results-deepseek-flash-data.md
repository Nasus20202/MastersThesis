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

<!-- TODO: 8 validation scenarios × 3, prompt agent; the bases are the runs
listed in results-qwen35-9b-data.md. -->

## Evidence

- [raw/run-20261126-194429-lora-deepseek-flash-data/](raw/run-20261126-194429-lora-deepseek-flash-data/):
  `session.json` and the Gemma variants' `metrics.json` and `log.txt`.
- [raw/run-20261126-230134-lora-deepseek-flash-data-qwen/](raw/run-20261126-230134-lora-deepseek-flash-data-qwen/):
  the same for the Qwen variants.

<!-- TODO: the benchmark runs. -->
