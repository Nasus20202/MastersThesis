"""The training loop on CPU with tiny random models of both architectures."""

import json
import re

import pytest

torch = pytest.importorskip("torch")

from finetune.config import STUDENTS  # noqa: E402
from finetune.data.rows import example  # noqa: E402
from finetune.training import loss as chunked_loss  # noqa: E402
from finetune.training import run as runner  # noqa: E402
from finetune.training.gpu import compute_dtype  # noqa: E402
from finetune.training.lora import (  # noqa: E402
    LORA_MODULES,
    LORA_TARGET_REGEX,
    frozen_matrices_to,
    load,
)
from finetune.training.loss import loss_and_count  # noqa: E402
from finetune.training.models import Precision, Settings, SweepPoint, SweepStatus  # noqa: E402
from finetune.training.sweep import sweep, training_cap  # noqa: E402
from finetune.training.tokens import IGNORE, tokenize  # noqa: E402
from finetune.training.trainer import Trainer  # noqa: E402
from tests.conftest import tiny_gemma, tiny_qwen  # noqa: E402

# ---------------------------------------------------------------- fixtures


@pytest.fixture(scope="module")
def rows():
    """Training rows of both students rendered from the fixture attempt."""
    from finetune.data import trajectories
    from tests.conftest import ATTEMPT, TemplateTokenizer

    traj = trajectories.load(ATTEMPT)
    return {
        name: example(TemplateTokenizer(STUDENTS[name].family), STUDENTS[name], traj)
        for name in MODELS
    }


@pytest.fixture(scope="module")
def tokenizer(tmp_path_factory, rows):
    """A small byte-level BPE tokenizer trained on the rendered texts."""
    from tokenizers import Tokenizer, models, pre_tokenizers, trainers
    from transformers import PreTrainedTokenizerFast

    specials = [
        "<bos>",
        "<|im_start|>",
        "<|im_end|>",
        "<|turn>",
        "<turn|>",
        "<|tool_response>",
        "<tool_response|>",
        "<|channel>",
        "<channel|>",
    ]
    tok = Tokenizer(models.BPE())
    tok.pre_tokenizer = pre_tokenizers.ByteLevel(add_prefix_space=False)
    tok.train_from_iterator(
        [r.text for r in rows.values()],
        trainers.BpeTrainer(
            vocab_size=600,
            special_tokens=specials,
            initial_alphabet=pre_tokenizers.ByteLevel.alphabet(),
        ),
    )
    from tokenizers import decoders

    tok.decoder = decoders.ByteLevel()
    fast = PreTrainedTokenizerFast(tokenizer_object=tok, bos_token="<bos>")
    path = tmp_path_factory.mktemp("tok")
    fast.save_pretrained(path)
    return fast


MODELS = {"qwen35-4b": tiny_qwen, "gemma-4-e4b": tiny_gemma}


@pytest.fixture(params=sorted(MODELS))
def name(request):
    return request.param


@pytest.fixture
def checkpoint(tmp_path, name, tokenizer):
    torch.manual_seed(0)
    model = MODELS[name](len(tokenizer))
    model.save_pretrained(tmp_path / "model")
    tokenizer.save_pretrained(tmp_path / "model")
    return tmp_path / "model"


