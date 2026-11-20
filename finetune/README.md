# finetune

LoRA and QLoRA fine-tuning of the local models on benchmark agent trajectories
with Hugging Face transformers + PEFT (D-054). Results and the reasoning behind
the settings are in the
[technical check](../docs/research/finetune-technical-check/README.md).

## Flow

1. **Trajectories** ([`trajectories.py`](src/finetune/trajectories.py)): a
   benchmark attempt becomes a chat conversation; only clean attempts (full
   score, ended by the agent) are used, never the validation scenarios.
2. **Rendering** ([`render.py`](src/finetune/render.py),
   [`dataset.py`](src/finetune/dataset.py)): the whole conversation is rendered
   once with the checkpoint's chat template, and the text the model generates
   when serving is marked as the loss target. `render.compare` checks the
   render against the prompt llama.cpp builds when serving.
3. **Training** ([`train.py`](src/finetune/train.py),
   [`placement.py`](src/finetune/placement.py)): one model, one precision; the
   model is split over all GPUs; a memory sweep picks the longest sequence
   that fits; evaluation loss before and after; the adapter is saved.
4. **Served model** ([`llama.py`](src/finetune/llama.py),
   [`serving.py`](src/finetune/serving.py)): the adapter is converted to GGUF
   and applied to the served 4-bit file with llama.cpp (perplexity, router,
   multi-token-prediction drafter).

[`check.py`](src/finetune/check.py) runs steps 1 to 3 for every model and
precision on Kaggle's 2x T4; [`kaggle/check/kernel.py`](kaggle/check/kernel.py)
starts it from a clone of the repository.

## Commands

```sh
make finetune-test        # CPU tests with tiny random models
make finetune-lint

# Kaggle (GPU T4 x2): push the starter, then download its output
kaggle kernels push -p finetune/kaggle/check
kaggle kernels output krzysztofnasuta/finetune-check -p <dir>

# Served perplexity of one adapter (pinned llama.cpp image, local Vulkan GPU;
# needs the served GGUF in benchmark/models)
cd finetune && uv run --locked python -m finetune.llama \
  --student gemma-4-e4b --adapter <run>/gemma-4-e4b-lora/adapter \
  --eval <run>/data/gemma-4-e4b/eval.jsonl --out <dir>
```
