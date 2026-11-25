"""A training run's inputs: the attempts of a manifest, rendered for each model.

A manifest (YAML) lists benchmark attempt files to train on (``train``) and
to evaluate on (``eval``), as paths relative to the repository. Training
attempts on validation scenarios are refused (D-055).

With ``rows`` the training split is a frozen dataset instead (truncation or
continuation): its rows are copied as they were reviewed, and the manifest
supplies only the evaluation split, so every variant is compared on the same
validation attempts.

``prepare`` writes to ``out``:

- ``data/<student>/{train,eval}.jsonl``: the rendered rows;
- ``session.json``: repository commit, manifest, variants, GPU, settings and
  the SHA-256 of every data file.

Training itself is ``finetune.training.run``, one variant (model x precision) per
GPU; on Modal through ``finetune.modal``.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import logging
import subprocess
from collections.abc import Iterable, Iterator
from dataclasses import asdict, replace
from pathlib import Path

from finetune import logs
from finetune.config import CONFIG, ROOT, STUDENTS, Student
from finetune.data import frozen, render, rows, trajectories
from finetune.data import manifest as manifests
from finetune.data.models import Manifest, Row
from finetune.runs.models import VARIANTS, Provenance, Session, Variant
from finetune.training.models import Job, Settings

log = logging.getLogger(__name__)


def _write(path: Path, examples: Iterable[Row]) -> None:
    with path.open("w") as f:
        for row in examples:
            f.write(row.to_json() + "\n")


def rendered(tok, student: Student, paths: list[Path]) -> Iterator[Row]:
    for path in paths:
        row = rows.example(tok, student, trajectories.load(path))
        yield replace(row, source=str(path.relative_to(ROOT)))


def render_data(
    out: Path, students: list[str], manifest: Manifest, dataset: frozen.Dataset | None = None
) -> None:
    """The training and evaluation rows of each student.

    Without ``dataset`` the training split is rendered from the manifest; with
    it, the frozen rows are written unchanged and the mask or cut stays.
    """
    for name in students:
        student = STUDENTS[name]
        tok = render.tokenizer(student)
        folder = out / name
        folder.mkdir(parents=True, exist_ok=True)
        if dataset is None:
            _write(folder / "train.jsonl", rendered(tok, student, manifest.train))
        else:
            if name not in dataset.rows:
                raise ValueError(f"no {name} rows in {dataset.directory}")
            _write(folder / "train.jsonl", dataset.rows[name])
        _write(folder / "eval.jsonl", rendered(tok, student, manifest.eval))


def job(variant: Variant, data: Path, out: Path, steps: int | None = None) -> Job:
    """The training job of one variant on ``prepare``'s data."""
    return Job(
        settings=Settings(precision=variant.precision),
        data=data / variant.student / "train.jsonl",
        eval=data / variant.student / "eval.jsonl",
        out=out,
        student=variant.student,
        steps=steps,
    )


def repository_path(path: Path) -> str:
    """``path`` relative to the repository when it lies inside it."""
    try:
        return str(path.resolve().relative_to(ROOT))
    except ValueError:
        return str(path)


def provenance() -> Provenance:
    """The checked-out commit and whether the working tree differs from it."""

    def git(*args: str) -> str:
        return subprocess.run(
            ["git", *args], cwd=ROOT, capture_output=True, text=True, check=True
        ).stdout.strip()

    return Provenance(git("rev-parse", "HEAD"), bool(git("status", "--porcelain")))


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def prepare(
    out: Path,
    manifest: Manifest,
    variants: list[Variant],
    steps: int | None = None,
    gpu: str | None = None,
    rows: Path | None = None,
) -> Session:
    """The rows of ``variants`` and ``session.json``; ``rows`` is a frozen dataset."""
    data = out / "data"
    dataset = frozen.read(rows) if rows is not None else None
    if dataset is None:
        log.info(
            "rendering %d training and %d evaluation attempts",
            len(manifest.train),
            len(manifest.eval),
        )
    else:
        log.info(
            "training rows from %s (revision %s), %d evaluation attempts",
            dataset.directory,
            dataset.revision[:12],
            len(manifest.eval),
        )
    render_data(data, sorted({v.student for v in variants}), manifest, dataset)
    session = Session(
        provenance=provenance(),
        manifest=str(manifest.path.relative_to(ROOT)),
        data_source=repository_path(rows) if rows is not None else None,
        data_revision=dataset.revision if dataset else None,
        variants=[v.name for v in variants],
        gpu=gpu,
        steps=steps,
        training=CONFIG.training,
        data={str(p.relative_to(data)): sha256(p) for p in sorted(data.rglob("*.jsonl"))},
    )
    (out / "session.json").write_text(json.dumps(asdict(session), indent=2))
    return session


def main(argv: list[str] | None = None) -> None:
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("--manifest", type=Path, required=True)
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument(
        "--variants",
        nargs="+",
        choices=[v.name for v in VARIANTS],
        help="default: the manifest's, or all",
    )
    parser.add_argument("--steps", type=int, help="a fixed step count instead of the epochs")
    args = parser.parse_args(argv)
    logs.setup()
    manifest = manifests.read(args.manifest)
    variants = [Variant.parse(v) for v in args.variants or variant_names(manifest)]
    prepare(args.out, manifest, variants, args.steps, rows=manifest.rows)


def variant_names(manifest: Manifest) -> list[str]:
    """The manifest's variants, or every variant when it names none."""
    return manifest.variants or [v.name for v in VARIANTS]


if __name__ == "__main__":
    main()
