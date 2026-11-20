"""Kaggle kernel of the technical check, step 7: an adapter in the Kaggle server.

Uses the llama-server that KaggleInferenceServer compiles for the T4 (the
``kis-llamacpp-build`` kernel output, attached as a kernel source) and the
served 4-bit GGUF at its pinned revision. GPU 0 serves the base model, GPU 1
the same file with ``--lora``; both get the same fixed prompts greedily, with
the benchmark's server settings (MTP drafts, reasoning on). Writes
/kaggle/working/serve.json.

The bundle dataset holds ``adapter.gguf``, ``prompts.json`` (from
``finetune.serving.prompts``) and ``model.json`` (repository, revision, file,
sha256, alias, draft length).
"""

from __future__ import annotations

import hashlib
import json
import os
import shutil
import subprocess
import sys
import time
import urllib.request
from pathlib import Path

WORK = Path("/kaggle/working")
MAX_TOKENS = 256
# The KIS binary links the CUDA 12 runtime and cuBLAS dynamically; newer Kaggle
# images ship another CUDA, so the CUDA 12 libraries come from these wheels.
CUDA12 = ["nvidia-cuda-runtime-cu12==12.9.79", "nvidia-cublas-cu12==12.9.3.13"]


def find(name: str) -> Path:
    found = sorted(Path("/kaggle/input").rglob(name))
    if not found:
        raise SystemExit(f"{name} not found in /kaggle/input")
    return found[0]


def download(model: dict) -> Path:
    subprocess.run(
        [sys.executable, "-m", "pip", "install", "--quiet", "huggingface_hub"], check=True
    )
    from huggingface_hub import hf_hub_download

    path = Path(
        hf_hub_download(
            model["repo"],
            model["file"],
            revision=model["revision"],
            local_dir="/kaggle/temp/models",
        )
    )
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    if digest != model["sha256"]:
        raise SystemExit(f"sha256 mismatch for {path}: {digest}")
    return path


def cuda12_libraries() -> str:
    """Installs the CUDA 12 runtime libraries; returns their directories."""
    target = Path("/tmp/cuda12")
    subprocess.run(
        [sys.executable, "-m", "pip", "install", "--quiet", "--target", str(target), *CUDA12],
        check=True,
    )
    return ":".join(sorted({str(p.parent) for p in target.rglob("lib*.so.12")}))


def start(
    binary: Path,
    gguf: Path,
    port: int,
    gpu: int,
    model: dict,
    lora: Path | None,
    libraries: str = "",
) -> subprocess.Popen:
    args = [
        str(binary),
        "-m",
        str(gguf),
        "--alias",
        model["alias"],
        "--host",
        "127.0.0.1",
        "--port",
        str(port),
        "-c",
        "8192",
        "--parallel",
        "1",
        "-ngl",
        "999",
        "--flash-attn",
        "auto",
        "--reasoning",
        "on",
        "--reasoning-budget",
        "4096",
        "--spec-type",
        "draft-mtp",
        "--spec-draft-n-max",
        str(model["draft_n_max"]),
    ]
    if lora:
        args += ["--lora", str(lora)]
    log = (WORK / f"llama-{port}.log").open("w")
    env = {**os.environ, "CUDA_VISIBLE_DEVICES": str(gpu)}
    if libraries:
        env["LD_LIBRARY_PATH"] = ":".join(filter(None, [libraries, env.get("LD_LIBRARY_PATH")]))
    return subprocess.Popen(args, stdout=log, stderr=subprocess.STDOUT, env=env)


def wait(port: int, process: subprocess.Popen) -> None:
    for _ in range(300):
        if process.poll() is not None:
            raise SystemExit(f"llama-server on {port} exited; see llama-{port}.log")
        try:
            with urllib.request.urlopen(f"http://127.0.0.1:{port}/health", timeout=5):
                return
        except OSError:
            time.sleep(2)
    raise SystemExit(f"llama-server on {port} did not become ready")


def complete(port: int, model: str, prompt: dict) -> dict:
    body = json.dumps(
        {
            "model": model,
            "messages": prompt["messages"],
            "tools": prompt["tools"],
            "temperature": 0,
            "max_tokens": MAX_TOKENS,
        }
    ).encode()
    request = urllib.request.Request(
        f"http://127.0.0.1:{port}/v1/chat/completions", body, {"Content-Type": "application/json"}
    )
    with urllib.request.urlopen(request, timeout=900) as response:
        data = json.load(response)
    message = data["choices"][0]["message"]
    timings = data.get("timings", {})
    return {
        "text": (message.get("reasoning_content") or "") + (message.get("content") or ""),
        "predicted_per_second": timings.get("predicted_per_second"),
        "draft_n": timings.get("draft_n"),
        "draft_n_accepted": timings.get("draft_n_accepted"),
    }


def main() -> None:
    binary = Path("/tmp/llama-server")
    shutil.copy(find("llama-server"), binary)
    binary.chmod(0o755)
    libraries = cuda12_libraries()
    version = subprocess.run(
        [str(binary), "--version"],
        capture_output=True,
        text=True,
        env={
            **os.environ,
            "LD_LIBRARY_PATH": f"{libraries}:{os.environ.get('LD_LIBRARY_PATH', '')}",
        },
    )
    model = json.loads(find("model.json").read_text())
    prompts = json.loads(find("prompts.json").read_text())
    gguf = download(model)
    adapter = find("adapter.gguf")

    servers = {
        8090: start(binary, gguf, 8090, 0, model, None, libraries),
        8091: start(binary, gguf, 8091, 1, model, adapter, libraries),
    }
    for port, process in servers.items():
        wait(port, process)
    runs = []
    for prompt in prompts:
        base, adapted = (
            complete(8090, model["alias"], prompt),
            complete(8091, model["alias"], prompt),
        )
        runs.append(
            {
                "source": prompt["source"],
                "base": base,
                "adapted": adapted,
                "same_text": base["text"] == adapted["text"],
            }
        )
    for process in servers.values():
        process.terminate()
    (WORK / "serve.json").write_text(
        json.dumps(
            {
                "llama_server_version": (version.stdout + version.stderr).strip(),
                "model": model,
                "runs": runs,
            },
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
