# finetune

LoRA and QLoRA fine-tuning of the benchmark's local models on agent
trajectories, with Hugging Face transformers + PEFT (D-052), trained on Modal
(D-055). Settings are in [config.yaml](config.yaml); rules for changing the
module are in [AGENTS.md](AGENTS.md). The GPU POC is planned in #24;
its attempts are listed in [manifests/poc.yaml](manifests/poc.yaml).

## Flow

The package has one subpackage per stage; each keeps its dataclasses in
`models.py` and its logic in the other modules.

1. **Data** ([`data/`](src/finetune/data/)): a benchmark attempt becomes a
   chat conversation (`trajectories`), rendered once with the checkpoint's
   chat template (`render`), with the text the model generates when serving
   marked as the loss target (`rows`). A manifest lists the attempts to train
   and evaluate on (`manifest`); training on validation scenarios is refused.
2. **Run inputs** ([`runs/prepare.py`](src/finetune/runs/prepare.py)): the
   manifest rendered for each model, with the commit and data hashes.
3. **Training** ([`training/`](src/finetune/training/)): one model, one
   precision (`run`): the checkpoint with a LoRA adapter on one GPU
   (`lora`), a memory and speed sweep (`sweep`), evaluation loss before and
   after (`evaluate`), training steps (`trainer`) with a chunked loss
   (`loss`), all in bfloat16 (`gpu`).
4. **Modal** ([`modal.py`](src/finetune/modal.py)): steps 2 and 3, one
   container and GPU per variant, results in a Modal volume.
5. **Serving** ([`serving/`](src/finetune/serving/)): the training render
   equals the served prompt (`templates`); the adapter is converted to GGUF
   and applied to the served 4-bit file (`llama`).

## Commands

```sh
make finetune-test        # CPU tests with tiny random models
make finetune-lint

# Modal, from finetune/ (once: `uv run modal setup`). Costs money: only with
# the researcher's approval.
uv run modal run -m finetune.modal --manifest manifests/poc.yaml \
  --variants qwen35-4b-lora,gemma-4-e4b-lora --gpu H100
uv run modal volume get finetune-results <run> <dir>

# Served perplexity of one adapter (local Vulkan GPU; needs the served GGUF
# in benchmark/models)
cd finetune && uv run --locked python -m finetune.serving.llama \
  --student gemma-4-e4b --adapter <run>/gemma-4-e4b-lora/adapter \
  --eval <run>/data/gemma-4-e4b/eval.jsonl --out <dir>
```
