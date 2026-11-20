"""Models that can be fine-tuned, and the checkpoints they are trained from (#24).

The study fine-tunes two students (D-046): Gemma 4 E4B and Qwen3.5 4B. The
technical check also trains the benchmark's other two local models, Gemma 4
E2B and Qwen3.5 9B, to show what fits on Kaggle's T4s; they are not in the
study.

Each training checkpoint is the base model that the served GGUF's model card
declares, at a pinned revision; the served files are pinned in
``benchmark/model-profiles``. Within a family the chat templates are
identical, so rendering and target spans depend only on the family.
"""

from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Student:
    family: str  # "gemma" or "qwen"
    checkpoint: str
    revision: str
    served_model: str  # llama.cpp model name; the GGUF file is ``<name>.gguf``
    in_study: bool  # one of D-046's two students

    @property
    def served_gguf(self) -> str:
        return f"{self.served_model}.gguf"


STUDENTS = {
    "gemma-4-e2b": Student(
        family="gemma",
        checkpoint="google/gemma-4-E2B-it-qat-q4_0-unquantized",
        revision="6befbaca7398925921802abd1f277b495b78b738",
        served_model="gemma-4-E2B-it-qat-UD-Q4_K_XL",
        in_study=False,
    ),
    "gemma-4-e4b": Student(
        family="gemma",
        checkpoint="google/gemma-4-E4B-it-qat-q4_0-unquantized",
        revision="476025a01dbf99361c062bbeca3d6a76bb4c4566",
        served_model="gemma-4-E4B-it-qat-UD-Q4_K_XL",
        in_study=True,
    ),
    "qwen35-4b": Student(
        family="qwen",
        checkpoint="Qwen/Qwen3.5-4B",
        revision="851bf6e806efd8d0a36b00ddf55e13ccb7b8cd0a",
        served_model="Qwen3.5-4B-Q4_K_M",
        in_study=True,
    ),
    "qwen35-9b": Student(
        family="qwen",
        checkpoint="Qwen/Qwen3.5-9B",
        revision="c202236235762e1c871ad0ccb60c8ee5ba337b9a",
        served_model="Qwen3.5-9B-Q4_K_M",
        in_study=False,
    ),
}
