"""A trajectory rendered with the checkpoint's chat template.

Training renders a whole conversation once; ``finetune.serving.templates``
checks that this equals what llama.cpp serves.
"""

from __future__ import annotations

from transformers import AutoTokenizer

from finetune.config import Family, Student
from finetune.data.models import Trajectory


def tokenizer(student: Student):
    return AutoTokenizer.from_pretrained(student.checkpoint, revision=student.revision)


def bos_text(student: Student) -> str:
    """BOS as the chat template writes it; llama.cpp adds it as a token instead."""
    return "<bos>" if student.family == Family.GEMMA else ""


def render(tok, traj: Trajectory, upto: int | None = None) -> str:
    """Full conversation, or the generation prompt before message ``upto``.

    Reasoning is on, as llama.cpp serves both families.
    """
    messages = traj.messages if upto is None else traj.messages[:upto]
    return tok.apply_chat_template(
        messages,
        tools=traj.tools,
        tokenize=False,
        add_generation_prompt=upto is not None,
        enable_thinking=True,
    )
