"""The adapter on the served model: GGUF conversion and served perplexity.

Both run in the llama.cpp image pinned in ``config.yaml`` (the release of the
benchmark's server), so the converter and the runtime are the ones the
benchmark uses:

- ``convert``: PEFT adapter -> GGUF LoRA with ``convert_lora_to_gguf.py``,
  using only the pinned checkpoint's config (no weights);
- ``perplexity``: ``llama-perplexity`` on the served 4-bit GGUF, with or
  without ``--lora``, over held-out rendered trajectories. It scores every
  token in the second half of each context window, tool results included, so
  it is a coarse, served-format check that the adapter transfers, not a
  target-token loss.
"""

from __future__ import annotations

import argparse
import json
import logging
import re
import shutil
import subprocess
from dataclasses import asdict
from pathlib import Path

from finetune import logs
from finetune.config import CONFIG, ROOT, STUDENTS, Student
from finetune.data import render
from finetune.data.rows import read_rows
from finetune.serving.models import Conversion, Perplexity, Served

MODELS = ROOT / "benchmark" / "models"


PPL = re.compile(r"Final estimate: PPL = ([0-9.]+) \+/- ([0-9.]+)")
log = logging.getLogger(__name__)


def snapshot(student: Student) -> Path:
    from huggingface_hub import snapshot_download

    return Path(
        snapshot_download(
            student.checkpoint, revision=student.revision, allow_patterns=["*.json", "*.jinja"]
        )
    )


def docker(
    image: str,
    args: list[str],
    mounts: list[tuple[Path, str]],
    gpu: bool = False,
    entrypoint: str | None = None,
) -> subprocess.CompletedProcess:
    cmd = ["docker", "run", "--rm"]
    if gpu:
        cmd += ["--device", "/dev/dri"]
    for host, container in mounts:
        cmd += ["-v", f"{host}:{container}"]
    if entrypoint:
        cmd += ["--entrypoint", entrypoint]
    return subprocess.run(cmd + [image, *args], capture_output=True, text=True)


def convert(student: Student, adapter: Path, out: Path) -> Conversion:
    log.info("converting %s to %s", adapter, out)
    out.parent.mkdir(parents=True, exist_ok=True)
    # The HF cache snapshot holds relative symlinks into its blob store; copy
    # the resolved config files so the mounted directory is self-contained.
    base = out.parent / "base-config"
    base.mkdir(exist_ok=True)
    for file in snapshot(student).iterdir():
        if file.suffix in (".json", ".jinja"):
            shutil.copyfile(file, base / file.name)
    result = docker(
        CONFIG.llama_cpp.ref(gpu=False),
        [
            "/app/convert_lora_to_gguf.py",
            "--outtype",
            "f16",
            "--base",
            "/base",
            "--outfile",
            f"/out/{out.name}",
            "/adapter",
        ],
        [
            (base.resolve(), "/base:ro"),
            (adapter.resolve(), "/adapter:ro"),
            (out.parent.resolve(), "/out"),
        ],
        entrypoint="python3",
    )
    return Conversion(
        ok=result.returncode == 0 and out.exists(),
        returncode=result.returncode,
        log_tail=(result.stdout + result.stderr)[-4000:],
    )


def perplexity(
    student: Student, text: Path, lora: Path | None, ctx: int = 8192, gpu: bool = True
) -> Perplexity:
    log.info("perplexity of %s on %s, adapter %s", student.served_model, text.name, lora)
    mounts = [(MODELS.resolve(), "/models:ro"), (text.parent.resolve(), "/data:ro")]
    args = [
        "-m",
        f"/models/{student.served_gguf}",
        "-f",
        f"/data/{text.name}",
        "-c",
        str(ctx),
        "-b",
        "2048",
        "-ngl",
        "999" if gpu else "0",
    ]
    if lora:
        mounts.append((lora.parent.resolve(), "/lora:ro"))
        args += ["--lora", f"/lora/{lora.name}"]
    result = docker(
        CONFIG.llama_cpp.ref(gpu), args, mounts, gpu=gpu, entrypoint="/app/llama-perplexity"
    )
    output = result.stdout + result.stderr
    match = PPL.search(output)
    return Perplexity(
        ok=bool(match),
        ppl=float(match.group(1)) if match else None,
        ppl_error=float(match.group(2)) if match else None,
        lora_log=[line for line in output.splitlines() if "lora" in line.lower()][:20],
        log_tail=output[-3000:],
    )


def eval_text(rows: Path, out: Path, bos: str) -> Path:
    """Held-out rendered trajectories as one text file, BOS text removed
    (llama.cpp adds BOS as a token)."""
    out.write_text("\n".join(row.text.removeprefix(bos) for row in read_rows(rows)))
    return out


def main() -> None:
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    parser.add_argument("--student", choices=sorted(STUDENTS), required=True)
    parser.add_argument("--adapter", type=Path, required=True, help="PEFT adapter directory")
    parser.add_argument("--eval", type=Path, required=True, help="held-out JSONL")
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--cpu", action="store_true", help="llama.cpp on CPU instead of Vulkan")
    args = parser.parse_args()
    logs.setup()
    student = STUDENTS[args.student]
    args.out.mkdir(parents=True, exist_ok=True)
    gguf = args.out / "adapter-f16.gguf"
    served = Served(convert(student, args.adapter, gguf))
    if served.convert.ok:
        text = eval_text(args.eval, args.out / "eval.txt", render.bos_text(student))
        served = Served(
            served.convert,
            without_adapter=perplexity(student, text, None, gpu=not args.cpu),
            with_adapter=perplexity(student, text, gguf, gpu=not args.cpu),
        )
    (args.out / "served.json").write_text(json.dumps(asdict(served), indent=2))
    if not served.convert.ok:
        log.error("conversion failed:\n%s", served.convert.log_tail)
    for name, result in (
        ("without adapter", served.without_adapter),
        ("with adapter", served.with_adapter),
    ):
        if result:
            log.info("served perplexity %s: %s +/- %s", name, result.ppl, result.ppl_error)


if __name__ == "__main__":
    main()
