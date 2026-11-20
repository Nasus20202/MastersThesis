"""The real training checkpoints on CPU (opt-in: about 25 GB download, 20 GB RAM).

Run with ``FINETUNE_REAL_CHECKPOINTS=<adapter output dir> pytest
tests/test_real_checkpoints.py``. Checks that each pinned checkpoint loads,
that adapters sit only on language-model projections, and that one training
step runs and saves an adapter (used to test the GGUF
conversion before any paid run).
"""

import os
from pathlib import Path

import pytest

torch = pytest.importorskip("torch")

from finetune.config import STUDENTS  # noqa: E402
from finetune.data import render, trajectories  # noqa: E402
from finetune.data.rows import example  # noqa: E402
from finetune.training.lora import adapted_modules, load, trainable_parameters  # noqa: E402
from finetune.training.models import Precision, Settings, Tokens  # noqa: E402
from finetune.training.tokens import IGNORE, tokenize  # noqa: E402
from finetune.training.trainer import Trainer  # noqa: E402
from tests.conftest import ATTEMPT  # noqa: E402

OUT = os.environ.get("FINETUNE_REAL_CHECKPOINTS")
pytestmark = pytest.mark.skipif(not OUT, reason="set FINETUNE_REAL_CHECKPOINTS to run")


@pytest.mark.parametrize("name", sorted(STUDENTS))
def test_checkpoint_trains_one_step_and_saves_an_adapter(name):
    student = STUDENTS[name]
    cfg = Settings(precision=Precision.LORA)
    model, tok = load(student.checkpoint, student.revision, cfg, dtype=torch.bfloat16)

    adapted = adapted_modules(model)
    assert adapted
    assert not [n for n in adapted if "vision" in n or "audio" in n or "visual" in n]
    layers = model.config.get_text_config().num_hidden_layers
    print(name, "adapted modules", len(adapted), "layers", layers, adapted[:3])
    print(name, "trainable parameters", trainable_parameters(model))

    row = example(render.tokenizer(student), student, trajectories.load(ATTEMPT))
    tokens = tokenize(tok, row)
    first = next(i for i, label in enumerate(tokens.labels) if label != IGNORE)
    window = slice(max(0, first - 128), first + 384)
    trainer = Trainer(model, cfg)
    step = trainer.step(Tokens(tokens.ids[window], tokens.labels[window]))
    assert step.target_tokens > 0 and step.loss > 0
    print(name, step)
    model.save_pretrained(Path(OUT) / name)
    assert (Path(OUT) / name / "adapter_model.safetensors").exists()
