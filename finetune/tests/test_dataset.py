import yaml

from finetune.config import CONFIG
from finetune.data import dataset, manifest, review, select
from tests.test_select import attempt, write


def test_build_excludes_validation_scenarios_and_lists_scenarios_without_examples(
    tmp_path, monkeypatch
):
    monkeypatch.setattr(dataset, "DATASETS", tmp_path / "datasets")
    monkeypatch.setattr(select, "ROOT", tmp_path)
    monkeypatch.setattr(review, "ROOT", tmp_path)
    validation = sorted(CONFIG.validation_scenarios)[0]
    training = sorted(dataset.training_scenarios())
    runs = tmp_path / "results"
    write(runs, "001", attempt(scenario=training[0]))
    write(runs, "001", attempt(scenario=validation))

    out = dataset.build("teacher", [runs / "run-1"])

    raw = yaml.safe_load((out / "manifest.yaml").read_text())
    assert [a["scenario"] for a in raw["attempts"]] == [training[0]]
    assert raw["train"] == [a["path"] for a in raw["attempts"]]
    assert raw["attempts"][0]["review"] == "pending"
    assert set(raw["scenarios_without_examples"]) == set(training[1:])
    assert (tmp_path / raw["train"][0]).is_file()
    # Every attempt of the runs is kept, the one that fails the filter too.
    assert (out / "runs" / "run-1" / "run.json").is_file()


def test_committed_datasets_hold_only_training_scenarios_with_raw_records():
    for path in dataset.DATASETS.glob("*/manifest.yaml"):
        raw = yaml.safe_load(path.read_text())
        scenarios = {a["scenario"] for a in raw["attempts"]}
        assert not scenarios & CONFIG.validation_scenarios, path
        assert scenarios <= dataset.training_scenarios(), path
        assert not scenarios & set(raw["scenarios_without_examples"]), path
        manifest.read(path)  # every record exists, is clean and not a validation scenario


def test_build_can_be_redone_from_the_dataset_itself(tmp_path, monkeypatch):
    monkeypatch.setattr(dataset, "DATASETS", tmp_path / "datasets")
    monkeypatch.setattr(select, "ROOT", tmp_path)
    monkeypatch.setattr(review, "ROOT", tmp_path)
    training = sorted(dataset.training_scenarios())[0]
    runs = tmp_path / "results"
    write(runs, "001", attempt(scenario=training))
    out = dataset.build("teacher", [runs / "run-1"])

    again = dataset.build("teacher", [out / "runs" / "run-1"])

    assert again == out
