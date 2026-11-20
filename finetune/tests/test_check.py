"""The Kaggle check driver (selection, data, commands, statuses), without Kaggle."""

import hashlib
import json
import sys

import pytest

from finetune import check, render, train, trajectories
from tests.conftest import TemplateTokenizer


@pytest.fixture
def template_tokenizers(monkeypatch):
    monkeypatch.setattr(render, "tokenizer", lambda student: TemplateTokenizer(student.family))


def test_selection_is_clean_and_never_trains_on_validation_scenarios():
    train_paths, evals = check.select()
    assert len(train_paths) == 10
    assert len(evals) == 6
    assert all(trajectories.is_clean(p) for p in train_paths + evals)
    assert not {p.parts[-3] for p in train_paths} & trajectories.VALIDATION_SCENARIOS
    assert {p.parts[-3] for p in evals} <= trajectories.VALIDATION_SCENARIOS
    assert len({p.parts[-3] for p in train_paths}) == 10


def test_variants_are_every_model_in_both_precisions():
    assert check.VARIANTS == (
        "gemma-4-e2b-qlora",
        "gemma-4-e2b-lora",
        "gemma-4-e4b-qlora",
        "gemma-4-e4b-lora",
        "qwen35-4b-qlora",
        "qwen35-4b-lora",
        "qwen35-9b-qlora",
        "qwen35-9b-lora",
    )
    assert check.split_variant("gemma-4-e4b-qlora") == ("gemma-4-e4b", "qlora")
    with pytest.raises(ValueError):
        check.split_variant("gemma-4-e4b-full")
    with pytest.raises(ValueError):
        check.split_variant("llama-3-8b-lora")


def test_train_command_is_accepted_by_finetune_train(tmp_path):
    command = check.train_command("qwen35-4b-qlora", tmp_path / "data", tmp_path / "out", 2, [64])
    assert command[:3] == [sys.executable, "-m", "finetune.train"]
    args = train.parse_args(command[3:])
    assert args.student == "qwen35-4b"
    assert args.precision == "qlora"
    assert args.data == tmp_path / "data" / "qwen35-4b" / "train.jsonl"
    assert args.eval == tmp_path / "data" / "qwen35-4b" / "eval.jsonl"
    assert (args.steps, args.sweep, args.max_len) == (2, [64], train.MAX_LEN)


@pytest.mark.parametrize(
    "script,expected", [("pass", "ok"), ("import sys; print('x'); sys.exit(3)", "exit 3")]
)
def test_variant_status_and_log_are_recorded(tmp_path, capsys, script, expected):
    result = check.run_variant([sys.executable, "-c", script], tmp_path / "v", minutes=1)
    assert result["status"] == expected
    saved = json.loads((tmp_path / "v" / "status.json").read_text())
    assert saved["status"] == expected
    assert (tmp_path / "v" / "log.txt").read_text().startswith("$ ")
    if "print" in script:
        assert "x\n" in (tmp_path / "v" / "log.txt").read_text()
        assert "[v] x\n" in capsys.readouterr().out


def test_variant_that_runs_too_long_is_stopped(tmp_path):
    command = [sys.executable, "-c", "import time; time.sleep(30)"]
    result = check.run_variant(command, tmp_path / "v", minutes=0.01)
    assert result["status"] == "timeout"
    assert result["wall_seconds"] < 10


def test_main_renders_data_and_runs_each_variant_in_order(
    tmp_path, monkeypatch, template_tokenizers
):
    ran = []

    def fake_run(command, out, minutes):
        ran.append(train.parse_args(command[3:]))
        return {"status": "ok", "wall_seconds": 1.0}

    monkeypatch.setattr(check, "run_variant", fake_run)
    check.main(
        [
            "--out",
            str(tmp_path),
            "--variants",
            "qwen35-4b-lora",
            "gemma-4-e4b-qlora",
            "--steps",
            "2",
        ]
    )

    assert [(a.student, a.precision) for a in ran] == [
        ("qwen35-4b", "lora"),
        ("gemma-4-e4b", "qlora"),
    ]
    session = json.loads((tmp_path / "session.json").read_text())
    assert session["steps"] == 2
    assert session["sweep"] == train.SWEEP
    assert len(session["train"]) == 10 and len(session["eval"]) == 6
    assert set(session["results"]) == {"qwen35-4b-lora", "gemma-4-e4b-qlora"}
    assert len(session["commit"]) == 40
    for name, digest in session["data"].items():
        path = tmp_path / "data" / name
        assert hashlib.sha256(path.read_bytes()).hexdigest() == digest
    rows = (tmp_path / "data" / "gemma-4-e4b" / "train.jsonl").read_text().splitlines()
    assert len(rows) == 10 and all(json.loads(r)["targets"] for r in rows)
    assert sorted(p.name for p in (tmp_path / "data").iterdir()) == ["gemma-4-e4b", "qwen35-4b"]
