"""A training run's inputs: the attempts of a manifest, rendered for each model.

A manifest (YAML) lists benchmark attempt files to train on (``train``) and
to evaluate on (``eval``), as paths relative to the repository. Training
attempts on validation scenarios are refused (D-055).

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
from dataclasses import asdict, replace
from pathlib import Path

from finetune import logs
from finetune.config import CONFIG, ROOT, STUDENTS
from finetune.data import manifest as manifests
from finetune.data import render, rows, trajectories
from finetune.data.models import Manifest
from finetune.runs.models import VARIANTS, Provenance, Session, Variant
from finetune.training.models import Job, Settings

log = logging.getLogger(__name__)


def render_data(out: Path, students: list[str], manifest: Manifest) -> None:
    for name in students:
        student = STUDENTS[name]
        tok = render.tokenizer(student)
        folder = out / name
        folder.mkdir(parents=True, exist_ok=True)
        for split, paths in (("train", manifest.train), ("eval", manifest.eval)):
            with (folder / f"{split}.jsonl").open("w") as f:
                for path in paths:
                    row = rows.example(tok, student, trajectories.load(path))
                    row = replace(row, source=str(path.relative_to(ROOT)))
                    f.write(row.to_json() + "\n")


def job(variant: Variant, data: Path, out: Path, steps: int) -> Job:
    """The training job of one variant on ``prepare``'s data."""
    return Job(
        settings=Settings(precision=variant.precision),
        data=data / variant.student / "train.jsonl",
        eval=data / variant.student / "eval.jsonl",
        out=out,
        student=variant.student,
        steps=steps,
    )


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
    out: Path, manifest: Manifest, variants: list[Variant], steps: int, gpu: str | None = None
) -> Session:
    """Renders the manifest for ``variants`` and writes ``session.json``."""
    data = out / "data"
    log.info(
        "rendering %d training and %d evaluation attempts", len(manifest.train), len(manifest.eval)
    )
    render_data(data, sorted({v.student for v in variants}), manifest)
    session = Session(
        provenance=provenance(),
        manifest=str(manifest.path.relative_to(ROOT)),
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
        default=[v.name for v in VARIANTS],
        help="default: all",
    )
    parser.add_argument("--steps", type=int, default=CONFIG.training.steps)
    args = parser.parse_args(argv)
    logs.setup()
    variants = [Variant.parse(v) for v in args.variants]
    prepare(args.out, manifests.read(args.manifest), variants, args.steps)


if __name__ == "__main__":
    main()
