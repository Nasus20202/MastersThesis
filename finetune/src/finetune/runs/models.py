"""Records of a run: which variants, from which commit and data."""

from __future__ import annotations

from dataclasses import dataclass

from finetune.config import STUDENTS, Training
from finetune.training.models import Precision


@dataclass(frozen=True)
class Variant:
    """A model and a precision, named ``<student>-<precision>``."""

    student: str
    precision: Precision

    @classmethod
    def parse(cls, name: str) -> Variant:
        student, _, precision = name.rpartition("-")
        if student not in STUDENTS or precision not in set(Precision):
            raise ValueError(f"unknown variant {name}")
        return cls(student, Precision(precision))

    @property
    def name(self) -> str:
        return f"{self.student}-{self.precision}"


VARIANTS = tuple(Variant(s, p) for s in STUDENTS for p in Precision)


@dataclass(frozen=True)
class Provenance:
    commit: str
    uncommitted_changes: bool


@dataclass(frozen=True)
class Session:
    """``session.json``: what a run was trained from."""

    provenance: Provenance
    manifest: str  # relative to the repository
    variants: list[str]
    gpu: str | None
    steps: int
    training: Training
    data: dict[str, str]  # data file -> SHA-256
