"""The adapter on the served model: GGUF conversion and served perplexity.

Both run in the pinned llama.cpp image (the release of the benchmark's server,
``b10964``), so the converter and the runtime are the ones the benchmark uses:

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
import re
import shutil
import subprocess
from pathlib import Path

from finetune import render
from finetune.students import STUDENTS, Student

IMAGE = (
    "ghcr.io/ggml-org/llama.cpp:full-b10964"
    "@sha256:c983c142e666c6d69797b789998790b26305c8b52305dcafa78ffcfb9e06a0ea"
)
IMAGE_VULKAN = (
    "ghcr.io/ggml-org/llama.cpp:full-vulkan-b10964"
    "@sha256:f4ab9e13181311d341b05d52dd4c324b66c592a7d4b90b7681c105febdc84e83"
)
REPO = Path(__file__).resolve().parents[3]
MODELS = REPO / "benchmark" / "models"
PPL = re.compile(r"Final estimate: PPL = ([0-9.]+) \+/- ([0-9.]+)")


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


def convert(student: Student, adapter: Path, out: Path) -> dict:
    out.parent.mkdir(parents=True, exist_ok=True)
    # The HF cache snapshot holds relative symlinks into its blob store; copy
    # the resolved config files so the mounted directory is self-contained.
    base = out.parent / "base-config"
    base.mkdir(exist_ok=True)
    for file in snapshot(student).iterdir():
        if file.suffix in (".json", ".jinja"):
            shutil.copyfile(file, base / file.name)
    result = docker(
        IMAGE,
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
    return {
        "ok": result.returncode == 0 and out.exists(),
        "returncode": result.returncode,
        "log_tail": (result.stdout + result.stderr)[-4000:],
    }


def perplexity(
    student: Student, text: Path, lora: Path | None, ctx: int = 8192, gpu: bool = True
) -> dict:
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
    image = IMAGE_VULKAN if gpu else IMAGE
    result = docker(image, args, mounts, gpu=gpu, entrypoint="/app/llama-perplexity")
    log = result.stdout + result.stderr
    match = PPL.search(log)
    return {
        "ok": bool(match),
        "ppl": float(match.group(1)) if match else None,
        "ppl_error": float(match.group(2)) if match else None,
        # llama.cpp's own lines about loading the adapter (empty without one).
        "lora_log": [line for line in log.splitlines() if "lora" in line.lower()][:20],
        "log_tail": log[-3000:],
    }


def eval_text(rows: Path, out: Path, bos: str) -> Path:
    """Held-out rendered trajectories as one text file, BOS text removed
    (llama.cpp adds BOS as a token)."""
    texts = [json.loads(line)["text"] for line in rows.read_text().splitlines() if line]
    out.write_text("\n".join(t.removeprefix(bos) for t in texts))
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
    student = STUDENTS[args.student]
    args.out.mkdir(parents=True, exist_ok=True)
    gguf = args.out / "adapter-f16.gguf"
    result = {"convert": convert(student, args.adapter, gguf)}
    if result["convert"]["ok"]:
        text = eval_text(args.eval, args.out / "eval.txt", render.bos_text(student))
        result["without_adapter"] = perplexity(student, text, None, gpu=not args.cpu)
        result["with_adapter"] = perplexity(student, text, gguf, gpu=not args.cpu)
    (args.out / "served.json").write_text(json.dumps(result, indent=2))
    print(
        json.dumps(
            {k: {kk: vv for kk, vv in v.items() if kk != "log_tail"} for k, v in result.items()},
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
