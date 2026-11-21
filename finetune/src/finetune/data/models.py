"""Records of the data side: a trajectory, a training row, a manifest."""

from __future__ import annotations

import json
from dataclasses import asdict, dataclass
from pathlib import Path


@dataclass(frozen=True)
class Trajectory:
    source: str
    scenario_id: str
    condition: str
    model: str
    peak_context_tokens: int
    tools: list[dict]
    messages: list[dict]


@dataclass(frozen=True)
class Candidate:
    """A clean attempt that may enter a training dataset."""

    path: Path
    run: str
    scenario_id: str
    attempt: int
    teacher: str  # model id the attempt ran on
    tokens: int  # peak context of the attempt
    writes: tuple[str, ...]  # normalised write commands, in order


@dataclass(frozen=True)
class Entry:
    """One kept attempt in a dataset manifest, with its provenance."""

    scenario: str
    teacher: str
    run: str
    attempt: int
    tokens: int
    review: str  # pending | accepted | rejected, from the review sheet
    path: str  # raw attempt record, relative to the repository


@dataclass(frozen=True)
class Row:
    """One training example: the rendered conversation and its target spans."""

    source: str  # attempt file
    scenario_id: str
    text: str
    targets: list[tuple[int, int]]  # character spans the loss is computed on

    def to_json(self) -> str:
        return json.dumps(asdict(self))

    @classmethod
    def from_json(cls, line: str) -> Row:
        raw = json.loads(line)
        return cls(**{**raw, "targets": [tuple(span) for span in raw["targets"]]})


@dataclass(frozen=True)
class Manifest:
    """Attempt files to train on and to evaluate on."""

    path: Path
    train: list[Path]
    eval: list[Path]
