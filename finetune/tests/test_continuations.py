import json

import yaml

from finetune.data import continuations, dataset, review, select
from tests.test_select import attempt, write

TRAIN = sorted(dataset.training_scenarios())[0]


def record(tmp_path, scenario=TRAIN, full=False, termination="completed", last=None):
    messages = [
        {"role": "system", "content": "s"},
        {"role": "user", "content": "u"},
        {"role": "assistant", "content": "", "tool_calls": [{"id": "c"}]},
        {"role": "tool", "content": "out"},
        last or {"role": "assistant", "content": "done"},
    ]
    path = tmp_path / "001.json"
    path.write_text(
        json.dumps(
            {
                "scenario_id": scenario,
                "grading": {"full_success": full, "score": 0.5},
                "agent": {
                    "termination": termination,
                    "messages": messages,
                    "inference": {"model": "gemma"},
                },
            }
        )
    )
    return path


def test_a_failed_attempt_that_stopped_on_a_final_message_is_a_prefix(tmp_path, monkeypatch):
    monkeypatch.setattr(continuations, "ROOT", tmp_path)
    found = continuations.prefix(record(tmp_path), {TRAIN})
    assert found.prefix_messages == 4  # everything before the premature final message
    assert found.scenario == TRAIN


def test_other_attempts_are_not_prefixes(tmp_path, monkeypatch):
    monkeypatch.setattr(continuations, "ROOT", tmp_path)
    allowed = {TRAIN}
    assert continuations.prefix(record(tmp_path, full=True), allowed) is None
    assert continuations.prefix(record(tmp_path, scenario="validation-one"), allowed) is None
    assert continuations.prefix(record(tmp_path, termination="turn_limit"), allowed) is None
    tool_call = {"role": "assistant", "content": "", "tool_calls": [{"id": "d"}]}
    assert continuations.prefix(record(tmp_path, last=tool_call), allowed) is None


def test_build_keeps_the_clean_continuations_and_writes_a_sheet(tmp_path, monkeypatch):
    monkeypatch.setattr(dataset, "DATASETS", tmp_path / "datasets")
    monkeypatch.setattr(
        continuations, "OUT", tmp_path / "datasets" / "continuation" / "candidates.yaml"
    )
    monkeypatch.setattr(select, "ROOT", tmp_path)
    monkeypatch.setattr(review, "ROOT", tmp_path)
    monkeypatch.setattr(continuations, "ROOT", tmp_path)
    scenario = tmp_path / "benchmark" / "scenarios" / "storage" / TRAIN / "scenario.yaml"
    scenario.parent.mkdir(parents=True)
    scenario.write_text("sources: []")
    continuations.OUT.parent.mkdir(parents=True)
    continuations.OUT.write_text(
        yaml.safe_dump([{"source": "x", "model": "gemma", "scenario": TRAIN}])
    )
    runs = tmp_path / "results"
    write(runs, "001", attempt(scenario=TRAIN, agent={"prefix_messages": 2}))
    write(runs, "002", attempt(scenario=TRAIN, grading={"full_success": False, "score": 0.5}))

    out = continuations.build("teacher", [runs / "run-1"])

    raw = yaml.safe_load((out / "manifest.yaml").read_text())
    assert [a["attempt"] for a in raw["attempts"]] == [1]
    assert raw["attempts"][0]["review"] == review.PENDING
    assert raw["scenarios_without_examples"] == []
    assert len(list((out / "review").glob("*.md"))) == 1


def test_freeze_refuses_while_a_continuation_waits_for_review(tmp_path, monkeypatch):
    monkeypatch.setattr(dataset, "DATASETS", tmp_path / "datasets")
    out = continuations.directory("teacher")
    out.mkdir(parents=True)
    (out / "manifest.yaml").write_text(
        yaml.safe_dump(
            {
                "teacher": "teacher",
                "train": ["x.json"],
                "attempts": [{"path": "x.json", "review": review.PENDING}],
            }
        )
    )

    try:
        continuations.freeze("teacher")
    except ValueError as error:
        assert "wait for review" in str(error)
    else:
        raise AssertionError("the freeze ran while a sheet was pending")
