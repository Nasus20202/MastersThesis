"""Step-7 kernel: server command lines and the base/adapted comparison, mocked."""

import importlib.util
import json
import subprocess
import sys
from pathlib import Path

SERVE = Path(__file__).resolve().parents[1] / "kaggle" / "serve" / "kernel.py"
spec = importlib.util.spec_from_file_location("serve_kernel", SERVE)
serve = importlib.util.module_from_spec(spec)
sys.modules["serve_kernel"] = serve
spec.loader.exec_module(serve)

MODEL = {"alias": "Qwen3.5-4B-Q4_K_M", "draft_n_max": 2}


def test_only_the_adapted_server_gets_the_lora(tmp_path, monkeypatch):
    started = []

    class FakePopen:
        def __init__(self, args, env, **kwargs):
            started.append((args, env["CUDA_VISIBLE_DEVICES"]))

    monkeypatch.setattr(serve, "WORK", tmp_path)
    monkeypatch.setattr(subprocess, "Popen", FakePopen)
    serve.start(Path("/b"), Path("/m.gguf"), 8090, 0, MODEL, None)
    serve.start(Path("/b"), Path("/m.gguf"), 8091, 1, MODEL, Path("/a.gguf"))
    (base, gpu0), (adapted, gpu1) = started
    assert (gpu0, gpu1) == ("0", "1")
    assert "--lora" not in base
    assert adapted[adapted.index("--lora") + 1] == "/a.gguf"
    assert base[base.index("--spec-draft-n-max") + 1] == "2"

    def without(args, *flags):
        out, skip = [], False
        for arg in args:
            if skip:
                skip = False
            elif arg in flags:
                skip = True
            else:
                out.append(arg)
        return out

    assert without(base, "--port") == without(adapted, "--port", "--lora")


def test_servers_find_the_cuda12_libraries(tmp_path, monkeypatch):
    envs = []

    class FakePopen:
        def __init__(self, args, env, **kwargs):
            envs.append(env)

    monkeypatch.setattr(serve, "WORK", tmp_path)
    monkeypatch.setattr(subprocess, "Popen", FakePopen)
    monkeypatch.setenv("LD_LIBRARY_PATH", "/usr/lib")
    serve.start(Path("/b"), Path("/m.gguf"), 8090, 0, MODEL, None, "/tmp/cuda12/lib")
    assert envs[0]["LD_LIBRARY_PATH"] == "/tmp/cuda12/lib:/usr/lib"


def test_complete_reads_text_and_timings(monkeypatch):
    class Response:
        def __enter__(self):
            return self

        def __exit__(self, *a):
            return False

        def read(self):
            return json.dumps(
                {
                    "choices": [{"message": {"reasoning_content": "r", "content": "c"}}],
                    "timings": {"predicted_per_second": 30.0, "draft_n": 10, "draft_n_accepted": 7},
                }
            ).encode()

    monkeypatch.setattr(serve.urllib.request, "urlopen", lambda request, timeout: Response())
    out = serve.complete(8090, "m", {"messages": [], "tools": []})
    assert out == {"text": "rc", "predicted_per_second": 30.0, "draft_n": 10, "draft_n_accepted": 7}
