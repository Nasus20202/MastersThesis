"""Settings from ``config.yaml``: Modal resources, pinned versions, training
values and the model checkpoints. ``FINETUNE_CONFIG`` points to another file
(the Modal containers use it)."""

from __future__ import annotations

import os
from dataclasses import dataclass
from enum import StrEnum
from pathlib import Path

import yaml

PATH = Path(os.environ.get("FINETUNE_CONFIG", Path(__file__).resolve().parents[2] / "config.yaml"))
ROOT = PATH.resolve().parents[1]  # the repository (locally)


@dataclass(frozen=True)
class Modal:
    app: str
    gpu: str
    cpu: float
    memory_mib: int
    timeout_minutes: int
    uv: str
    results_volume: str
    cache_volume: str


@dataclass(frozen=True)
class LlamaCpp:
    image: str
    cpu: str  # tag@digest
    vulkan: str

    def ref(self, gpu: bool) -> str:
        return f"{self.image}:{self.vulkan if gpu else self.cpu}"


@dataclass(frozen=True)
class Training:
    rank: int
    alpha: int
    epochs: int
    learning_rate: float
    warmup_ratio: float
    lr_schedule: str
    weight_decay: float
    max_grad_norm: float
    seed: int
    max_len: int
    sweep: list[int]
    lora_modules: list[str]


@dataclass(frozen=True)
class Dataset:
    max_per_scenario: int


class Family(StrEnum):
    """Model family; chat template and target spans depend only on it."""

    GEMMA = "gemma"
    QWEN = "qwen"


@dataclass(frozen=True)
class Student:
    family: Family
    checkpoint: str
    revision: str
    served_model: str  # llama.cpp model name; the GGUF file is ``<name>.gguf``

    @property
    def served_gguf(self) -> str:
        return f"{self.served_model}.gguf"


@dataclass(frozen=True)
class Config:
    modal: Modal
    llama_cpp: LlamaCpp
    training: Training
    dataset: Dataset
    validation_scenarios: frozenset[str]  # never trained on (D-055)
    students: dict[str, Student]


def load(path: Path = PATH) -> Config:
    raw = yaml.safe_load(path.read_text())
    return Config(
        modal=Modal(**raw["modal"]),
        llama_cpp=LlamaCpp(**raw["llama_cpp"]),
        training=Training(**raw["training"]),
        dataset=Dataset(**raw["dataset"]),
        validation_scenarios=frozenset(raw["validation_scenarios"]),
        students={
            name: Student(**{**s, "family": Family(s["family"])})
            for name, s in raw["students"].items()
        },
    )


CONFIG = load()
STUDENTS = CONFIG.students
