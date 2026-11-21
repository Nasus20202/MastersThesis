import json
from dataclasses import replace

import pytest
import yaml

from finetune.data import dataset, freeze, review, select
from tests.conftest import ATTEMPT, TemplateTokenizer


class CountingTokenizer(TemplateTokenizer):
    """Template rendering plus one token per character."""

    def __call__(self, text, add_special_tokens=False):
        return {"input_ids": list(text)}


def limit(monkeypatch, max_len):
    config = freeze.CONFIG
    monkeypatch.setattr(
        freeze, "CONFIG", replace(config, training=replace(config.training, max_len=max_len))
    )


def manifest_for(tmp_path, monkeypatch, reviews=("accepted",)):
    monkeypatch.setattr(freeze, "ROOT", tmp_path)
    raw = tmp_path / "raw" / "run-1" / "pvc-pending" / "prompt"
    raw.mkdir(parents=True)
    attempts = []
    for i, status in enumerate(reviews, 1):
        (raw / f"{i:03d}.json").write_text(ATTEMPT.read_text())
        attempts.append(
            {
                "scenario": "pvc-pending",
                "review": status,
                "path": f"raw/run-1/pvc-pending/prompt/{i:03d}.json",
            }
        )
    directory = tmp_path / "dataset"
    directory.mkdir()
    manifest = {
        "teacher": "t",
        "train": [a["path"] for a in attempts],
        "scenarios_without_examples": ["x"],
        "attempts": attempts,
    }
    (directory / "manifest.yaml").write_text(yaml.safe_dump(manifest))
    return directory


def test_freeze_renders_counts_and_leaves_out_over_length_examples(tmp_path, monkeypatch):
    directory = manifest_for(tmp_path, monkeypatch)
    limit(monkeypatch, 10**9)
    tokenizers = {"qwen35-9b": CountingTokenizer("qwen"), "gemma-4-e4b": CountingTokenizer("gemma")}

    first = freeze.freeze(directory, tokenizers)

    assert first.attempts == 1 and first.per_scenario == {"pvc-pending": 1}
    assert [r.examples for r in first.rendered] == [1, 1]
    rows = (directory / "rendered" / "qwen35-9b.jsonl").read_text().splitlines()
    assert json.loads(rows[0])["source"] == "raw/run-1/pvc-pending/prompt/001.json"
    assert freeze.freeze(directory, tokenizers).revision == first.revision

    limit(monkeypatch, 5)
    short = freeze.freeze(directory, tokenizers)
    assert [r.examples for r in short.rendered] == [0, 0]
    assert short.rendered[0].over_length == ["raw/run-1/pvc-pending/prompt/001.json"]


@pytest.mark.parametrize("status", ["pending", "rejected"])
def test_freeze_refuses_unreviewed_or_rejected_attempts(tmp_path, monkeypatch, status):
    directory = manifest_for(tmp_path, monkeypatch, reviews=("accepted", status))
    with pytest.raises(ValueError):
        freeze.freeze(directory, {})


def test_rejected_attempt_is_replaced_by_the_next_candidate_and_reviewed(tmp_path, monkeypatch):
    monkeypatch.setattr(dataset, "DATASETS", tmp_path / "datasets")
    monkeypatch.setattr(select, "ROOT", tmp_path)
    monkeypatch.setattr(review, "ROOT", tmp_path)
    scenario = sorted(dataset.training_scenarios())[0]
    runs = tmp_path / "results"
    from tests.test_select import attempt, write

    for i, tokens in enumerate((100, 200, 300), 1):
        write(
            runs,
            f"{i:03d}",
            attempt(scenario=scenario, commands=(f"kubectl delete pod p{i}",), tokens=tokens),
        )
    config = dataset.CONFIG
    monkeypatch.setattr(
        dataset, "CONFIG", replace(config, dataset=replace(config.dataset, max_per_scenario=2))
    )
    out = dataset.build("t", [runs / "run-1"])
    first = yaml.safe_load((out / "manifest.yaml").read_text())["attempts"]
    assert [a["tokens"] for a in first] == [100, 200]

    sheet = next(out.joinpath("review").glob("*-001.md"))
    sheet.write_text(sheet.read_text().replace("verdict: pending", "verdict: rejected"))
    dataset.build("t", [runs / "run-1"])

    second = {
        a["tokens"]: a["review"]
        for a in yaml.safe_load((out / "manifest.yaml").read_text())["attempts"]
    }
    assert second == {200: "pending", 300: "pending"}


def test_pending_sheet_of_an_attempt_that_is_no_longer_kept_is_removed(tmp_path, monkeypatch):
    monkeypatch.setattr(dataset, "DATASETS", tmp_path / "datasets")
    monkeypatch.setattr(select, "ROOT", tmp_path)
    monkeypatch.setattr(review, "ROOT", tmp_path)
    scenario = sorted(dataset.training_scenarios())[0]
    runs = tmp_path / "results"
    from tests.test_select import attempt, write

    config = dataset.CONFIG
    monkeypatch.setattr(
        dataset, "CONFIG", replace(config, dataset=replace(config.dataset, max_per_scenario=1))
    )
    write(runs, "001", attempt(scenario=scenario, commands=("kubectl delete pod a",), tokens=300))
    out = dataset.build("t", [runs / "run-1"])
    assert [p.name for p in (out / "review").glob("*.md")] == [f"{scenario}-run-1-001.md"]

    write(
        runs,
        "002",
        attempt(scenario=scenario, commands=("kubectl delete pod b",), tokens=100, attempt=2),
    )
    dataset.build("t", [runs / "run-1"])

    assert [p.name for p in (out / "review").glob("*.md")] == [f"{scenario}-run-1-002.md"]
