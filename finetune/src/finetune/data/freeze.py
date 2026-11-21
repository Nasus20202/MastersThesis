"""Freezes a reviewed dataset: renders it for each student, counts it and
hashes it (#25).

    uv run python -m finetune.data.freeze --name deepseek-flash

The hash of the manifest and the rendered files is the dataset revision.
Trajectories longer than ``training.max_len`` tokens for a student are left out
of that student's file, never cut, and listed in ``freeze.json``.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import logging
from collections import Counter
from dataclasses import asdict, dataclass
from pathlib import Path

import yaml

from finetune import logs
from finetune.config import CONFIG, ROOT
from finetune.data import dataset, render, review, rows, trajectories

log = logging.getLogger(__name__)


@dataclass(frozen=True)
class Rendered:
    student: str
    examples: int
    over_length: list[str]  # raw records left out: longer than max_len tokens


@dataclass(frozen=True)
class Freeze:
    revision: str  # sha256 of the manifest and the rendered files
    teacher: str
    attempts: int
    per_scenario: dict[str, int]
    per_area: dict[str, int]
    scenarios_without_examples: list[str]
    rendered: list[Rendered]


def areas() -> dict[str, str]:
    return {
        p.parent.name: p.parent.parent.name for p in dataset.SCENARIOS.glob("*/*/scenario.yaml")
    }


def render_student(directory: Path, name: str, paths: list[str], tok) -> Rendered:
    student = CONFIG.students[name]
    folder = directory / "rendered"
    folder.mkdir(exist_ok=True)
    kept, over = 0, []
    with (folder / f"{name}.jsonl").open("w") as f:
        for path in paths:
            row = rows.example(tok, student, trajectories.load(ROOT / path))
            if len(tok(row.text, add_special_tokens=False)["input_ids"]) > CONFIG.training.max_len:
                over.append(path)
                continue
            f.write(rows.replace_source(row, path).to_json() + "\n")
            kept += 1
    log.info("%s: %d examples, %d over %d tokens", name, kept, len(over), CONFIG.training.max_len)
    return Rendered(name, kept, over)


def digest(directory: Path) -> str:
    files = [directory / "manifest.yaml", *sorted((directory / "rendered").glob("*.jsonl"))]
    return hashlib.sha256(b"".join(p.read_bytes() for p in files)).hexdigest()


def freeze(directory: Path, tokenizers: dict | None = None) -> Freeze:
    manifest = yaml.safe_load((directory / "manifest.yaml").read_text())
    pending = [a["path"] for a in manifest["attempts"] if a["review"] == review.PENDING]
    if pending:
        raise ValueError(f"{len(pending)} attempts still wait for review, e.g. {pending[0]}")
    rejected = [a["path"] for a in manifest["attempts"] if a["review"] == review.REJECTED]
    if rejected:
        raise ValueError(
            f"rejected attempts are in the manifest: build the dataset again: {rejected}"
        )
    tokenizers = tokenizers or {n: render.tokenizer(s) for n, s in CONFIG.students.items()}
    rendered = [
        render_student(directory, name, manifest["train"], tok) for name, tok in tokenizers.items()
    ]
    scenarios = Counter(a["scenario"] for a in manifest["attempts"])
    area_of = areas()
    result = Freeze(
        revision=digest(directory),
        teacher=manifest["teacher"],
        attempts=len(manifest["attempts"]),
        per_scenario=dict(sorted(scenarios.items())),
        per_area=dict(sorted(Counter(area_of[s] for s in scenarios.elements()).items())),
        scenarios_without_examples=manifest["scenarios_without_examples"],
        rendered=rendered,
    )
    (directory / "freeze.json").write_text(json.dumps(asdict(result), indent=2) + "\n")
    log.info("%s frozen: revision %s", result.teacher, result.revision)
    return result


def main() -> None:
    logs.setup()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--name", required=True)
    freeze(dataset.DATASETS / parser.parse_args().name)


if __name__ == "__main__":
    main()
