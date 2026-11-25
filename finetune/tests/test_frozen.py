"""Frozen dataset rows: reading them, and refusing a dataset that changed."""

import json

import pytest

from finetune.data import frozen
from finetune.data.models import Row


def write(tmp_path, monkeypatch, examples: dict[str, list[Row]]):
    """A frozen dataset under ``tmp_path``, with its record."""
    monkeypatch.setattr(frozen, "ROOT", tmp_path)
    folder = tmp_path / "teacher" / "rendered"
    folder.mkdir(parents=True)
    for student, rows in examples.items():
        (folder / f"{student}.jsonl").write_text("".join(r.to_json() + "\n" for r in rows))
    manifest = tmp_path / "teacher" / "manifest.yaml"
    manifest.write_text("teacher: teacher\n")
    files = [manifest, *sorted(folder.glob("*.jsonl"))]
    (tmp_path / "teacher" / "freeze.json").write_text(
        json.dumps(
            {
                "teacher": "teacher",
                "attempts": 1,
                "revision": frozen.digest(files),
                "files": [str(p.relative_to(tmp_path)) for p in files],
            }
        )
    )
    return tmp_path / "teacher"


def test_read_returns_each_students_rows_and_the_revision(tmp_path, monkeypatch):
    example = Row("runs/run-1/a.json", "scenario", "text", [(0, 4)])
    directory = write(tmp_path, monkeypatch, {"gemma-4-e2b": [example], "qwen35-4b": [example]})

    found = frozen.read(directory)

    assert sorted(found.rows) == ["gemma-4-e2b", "qwen35-4b"]
    assert found.rows["gemma-4-e2b"] == [example]
    assert found.revision == json.loads((directory / "freeze.json").read_text())["revision"]


def test_a_teacher_dataset_is_read_by_its_manifest_and_renders(tmp_path, monkeypatch):
    example = Row("runs/run-1/a.json", "scenario", "text", [(0, 4)])
    directory = write(tmp_path, monkeypatch, {"qwen35-4b": [example]})
    record = directory / "freeze.json"
    saved = json.loads(record.read_text())
    del saved["files"]  # a teacher freeze records no file list
    record.write_text(json.dumps(saved))

    assert frozen.read(directory).rows["qwen35-4b"] == [example]
    (directory / "manifest.yaml").write_text("teacher: edited\n")
    with pytest.raises(ValueError, match="no longer matches"):
        frozen.read(directory)


def test_read_refuses_rows_that_no_longer_match_the_revision(tmp_path, monkeypatch):
    directory = write(tmp_path, monkeypatch, {"gemma-4-e2b": [Row("r", "s", "text", [(0, 1)])]})
    path = directory / "rendered" / "gemma-4-e2b.jsonl"
    path.write_text(path.read_text().replace("text", "edited"))

    with pytest.raises(ValueError, match="no longer matches"):
        frozen.read(directory)


def test_read_needs_a_record(tmp_path):
    with pytest.raises(ValueError, match="no freeze record"):
        frozen.read(tmp_path)
