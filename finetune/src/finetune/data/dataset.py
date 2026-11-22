"""One teacher's training dataset: every attempt of its runs, and a manifest
that lists the kept ones with provenance.

    uv run python -m finetune.data.dataset --name deepseek-flash \\
        --runs <run dir>...

The run records (every attempt, kept or not, and each run's ``run.json`` and
``results.json``) are copied next to the manifest, so the dataset is
self-contained and the selection can be redone (another cap, another filter)
without the benchmark results. The manifest's ``train`` list is what
``finetune.data.manifest`` reads.
"""

from __future__ import annotations

import argparse
import logging
import shutil
from dataclasses import asdict
from pathlib import Path

import yaml

from finetune import logs
from finetune.config import CONFIG, ROOT
from finetune.data import review, select
from finetune.data.models import Candidate, Entry

log = logging.getLogger(__name__)

DATASETS = ROOT / "finetune" / "datasets"
SCENARIOS = ROOT / "benchmark" / "scenarios"


def development_scenarios() -> set[str]:
    return {p.parent.name for p in SCENARIOS.glob("*/*/scenario.yaml")}


def training_scenarios() -> set[str]:
    return development_scenarios() - CONFIG.validation_scenarios


def raw_target(c: Candidate, out: Path) -> Path:
    return out / "runs" / c.run / c.scenario_id / "prompt" / c.path.name


def _copy(source: Path, target: Path) -> None:
    """Copy unless the run is already the dataset's own, so a build can be redone
    from the records it keeps."""
    if source.resolve() == target.resolve():
        return
    target.parent.mkdir(parents=True, exist_ok=True)
    shutil.copyfile(source, target)


def copy_runs(run_dirs: list[Path], allowed: set[str], out: Path) -> None:
    """Every attempt record of the training scenarios, and each run's metadata."""
    for run in run_dirs:
        for name in ("run.json", "results.json"):
            _copy(run / name, out / "runs" / run.name / name)
        for path in select.attempt_files([run]):
            if path.parts[-3] in allowed:
                _copy(path, out / "runs" / run.name / path.parts[-3] / "prompt" / path.name)


def entries(kept: list[Candidate], out: Path) -> list[Entry]:
    """Manifest rows. Every attempt is reviewed (D-054): its verdict, once given
    in its sheet, is kept, and an attempt without one is pending."""
    done = review.verdicts(out / "review")
    rows = []
    for c in kept:
        path = select.relative(raw_target(c, out))
        status = done[path].verdict if path in done else review.PENDING
        rows.append(Entry(c.scenario_id, c.teacher, c.run, c.attempt, c.tokens, status, path))
    return sorted(rows, key=lambda e: e.path)


def rejected(out: Path) -> set[str]:
    return {
        v.attempt for v in review.verdicts(out / "review").values() if v.verdict == review.REJECTED
    }


def build(name: str, run_dirs: list[Path]) -> Path:
    out = DATASETS / name
    allowed = training_scenarios()
    copy_runs(run_dirs, allowed, out)
    clean = []
    for path in select.attempt_files(run_dirs):
        found = select.candidate(path)
        if found and found.scenario_id in allowed:
            clean.append(found)
    dropped = rejected(out)
    clean = [c for c in clean if select.relative(raw_target(c, out)) not in dropped]
    kept = select.select(clean, CONFIG.dataset.max_per_scenario)
    rows = entries(kept, out)
    review.write_missing(out / "review", rows)
    missing = sorted(allowed - {e.scenario for e in rows})
    manifest = {
        "teacher": name,
        "train": [e.path for e in rows],
        "scenarios_without_examples": missing,
        "attempts": [asdict(e) for e in rows],
    }
    (out / "manifest.yaml").write_text(yaml.safe_dump(manifest, sort_keys=False, width=200))
    log.info("%s: %d attempts kept, %d scenarios without examples", name, len(rows), len(missing))
    return out


def main() -> None:
    logs.setup()
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--name", required=True, help="dataset directory under finetune/datasets")
    parser.add_argument("--runs", nargs="+", type=Path, required=True)
    args = parser.parse_args()
    build(args.name, args.runs)


if __name__ == "__main__":
    main()
