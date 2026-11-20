"""GGUF conversion and served perplexity commands, with docker mocked."""

import json
import subprocess

from finetune import llama
from finetune.students import STUDENTS

PPL_LOG = "[1]5.1234,[2]4.9876,\nFinal estimate: PPL = 4.9876 +/- 0.04321\n"


def fake_docker(calls, stdout=PPL_LOG, code=0):
    def run(image, args, mounts, gpu=False, entrypoint=None):
        calls.append(
            {"image": image, "args": args, "mounts": mounts, "gpu": gpu, "entrypoint": entrypoint}
        )
        return subprocess.CompletedProcess(args, code, stdout, "")

    return run


def test_images_are_the_benchmark_release_pinned_by_digest():
    assert "b10964@sha256:" in llama.IMAGE
    assert "vulkan-b10964@sha256:" in llama.IMAGE_VULKAN


def test_perplexity_with_and_without_adapter(tmp_path, monkeypatch):
    calls = []
    monkeypatch.setattr(llama, "docker", fake_docker(calls))
    text = tmp_path / "eval.txt"
    text.write_text("x")
    lora = tmp_path / "a" / "adapter-f16.gguf"
    student = STUDENTS["qwen35-4b"]

    base = llama.perplexity(student, text, None)
    adapted = llama.perplexity(student, text, lora)

    assert base["ppl"] == 4.9876 and base["ppl_error"] == 0.04321
    assert "--lora" not in calls[0]["args"]
    assert calls[1]["args"][calls[1]["args"].index("--lora") + 1] == "/lora/adapter-f16.gguf"
    assert calls[0]["args"][calls[0]["args"].index("-m") + 1] == "/models/Qwen3.5-4B-Q4_K_M.gguf"
    assert adapted["ok"]
    assert adapted["lora_log"] == []


def test_missing_estimate_is_not_ok(tmp_path, monkeypatch):
    monkeypatch.setattr(llama, "docker", fake_docker([], stdout="error: failed to load"))
    text = tmp_path / "eval.txt"
    text.write_text("x")
    result = llama.perplexity(STUDENTS["qwen35-4b"], text, None)
    assert not result["ok"] and result["ppl"] is None


def test_eval_text_drops_bos_text(tmp_path):
    rows = tmp_path / "eval.jsonl"
    rows.write_text("\n".join(json.dumps({"text": t}) for t in ["<bos>a", "<bos>b"]) + "\n")
    out = llama.eval_text(rows, tmp_path / "eval.txt", "<bos>")
    assert out.read_text() == "a\nb"
    assert llama.eval_text(rows, tmp_path / "raw.txt", "").read_text() == "<bos>a\n<bos>b"


def test_convert_uses_only_the_checkpoint_config(tmp_path, monkeypatch):
    calls = []
    monkeypatch.setattr(llama, "docker", fake_docker(calls, stdout="done"))
    snap = tmp_path / "snap"
    snap.mkdir()
    (snap / "config.json").write_text("{}")
    monkeypatch.setattr(llama, "snapshot", lambda student: snap)
    out = tmp_path / "out" / "adapter-f16.gguf"
    out.parent.mkdir()
    out.write_text("gguf")
    result = llama.convert(STUDENTS["gemma-4-e4b"], tmp_path / "adapter", out)
    args = calls[0]["args"]
    assert result["ok"]
    assert args[0] == "/app/convert_lora_to_gguf.py"
    assert args[args.index("--outtype") + 1] == "f16"
    assert args[args.index("--base") + 1] == "/base"
    assert calls[0]["image"] == llama.IMAGE
    assert (out.parent / "base-config" / "config.json").read_text() == "{}"