def window(tokens, size):
    """A slice of ``size`` tokens that starts before the first target token."""
    first = next(i for i, label in enumerate(tokens.labels) if label != IGNORE)
    start = max(0, first - size // 3)
    return (
        torch.tensor(tokens.ids[start : start + size]),
        torch.tensor(tokens.labels[start : start + size]),
    )


def peft_model(checkpoint):
    cfg = Settings(precision=Precision.LORA)
    model, tok = load(str(checkpoint), None, cfg)
    return model, tok


def run_args(tmp_path, checkpoint, rows, name, *extra):
    data = tmp_path / "train.jsonl"
    data.write_text(rows[name].to_json() + "\n")
    return runner.parse_args(
        [
            "--checkpoint",
            str(checkpoint),
            "--precision",
            "lora",
            "--data",
            str(data),
            "--out",
            str(tmp_path / "out"),
            *extra,
        ]
    )


# ---------------------------------------------------------------- tests


def test_labels_cover_exactly_the_target_spans(rows, tokenizer, name):
    row = rows[name]
    tokens = tokenize(tokenizer, row)
    enc = tokenizer(row.text, add_special_tokens=False, return_offsets_mapping=True)
    for (start, end), label in zip(enc["offset_mapping"], tokens.labels, strict=True):
        inside = any(a <= start < b for a, b in row.targets)
        assert (label != IGNORE) == (inside and end > start)
    assert len(tokens.ids) == len(tokens.labels)
    assert 0 < sum(label != IGNORE for label in tokens.labels) < len(tokens.labels)


def test_lora_targets_skip_vision_and_audio_towers():
    pattern = re.compile(LORA_TARGET_REGEX)
    assert pattern.match("model.language_model.layers.0.self_attn.q_proj")
    assert pattern.match("model.layers.3.mlp.down_proj")
    assert not pattern.match("model.vision_tower.encoder.layers.0.self_attn.q_proj")
    assert not pattern.match("model.audio_tower.layers.0.mlp.up_proj")
    assert not pattern.match("model.visual.blocks.0.attn.qkv")
    assert pattern.match("model.layers.0.linear_attn.in_proj_qkv")
    assert pattern.match("model.layers.0.linear_attn.in_proj_z")
    assert not pattern.match("model.layers.0.linear_attn.out_proj")


def test_adapters_are_attached_to_language_projections(checkpoint):
    model, _ = peft_model(checkpoint)
    adapted = [
        n for n, _ in model.named_modules() if n.endswith(LORA_MODULES) and hasattr(_, "lora_A")
    ]
    assert adapted
    assert all("vision" not in n and "audio" not in n for n in adapted)
    trainable = [n for n, p in model.named_parameters() if p.requires_grad]
    assert trainable and all("lora_" in n for n in trainable)


def test_chunked_loss_equals_the_models_own_loss(checkpoint, rows, name, monkeypatch):
    """Including Gemma's final logit softcapping, with chunks smaller than the
    number of targets."""
    model, tok = peft_model(checkpoint)
    ids, labels = window(tokenize(tok, rows[name]), 300)
    monkeypatch.setattr(chunked_loss, "LOSS_CHUNK", 7)
    with torch.no_grad():
        ours, count = loss_and_count(model, ids, labels)
        reference = model(input_ids=ids.unsqueeze(0), labels=labels.unsqueeze(0)).loss
    assert count == int((labels[1:] != IGNORE).sum()) > 0
    assert torch.allclose(ours, reference, atol=1e-5)


def test_chunked_loss_gradients_equal_full_loss_gradients(checkpoint, rows, name, monkeypatch):
    model, tok = peft_model(checkpoint)
    ids, labels = window(tokenize(tok, rows[name]), 200)
    # Non-zero B so the adapter gradients are not trivially zero.
    for n, p in model.named_parameters():
        if "lora_B" in n:
            torch.nn.init.normal_(p, std=0.02)
    params = [p for p in model.parameters() if p.requires_grad]

    monkeypatch.setattr(chunked_loss, "LOSS_CHUNK", 5)
    loss, _ = loss_and_count(model, ids, labels)
    ours = torch.autograd.grad(loss, params)
    reference = model(input_ids=ids.unsqueeze(0), labels=labels.unsqueeze(0)).loss
    theirs = torch.autograd.grad(reference, params)
    for a, b in zip(ours, theirs, strict=True):
        assert torch.allclose(a, b, atol=1e-6, rtol=1e-4)


def test_sweep_does_not_change_the_adapter(checkpoint, rows, name):
    model, tok = peft_model(checkpoint)
    tokens = tokenize(tok, rows[name])
    before = {n: p.detach().clone() for n, p in model.named_parameters() if p.requires_grad}
    cfg = Settings(precision=Precision.LORA)
    results = sweep(model, cfg, tokens, [64, 128, 10**7], repeats=2)
    assert [r.status for r in results] == ["ok", "ok", "not enough tokens"]
    assert all(r.tokens_per_second > 0 for r in results[:2])
    for n, p in model.named_parameters():
        if p.requires_grad:
            assert torch.equal(p, before[n]), n


def test_out_of_memory_in_the_sweep_is_recorded_not_raised(checkpoint, rows, name, monkeypatch):
    model, tok = peft_model(checkpoint)
    tokens = tokenize(tok, rows[name])

    def oom(self, example):
        raise torch.OutOfMemoryError("CUDA out of memory")

    monkeypatch.setattr(Trainer, "step", oom)
    result = sweep(model, Settings(precision=Precision.LORA), tokens, [64])[0]
    assert result == SweepPoint(64, SweepStatus.OUT_OF_MEMORY, error="CUDA out of memory")


def test_run_trains_evaluates_and_saves(tmp_path, checkpoint, rows, name):
    args = run_args(
        tmp_path,
        checkpoint,
        rows,
        name,
        "--eval",
        str(tmp_path / "train.jsonl"),
        "--steps",
        "8",
        "--max-len",
        "100000",
        "--sweep",
        "128,256",
        "--learning-rate",
        "3e-3",
    )
    metrics = runner.run(args)

    out = tmp_path / "out"
    saved = json.loads((out / "metrics.json").read_text())
    assert saved["train_examples"]["total"] == saved["train_examples"]["kept"] == 1
    assert [r["status"] for r in saved["sweep"]] == ["ok", "ok"]
    assert saved["max_len"] == 100000
    assert all(e.finite for e in metrics.train_log)
    assert all(e.grad_norm is not None and e.grad_norm > 0 for e in metrics.train_log)
    assert [e["step"] for e in saved["train_log"]] == list(range(1, 9))
    assert set(saved["parameter_dtypes"]) == {"torch.float32"}
    assert saved["error"] is None and saved["train_error"] is None
    assert saved["host_peak_gib"] > 0
    losses = [e.loss for e in metrics.train_log]
    assert losses[-1] < losses[0]
    # Before training the adapter is the identity (B = 0); after, it helps on
    # the trained example.
    before, after = saved["eval_before"], saved["eval_after"]
    assert before["with_adapter"] == pytest.approx(before["without_adapter"])
    assert after["with_adapter"] < after["without_adapter"]
    assert after["without_adapter"] == pytest.approx(before["without_adapter"], rel=1e-5)
    # One evaluation example: its own loss is the mean, tagged with its source.
    (example,) = after["per_example"]
    assert example["source"] == rows[name].source != ""
    assert example["with_adapter"] == pytest.approx(after["with_adapter"])
    assert example["target_tokens"] > 0
    assert (out / "adapter" / "adapter_config.json").exists()
    assert (out / "adapter" / "adapter_model.safetensors").exists()
    log = (out / "log.txt").read_text()
    assert "sweep: 128 tokens fit" in log and "step 8/8" in log and "adapter saved" in log


def test_examples_longer_than_the_cap_are_dropped_not_truncated(tmp_path, checkpoint, rows, name):
    args = run_args(
        tmp_path, checkpoint, rows, name, "--steps", "1", "--max-len", "128", "--sweep", "64"
    )
    metrics = runner.run(args)
    assert metrics.max_len == 128
    assert metrics.train_examples.kept == 0
    assert metrics.train_log == []


def test_epochs_set_the_steps_and_each_but_the_last_is_evaluated(tmp_path, checkpoint, rows, name):
    args = run_args(
        tmp_path,
        checkpoint,
        rows,
        name,
        "--eval",
        str(tmp_path / "train.jsonl"),
        "--epochs",
        "3",
        "--max-len",
        "100000",
        "--sweep",
        "64",
    )
    metrics = runner.run(args)
    assert metrics.train_examples.kept == 1
    assert metrics.steps == 3
    assert [e.step for e in metrics.train_log] == [1, 2, 3]
    assert [(e.epoch, e.step) for e in metrics.eval_epochs] == [(1, 1), (2, 2)]
    saved = json.loads((tmp_path / "out" / "metrics.json").read_text())
    assert len(saved["eval_epochs"]) == 2 and saved["eval_after"] is not None


def test_cap_falls_to_the_longest_length_that_fitted(tmp_path, checkpoint, rows, name, monkeypatch):
    real_step = Trainer.step

    def step(self, example):
        if len(example.ids) > 100:
            raise torch.OutOfMemoryError("CUDA out of memory")
        return real_step(self, example)

    monkeypatch.setattr(Trainer, "step", step)
    args = run_args(tmp_path, checkpoint, rows, name, "--steps", "1", "--sweep", "64,96,128,256")
    metrics = runner.run(args)
    assert [r.status for r in metrics.sweep] == ["ok", "ok", "out of memory", "out of memory"]
    assert metrics.max_len == 96


def test_out_of_memory_in_training_is_recorded_and_the_rest_still_runs(
    tmp_path, checkpoint, rows, name, monkeypatch
):
    real_step = Trainer.step
    calls = []

    def step(self, example):
        if self.base_lr and len(calls) == 2:
            raise torch.OutOfMemoryError("CUDA out of memory")
        if self.base_lr:
            calls.append(len(example.ids))
        return real_step(self, example)

    monkeypatch.setattr(Trainer, "step", step)
    args = run_args(
        tmp_path,
        checkpoint,
        rows,
        name,
        "--eval",
        str(tmp_path / "train.jsonl"),
        "--steps",
        "5",
        "--max-len",
        "100000",
        "--sweep",
        "64",
    )
    metrics = runner.run(args)
    assert len(metrics.train_log) == 2
    assert metrics.train_error == "CUDA out of memory"
    assert metrics.eval_after is not None
    assert (tmp_path / "out" / "adapter" / "adapter_config.json").exists()


def test_a_failure_keeps_what_was_measured(tmp_path, checkpoint, rows, name, monkeypatch):
    def broken(*args, **kwargs):
        raise ValueError("broken evaluation")

    monkeypatch.setattr(runner, "evaluate", broken)
    args = run_args(
        tmp_path, checkpoint, rows, name, "--eval", str(tmp_path / "train.jsonl"), "--sweep", "64"
    )
    with pytest.raises(ValueError):
        runner.run(args)
    saved = json.loads((tmp_path / "out" / "metrics.json").read_text())
    assert saved["sweep"][0]["status"] == "ok"
    assert saved["error"] == "ValueError: broken evaluation"
    assert "run failed" in (tmp_path / "out" / "log.txt").read_text()


@pytest.mark.parametrize(
    "statuses,cap",
    [
        ({64: "ok", 128: "ok"}, 1000),
        ({64: "ok", 128: "out of memory", 256: "ok"}, 64),
        ({64: "out of memory"}, 0),
        ({64: "ok", 128: "not enough tokens"}, 1000),
    ],
)
def test_training_cap_is_the_longest_length_below_the_first_failure(statuses, cap):
    results = [SweepPoint(n, s) for n, s in statuses.items()]
    assert training_cap(results, 1000) == cap


def test_frozen_matrices_to_keeps_1d_and_trainable_weights_in_float32():
    model = torch.nn.Sequential(
        torch.nn.Embedding(8, 4), torch.nn.LayerNorm(4), torch.nn.Linear(4, 4)
    )
    for p in model[:2].parameters():
        p.requires_grad_(False)
    frozen_matrices_to(model, torch.bfloat16)
    assert model[0].weight.dtype == torch.bfloat16
    assert model[1].weight.dtype == torch.float32
    assert model[2].weight.dtype == torch.float32


def test_compute_dtype_is_bfloat16_on_the_gpu_and_float32_on_the_cpu(monkeypatch):
    monkeypatch.setattr(torch.cuda, "is_available", lambda: False)
    assert compute_dtype() == torch.float32
    monkeypatch.setattr(torch.cuda, "is_available", lambda: True)
    assert compute_dtype() == torch.bfloat16
