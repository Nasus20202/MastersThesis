"""Frozen dataset rows: the examples a freeze recorded, with the revision that
covers them.

Training reads these rows instead of rendering the attempt records again, so a
cut or a masked prefix cannot be lost on the way to the GPU. The revision is
checked first: a row file that no longer matches its record means the dataset
changed after it was reviewed, and the run is refused.
"""

from __future__ import annotations

import hashlib
import json
from dataclasses import dataclass
from pathlib import Path

from finetune.config import ROOT
from finetune.data.models import Row
from finetune.data.rows import read_rows


def record(directory: Path) -> Path:
    """The freeze record of a dataset, which names the files its revision covers."""
    found = directory / "freeze.json"
    if not found.is_file():
        raise ValueError(f"{directory} holds no freeze record")
    return found


def digest(files: list[Path]) -> str:
    return hashlib.sha256(b"".join(p.read_bytes() for p in files)).hexdigest()


@dataclass(frozen=True)
class Dataset:
    """A frozen dataset: the row file of each student and the revision."""

    directory: Path
    revision: str
    rows: dict[str, list[Row]]  # student -> rows, as frozen


def read(directory: Path) -> Dataset:
    """The rows of a frozen dataset, after checking they match its revision."""
    path = record(directory)
    saved = json.loads(path.read_text())
    if "files" in saved:  # a variant records the files its revision covers
        files = [ROOT / name for name in saved["files"]]
    else:  # a teacher dataset: its manifest and renders (finetune.data.freeze)
        files = [directory / "manifest.yaml", *sorted((directory / "rendered").glob("*.jsonl"))]
    if digest(files) != saved["revision"]:
        raise ValueError(f"{path} no longer matches its rows: freeze {directory} again")
    rows = {p.stem: read_rows(p) for p in sorted((directory / "rendered").glob("*.jsonl"))}
    if not rows:
        raise ValueError(f"{directory} has no rendered rows")
    return Dataset(directory, saved["revision"], rows)
