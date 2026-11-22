"""Run inputs: manifest, rendering, provenance and training jobs."""

import hashlib
import json

import pytest

from finetune.config import CONFIG, ROOT
from finetune.data import frozen, manifest, render, trajectories
from finetune.data.models import Manifest, Row
from finetune.runs import prepare
from finetune.runs.models import Variant
from finetune.training.models import Precision
from tests.conftest import ATTEMPT, TemplateTokenizer

POC_MANIFEST = ROOT / "finetune/manifests/poc.yaml"


def frozen_dataset(tmp_path, monkeypatch, examples):
    """A frozen dataset the run can train from, under a temporary repository root."""
    root = tmp_path / "datasets" / "gemma"
    folder = root / "rendered"
    folder.mkdir(parents=True)
    for student, rows in examples.items():
        (folder / f"{student}.jsonl").write_text("".join(r.to_json() + "\n" for r in rows))
    manifest_path = root / "manifest.yaml"
    manifest_path.write_text("teacher: gemma\n")
    files = [manifest_path, *sorted(folder.glob("*.jsonl"))]
    (root / "freeze.json").write_text(
        json.dumps(
            {
                "teacher": "gemma",
                "attempts": 1,
                "revision": frozen.digest(files),
                "files": [str(p.relative_to(tmp_path)) for p in files],
            }
        )
    )
    monkeypatch.setattr(frozen, "ROOT", tmp_path)
    return root


@pytest.fixture
def template_tokenizers(monkeypatch):
    monkeypatch.setattr(render, "tokenizer", lambda student: TemplateTokenizer(student.family))


def test_poc_manifest_keeps_validation_scenarios_for_evaluation():
    poc = manifest.read(POC_MANIFEST)
    assert (len(poc.train), len(poc.eval)) == (10, 6)
    evals = {trajectories.load(p).scenario_id for p in poc.eval}
    assert evals <= CONFIG.validation_scenarios


def test_manifest_that_trains_on_a_validation_scenario_is_refused(tmp_path):
    validation = manifest.read(POC_MANIFEST).eval[0].relative_to(ROOT)
    path = tmp_path / "manifest.yaml"
    path.write_text(f"train:\n  - {validation}\n")
    with pytest.raises(ValueError, match="validation scenarios"):
        manifest.read(path)


def test_manifest_with_an_unclean_attempt_is_refused(tmp_path):
    attempt = json.loads(ATTEMPT.read_text())
    attempt["agent"]["termination"] = "turn_limit"
    unclean = tmp_path / "attempt.json"
    unclean.write_text(json.dumps(attempt))
    path = tmp_path / "manifest.yaml"
    path.write_text(f"train:\n  - {unclean}\n")
    with pytest.raises(ValueError, match="without full score"):
        manifest.read(path)


def test_variant_names_split_into_model_and_precision():
    assert Variant.parse("gemma-4-e4b-qlora") == Variant("gemma-4-e4b", Precision.QLORA)
    assert Variant.parse("qwen35-9b-lora").name == "qwen35-9b-lora"
    for bad in ("gemma-4-e4b-full", "llama-3-8b-lora"):
        with pytest.raises(ValueError):
            Variant.parse(bad)


def test_job_points_at_the_variants_rendered_data(tmp_path):
    job = prepare.job(Variant("qwen35-4b", Precision.QLORA), tmp_path / "data", tmp_path / "o", 2)
    assert (job.student, job.settings.precision, job.steps) == ("qwen35-4b", "qlora", 2)
    assert job.data == tmp_path / "data" / "qwen35-4b" / "train.jsonl"
    assert job.eval == tmp_path / "data" / "qwen35-4b" / "eval.jsonl"


def test_prepare_renders_only_the_needed_models_and_records_provenance(
    tmp_path, template_tokenizers
):
    variants = [
        Variant("qwen35-4b", Precision.LORA),
        Variant("gemma-4-e4b", Precision.QLORA),
    ]
    session = prepare.prepare(tmp_path, manifest.read(POC_MANIFEST), variants, 2, "H100")

    saved = json.loads((tmp_path / "session.json").read_text())
    assert (saved["steps"], saved["gpu"]) == (2, "H100")
    assert saved["variants"] == ["qwen35-4b-lora", "gemma-4-e4b-qlora"]
    assert saved["manifest"] == "finetune/manifests/poc.yaml"
    assert len(session.provenance.commit) == 40
    for name, digest in session.data.items():
        assert hashlib.sha256((tmp_path / "data" / name).read_bytes()).hexdigest() == digest
    rows = (tmp_path / "data" / "gemma-4-e4b" / "train.jsonl").read_text().splitlines()
    assert len(rows) == 10 and all(json.loads(r)["targets"] for r in rows)
    assert sorted(p.name for p in (tmp_path / "data").iterdir()) == ["gemma-4-e4b", "qwen35-4b"]


def test_prepare_trains_on_the_frozen_rows_and_records_their_revision(
    tmp_path, monkeypatch, template_tokenizers
):
    poc = manifest.read(POC_MANIFEST)
    inputs = Manifest(POC_MANIFEST, [], poc.eval[:1])
    frozen_rows = [Row("frozen.json", "scenario", "frozen text", [(0, 4)])]
    directory = frozen_dataset(tmp_path, monkeypatch, {"gemma-4-e4b": frozen_rows})
    variants = [Variant("gemma-4-e4b", Precision.LORA)]

    session = prepare.prepare(tmp_path / "out", inputs, variants, 1, rows=directory)

    train = (tmp_path / "out" / "data" / "gemma-4-e4b" / "train.jsonl").read_text()
    assert train == frozen_rows[0].to_json() + "\n"
    saved = json.loads((tmp_path / "out" / "session.json").read_text())
    assert saved["data_source"] == str(directory)
    assert saved["data_revision"] == session.data_revision
    assert (tmp_path / "out" / "data" / "gemma-4-e4b" / "eval.jsonl").read_text().strip()


def test_prepare_refuses_frozen_rows_without_the_needed_student(
    tmp_path, monkeypatch, template_tokenizers
):
    poc = manifest.read(POC_MANIFEST)
    inputs = Manifest(POC_MANIFEST, [], poc.eval[:1])
    directory = frozen_dataset(tmp_path, monkeypatch, {"gemma-4-e4b": []})

    with pytest.raises(ValueError, match="no qwen35-4b rows"):
        prepare.prepare(
            tmp_path / "out", inputs, [Variant("qwen35-4b", Precision.LORA)], 1, rows=directory
        )
