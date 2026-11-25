"""Records of a training run: the job, its settings and what it measured."""

from __future__ import annotations

import json
from dataclasses import asdict, dataclass, field
from enum import StrEnum
from pathlib import Path

from finetune.config import CONFIG

TRAINING = CONFIG.training


class Precision(StrEnum):
    LORA = "lora"  # 16-bit frozen weights
    QLORA = "qlora"  # 4-bit NF4 frozen decoder projections


@dataclass(frozen=True)
class Settings:
    """Training values (``config.yaml`` defaults)."""

    precision: Precision
    rank: int = TRAINING.rank
    alpha: int = TRAINING.alpha
    learning_rate: float = TRAINING.learning_rate
    warmup_ratio: float = TRAINING.warmup_ratio
    lr_schedule: str = TRAINING.lr_schedule
    weight_decay: float = TRAINING.weight_decay
    max_grad_norm: float = TRAINING.max_grad_norm
    seed: int = TRAINING.seed


@dataclass(frozen=True)
class Job:
    """One training run: a model (``student`` or a local ``checkpoint``), its
    data and its settings."""

    settings: Settings
    data: Path  # training rows (``finetune.data.rows``)
    out: Path
    student: str | None = None
    checkpoint: Path | None = None  # local model directory, instead of a student (tests)
    eval: Path | None = None  # held-out rows for the evaluation loss
    epochs: int = TRAINING.epochs
    steps: int | None = None  # replaces ``epochs`` with a fixed step count (POC, tests)
    max_len: int = TRAINING.max_len
    sweep: tuple[int, ...] = tuple(TRAINING.sweep)


@dataclass(frozen=True)
class Tokens:
    """A tokenized example; labels are IGNORE outside target spans."""

    ids: list[int]
    labels: list[int]
    source: str = ""  # the attempt file the row was rendered from

    def window(self, length: int) -> Tokens:
        return Tokens(self.ids[:length], self.labels[:length], self.source)


@dataclass(frozen=True)
class Step:
    """One training step."""

    loss: float
    tokens: int
    target_tokens: int
    seconds: float
    finite: bool
    grad_norm: float | None = None  # before clipping
    step: int = 0  # position in the run; 0 in the sweep


class SweepStatus(StrEnum):
    OK = "ok"
    OUT_OF_MEMORY = "out of memory"
    NOT_ENOUGH_TOKENS = "not enough tokens"  # the training data is shorter


@dataclass(frozen=True)
class SweepPoint:
    """One sequence length of the sweep."""

    length: int
    status: SweepStatus
    peak_gib: float | None = None
    seconds_per_step: float | None = None
    tokens_per_second: float | None = None
    loss: float | None = None
    error: str | None = None


@dataclass(frozen=True)
class ExampleLoss:
    """Mean target-token loss of one held-out example."""

    source: str  # its attempt file; the path names the teacher
    target_tokens: int
    with_adapter: float
    without_adapter: float


@dataclass(frozen=True)
class Evaluation:
    """Mean target-token loss on held-out examples, and each example's own, so
    the loss can be split by teacher or scenario afterwards."""

    examples: int
    with_adapter: float
    without_adapter: float
    per_example: list[ExampleLoss] = field(default_factory=list)


@dataclass(frozen=True)
class EpochEvaluation:
    """The held-out loss after one epoch; the last epoch is ``eval_after``."""

    epoch: int
    step: int
    evaluation: Evaluation


@dataclass(frozen=True)
class Gpu:
    name: str
    memory_gib: float


@dataclass(frozen=True)
class Environment:
    python: str
    packages: dict[str, str]  # distribution name -> version
    dtype: str
    gpu: Gpu | None  # None on the CPU
    cuda: str | None


@dataclass(frozen=True)
class ExampleCount:
    total: int
    kept: int  # not longer than the cap
    tokens: list[int]  # length of every example, sorted


@dataclass
class Metrics:
    """``metrics.json``: filled phase by phase and saved after each."""

    job: Job
    environment: Environment
    attention: str | None = None
    load_seconds: float | None = None
    host_peak_gib: float | None = None  # peak resident host memory of the run
    load_peak_gib: float | None = None
    weights_gib: float | None = None
    parameter_dtypes: dict[str, int] = field(default_factory=dict)  # dtype -> parameters
    trainable_parameters: int = 0
    adapted_modules: list[str] = field(default_factory=list)
    sweep: list[SweepPoint] = field(default_factory=list)
    max_len: int | None = None
    train_examples: ExampleCount | None = None
    steps: int | None = None  # training steps planned: epochs x kept examples, or --steps
    eval_examples: ExampleCount | None = None
    eval_before: Evaluation | None = None
    train_log: list[Step] = field(default_factory=list)
    train_error: str | None = None  # out of memory during training
    train_peak_gib: float | None = None
    train_tokens_per_second: float | None = None
    train_seconds: float | None = None
    eval_epochs: list[EpochEvaluation] = field(default_factory=list)  # all but the last
    eval_after: Evaluation | None = None
    error: str | None = None  # the exception that ended the run

    def save(self) -> None:
        data = json.dumps(asdict(self), indent=2, default=str)
        (self.job.out / "metrics.json").write_text(data)
