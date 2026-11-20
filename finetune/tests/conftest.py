"""Shared fixtures: the students' real chat templates without downloading them.

The templates in ``fixtures/`` are the ``chat_template.jinja`` files of the
pinned training checkpoints (``finetune.students``), copied unchanged; both
models of a family have the same file.
"""

import sys
from pathlib import Path

import pytest

try:
    import torch

    if not torch.cuda.is_available():
        # flash-linear-attention is Triton-only; on CPU, make transformers use
        # its PyTorch reference kernels for Qwen3.5's linear attention.
        sys.modules["fla"] = None
except ImportError:
    pass
from transformers.utils.chat_template_utils import render_jinja_template

from finetune import trajectories
from finetune.students import STUDENTS

FIXTURES = Path(__file__).parent / "fixtures"
ATTEMPT = FIXTURES / "attempt-basic-deploy-service.json"


class TemplateTokenizer:
    """The part of a tokenizer that ``finetune.render`` uses: the chat template."""

    def __init__(self, family: str):
        self.template = (FIXTURES / f"{family}.jinja").read_text()
        self.bos_token = "<bos>" if family == "gemma" else None

    def apply_chat_template(
        self, messages, tools=None, tokenize=False, add_generation_prompt=False, **kwargs
    ):
        assert not tokenize
        rendered, _ = render_jinja_template(
            conversations=[messages],
            tools=tools,
            chat_template=self.template,
            add_generation_prompt=add_generation_prompt,
            bos_token=self.bos_token or "",
            eos_token="",
            **kwargs,
        )
        return rendered[0]


@pytest.fixture(params=["gemma-4-e4b", "qwen35-4b"])
def student_name(request):
    """The two students of the study; the other models share their templates."""
    return request.param


@pytest.fixture
def student(student_name):
    return STUDENTS[student_name]


@pytest.fixture
def template_tokenizer(student):
    return TemplateTokenizer(student.family)


@pytest.fixture
def attempt():
    return trajectories.load(ATTEMPT)


# Tiny random models of both architectures (CPU tests).


def tiny_qwen(vocab):
    from transformers import Qwen3_5ForCausalLM, Qwen3_5TextConfig

    config = Qwen3_5TextConfig(
        vocab_size=vocab,
        hidden_size=32,
        intermediate_size=64,
        num_hidden_layers=4,
        num_attention_heads=2,
        num_key_value_heads=1,
        head_dim=16,
        linear_num_key_heads=2,
        linear_num_value_heads=2,
        linear_key_head_dim=8,
        linear_value_head_dim=8,
        linear_conv_kernel_dim=4,
        layer_types=["linear_attention", "linear_attention", "linear_attention", "full_attention"],
        max_position_embeddings=4096,
    )
    return Qwen3_5ForCausalLM(config)


def tiny_gemma(vocab):
    from transformers import Gemma4ForCausalLM, Gemma4TextConfig

    config = Gemma4TextConfig(
        vocab_size=vocab,
        hidden_size=32,
        intermediate_size=64,
        num_hidden_layers=2,
        num_attention_heads=2,
        num_key_value_heads=1,
        head_dim=16,
        layer_types=["sliding_attention", "full_attention"],
        sliding_window=64,
        max_position_embeddings=4096,
        final_logit_softcapping=30.0,
        hidden_size_per_layer_input=8,
        vocab_size_per_layer_input=vocab,
    )
    return Gemma4ForCausalLM(config)
