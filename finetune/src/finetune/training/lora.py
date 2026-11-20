"""The checkpoint with a fresh LoRA adapter, on the GPU.

Weights and autocast are bfloat16; ``qlora`` stores the decoder projections
in 4-bit NF4 (bitsandbytes) and the norms in float32. On the CPU (tests)
everything is float32.
"""

from __future__ import annotations

import logging

import torch

from finetune.config import CONFIG
from finetune.training.gpu import allocated_gib, compute_dtype
from finetune.training.models import Precision, Settings

log = logging.getLogger(__name__)


LORA_MODULES = tuple(CONFIG.training.lora_modules)


# Language-model projections only: vision and audio towers stay frozen and
# without adapters (the served GGUF has no towers).
LORA_TARGET_REGEX = r"^(?!.*(?:vision|audio|visual)).*\.(?:" + "|".join(LORA_MODULES) + r")$"
# Left in 16-bit by the 4-bit quantization: the output head and the towers.
NOT_QUANTIZED = ["lm_head", "vision", "audio", "visual"]


def load(
    checkpoint_path: str,
    revision: str | None,
    cfg: Settings,
    dtype: torch.dtype | None = None,
):
    """The model with a fresh LoRA adapter, and its tokenizer. Weights are in
    ``compute_dtype()`` unless ``dtype`` is given."""
    from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
    from transformers import AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig

    dtype = dtype or compute_dtype()
    kwargs: dict = {"revision": revision, "dtype": dtype}
    if cfg.precision == Precision.QLORA:
        kwargs["quantization_config"] = BitsAndBytesConfig(
            load_in_4bit=True,
            bnb_4bit_quant_type="nf4",
            bnb_4bit_use_double_quant=True,
            bnb_4bit_compute_dtype=dtype,
            llm_int8_skip_modules=NOT_QUANTIZED,
        )
    if torch.cuda.is_available():
        kwargs["device_map"] = "cuda"
    log.info("loading %s (%s, %s)", checkpoint_path, cfg.precision, dtype)
    model = AutoModelForCausalLM.from_pretrained(checkpoint_path, **kwargs)
    log.info("loaded; GPU memory %.1f GiB", allocated_gib())
    if cfg.precision == Precision.QLORA:
        model = prepare_model_for_kbit_training(
            model,
            use_gradient_checkpointing=True,
            gradient_checkpointing_kwargs={"use_reentrant": False},
        )
        frozen_matrices_to(model, dtype)
    model.gradient_checkpointing_enable(gradient_checkpointing_kwargs={"use_reentrant": False})
    model.enable_input_require_grads()
    model = get_peft_model(
        model,
        LoraConfig(
            r=cfg.rank,
            lora_alpha=cfg.alpha,
            lora_dropout=0.0,
            bias="none",
            target_modules=LORA_TARGET_REGEX,
            task_type="CAUSAL_LM",
        ),
    )
    tok = AutoTokenizer.from_pretrained(checkpoint_path, revision=revision)
    return model, tok


def frozen_matrices_to(model, dtype: torch.dtype) -> None:
    """Casts frozen float32 matrices (embeddings, output head, towers) to
    ``dtype``. The 4-bit preparation casts every weight that is not 4-bit to
    float32; only the small 1-D weights (norms, biases) need it, and Qwen3.5
    9B's embeddings and head alone would take 8 GB in float32."""
    for p in model.parameters():
        if not p.requires_grad and p.dtype == torch.float32 and p.dim() >= 2:
            p.data = p.data.to(dtype)


def input_device(model) -> torch.device:
    return model.get_input_embeddings().weight.device


def adapted_modules(model) -> list[str]:
    return sorted(
        {n.rsplit(".lora_A", 1)[0] for n, _ in model.named_parameters() if ".lora_A" in n}
    )


def trainable_parameters(model) -> int:
    return sum(p.numel() for p in model.parameters() if p.requires_grad)
