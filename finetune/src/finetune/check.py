"""The technical check (#24) on a Kaggle 2x T4 machine, from a clone of this repository.

Renders the training and evaluation trajectories for every model, then runs
``finetune.train`` once per variant (model x precision), each in its own
process with both GPUs. Writes to ``--out``:

- ``session.json``: repository commit, GPUs, settings, the selected attempts
  and the SHA-256 of every rendered data file;
- ``data/<student>/{train,eval}.jsonl``: the rendered rows;
- ``<variant>/``: the training outputs (``finetune.train``), ``log.txt`` and
  ``status.json``.

Selection (fixed before any run): Qwen3.5 9B prompt-agent attempts of the
October suite's main run that are clean (full score, ended by the agent).
Training: those outside the validation scenarios, sorted by scenario, every
third one. Evaluation: those on validation scenarios (never trained on).
"""

from __future__ import annotations

import argparse
import hashlib
import json
import logging
import os
import subprocess
import sys
import threading
import time
from pathlib import Path

from finetune import dataset, render, train, trajectories
from finetune.students import STUDENTS

ROOT = Path(__file__).resolve().parents[3]
SUITE = ROOT / "docs/research/kaggle-development-suite/raw/qwen35-9b/run-2026-10-10-12-24-44-626Z"
STEPS = 30
VARIANTS = tuple(
    f"{student}-{precision}" for student in STUDENTS for precision in ("qlora", "lora")
)
VARIANT_MINUTES = 150

log = logging.getLogger(__name__)


def select() -> tuple[list[Path], list[Path]]:
    clean = sorted(p for p in SUITE.glob("*/prompt/001.json") if trajectories.is_clean(p))
    train_paths = [p for p in clean if p.parts[-3] not in trajectories.VALIDATION_SCENARIOS][::3]
    evals = [p for p in clean if p.parts[-3] in trajectories.VALIDATION_SCENARIOS]
    return train_paths, evals


def render_data(out: Path, students: list[str], train_paths: list[Path], evals: list[Path]) -> None:
    for name in students:
        student = STUDENTS[name]
        tok = render.tokenizer(student)
        folder = out / name
        folder.mkdir(parents=True, exist_ok=True)
        for split, paths in (("train", train_paths), ("eval", evals)):
            with (folder / f"{split}.jsonl").open("w") as f:
                for path in paths:
                    row = dataset.example(tok, student, trajectories.load(path))
                    row["source"] = str(path.relative_to(ROOT))
                    f.write(json.dumps(row) + "\n")


def split_variant(variant: str) -> tuple[str, str]:
    student, _, precision = variant.rpartition("-")
    if student not in STUDENTS or precision not in ("lora", "qlora"):
        raise ValueError(f"unknown variant {variant}")
    return student, precision


def train_command(variant: str, data: Path, out: Path, steps: int, sweep: list[int]) -> list[str]:
    student, precision = split_variant(variant)
    return [
        sys.executable,
        "-m",
        "finetune.train",
        "--student",
        student,
        "--precision",
        precision,
        "--data",
        str(data / student / "train.jsonl"),
        "--eval",
        str(data / student / "eval.jsonl"),
        "--out",
        str(out),
        "--steps",
        str(steps),
        "--max-len",
        str(train.MAX_LEN),
        "--sweep",
        ",".join(str(n) for n in sweep),
    ]


def run_variant(command: list[str], out: Path, minutes: float) -> dict:
    """Runs one training process. Its output goes to ``out/log.txt`` and, with
    the variant's name in front, to this process's output (the notebook log)."""
    out.mkdir(parents=True, exist_ok=True)
    env = {**os.environ, "PYTHONUNBUFFERED": "1"}
    env.setdefault("PYTORCH_CUDA_ALLOC_CONF", "expandable_segments:True")
    # Progress bars would fill the log with one line per update.
    env.setdefault("TQDM_DISABLE", "1")
    env.setdefault("HF_HUB_DISABLE_PROGRESS_BARS", "1")
    started = time.time()
    timed_out = threading.Event()
    with (out / "log.txt").open("w") as log_file:
        log_file.write(f"$ {' '.join(command)}\n")
        process = subprocess.Popen(
            command, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, env=env, text=True
        )

        def stop() -> None:
            timed_out.set()
            process.kill()

        timer = threading.Timer(minutes * 60, stop)
        timer.start()
        try:
            for line in process.stdout:
                log_file.write(line)
                log_file.flush()
                print(f"[{out.name}] {line}", end="", flush=True)
            code = process.wait()
        finally:
            timer.cancel()
    status = "timeout" if timed_out.is_set() else "ok" if code == 0 else f"exit {code}"
    result = {"status": status, "wall_seconds": time.time() - started}
    (out / "status.json").write_text(json.dumps(result, indent=2))
    return result


def git(*args: str) -> str:
    result = subprocess.run(["git", *args], cwd=ROOT, capture_output=True, text=True)
    return result.stdout.strip() if result.returncode == 0 else ""


def gpus() -> list[str]:
    query = ["nvidia-smi", "--query-gpu=name,memory.total,driver_version", "--format=csv,noheader"]
    try:
        result = subprocess.run(query, capture_output=True, text=True)
    except FileNotFoundError:
        return []
    return result.stdout.strip().splitlines() if result.returncode == 0 else []


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main(argv: list[str] | None = None) -> None:
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument(
        "--variants", nargs="+", choices=VARIANTS, default=list(VARIANTS), help="default: all"
    )
    parser.add_argument("--steps", type=int, default=STEPS, help="smoke runs: fewer steps")
    parser.add_argument(
        "--sweep", type=train.int_list, default=train.SWEEP, help="comma-separated lengths"
    )
    args = parser.parse_args(argv)
    logging.basicConfig(format="%(asctime)s %(levelname)s %(message)s", level=logging.INFO)

    args.out.mkdir(parents=True, exist_ok=True)
    train_paths, evals = select()
    students = sorted({split_variant(v)[0] for v in args.variants})
    data = args.out / "data"
    log.info("rendering %d training and %d evaluation trajectories", len(train_paths), len(evals))
    render_data(data, students, train_paths, evals)
    session = {
        "commit": git("rev-parse", "HEAD"),
        "uncommitted_changes": bool(git("status", "--porcelain")),
        "gpus": gpus(),
        "variants": args.variants,
        "steps": args.steps,
        "sweep": args.sweep,
        "max_len": train.MAX_LEN,
        "train": [str(p.relative_to(ROOT)) for p in train_paths],
        "eval": [str(p.relative_to(ROOT)) for p in evals],
        "data": {str(p.relative_to(data)): sha256(p) for p in sorted(data.rglob("*.jsonl"))},
        "results": {},
    }
    session_file = args.out / "session.json"
    session_file.write_text(json.dumps(session, indent=2))
    log.info("commit %s, GPUs %s", session["commit"][:12], session["gpus"])
    for i, variant in enumerate(args.variants, 1):
        command = train_command(variant, data, args.out / variant, args.steps, args.sweep)
        log.info("variant %d/%d: %s", i, len(args.variants), variant)
        result = run_variant(command, args.out / variant, VARIANT_MINUTES)
        log.info("variant %s: %s in %.0f s", variant, result["status"], result["wall_seconds"])
        session["results"][variant] = result
        session_file.write_text(json.dumps(session, indent=2))


if __name__ == "__main__":
    main()
