"""The real training checkpoints on CPU (opt-in: about 25 GB download, 20 GB RAM).

Run with ``FINETUNE_REAL_CHECKPOINTS=<adapter output dir> pytest
tests/test_real_checkpoints.py``. Checks that each pinned checkpoint loads,
that adapters sit only on language-model projections, and that one training
step runs and saves an adapter (used to test the GGUF
conversion before any Kaggle run).
"""

import os
from pathlib import Path

import pytest

torch = pytest.importorskip("torch")

from finetune import dataset, render, train, trajectories  # noqa: E402
from finetune.students import STUDENTS  # noqa: E402
from tests.conftest import ATTEMPT  # noqa: E402

OUT = os.environ.get("FINETUNE_REAL_CHECKPOINTS")
pytestmark = pytest.mark.skipif(not OUT, reason="set FINETUNE_REAL_CHECKPOINTS to run")


@pytest.mark.parametrize("name", sorted(STUDENTS))
def test_checkpoint_trains_one_step_and_saves_an_adapter(name):
    student = STUDENTS[name]
    cfg = train.Config(precision="lora")
    model, tok, _ = train.load(student.checkpoint, student.revision, cfg, dtype=torch.bfloat16)

    adapted = train.adapted_modules(model)
    assert adapted
    assert not [n for n in adapted if "vision" in n or "audio" in n or "visual" in n]
    layers = model.config.get_text_config().num_hidden_layers
    print(name, "adapted modules", len(adapted), "layers", layers, adapted[:3])
    print(name, "trainable parameters", train.trainable_parameters(model))

    row = dataset.example(render.tokenizer(student), student, trajectories.load(ATTEMPT))
    ids, labels = train.tokenize(tok, row)
    first = next(i for i, label in enumerate(labels) if label != train.IGNORE)
    window = slice(max(0, first - 128), first + 384)
    trainer = train.Trainer(model, cfg)
    log = trainer.step(ids[window], labels[window])
    assert log["target_tokens"] > 0 and log["loss"] > 0
    print(name, log)
    model.save_pretrained(Path(OUT) / name)
    assert (Path(OUT) / name / "adapter_model.safetensors").exists()
