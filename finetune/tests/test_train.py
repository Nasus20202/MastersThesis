"""The training loop on CPU with tiny random models of both architectures."""

import json
import re

import pytest

torch = pytest.importorskip("torch")

from finetune import dataset, placement, train  # noqa: E402
from finetune.students import STUDENTS  # noqa: E402
from tests.conftest import tiny_gemma, tiny_qwen  # noqa: E402

# ---------------------------------------------------------------- fixtures


@pytest.fixture(scope="module")
def rows():
    """Training rows of both students rendered from the fixture attempt."""
    from finetune import trajectories
    from tests.conftest import ATTEMPT, TemplateTokenizer

    traj = trajectories.load(ATTEMPT)
    return {
        name: dataset.example(TemplateTokenizer(STUDENTS[name].family), STUDENTS[name], traj)
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
        [r["text"] for r in rows.values()],
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


def window(ids, labels, size):
    """A slice of ``size`` tokens that starts before the first target token."""
    first = next(i for i, label in enumerate(labels) if label != train.IGNORE)
    start = max(0, first - size // 3)
    return torch.tensor(ids[start : start + size]), torch.tensor(labels[start : start + size])


def peft_model(checkpoint, host_tables=True):
    cfg = train.Config(precision="lora")
    model, tok, _ = train.load(str(checkpoint), None, cfg, host_tables=host_tables)
    return model, tok


def run_args(tmp_path, checkpoint, rows, name, *extra):
    data = tmp_path / "train.jsonl"
    data.write_text(json.dumps(rows[name]) + "\n")
    return train.parse_args(
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
    ids, labels = train.tokenize(tokenizer, row)
    enc = tokenizer(row["text"], add_special_tokens=False, return_offsets_mapping=True)
    for (start, end), label in zip(enc["offset_mapping"], labels, strict=True):
        inside = any(a <= start < b for a, b in row["targets"])
        assert (label != train.IGNORE) == (inside and end > start)
    assert len(ids) == len(labels)
    assert 0 < sum(label != train.IGNORE for label in labels) < len(labels)


def test_lora_targets_skip_vision_and_audio_towers():
    pattern = re.compile(train.LORA_TARGET_REGEX)
    assert pattern.match("model.language_model.layers.0.self_attn.q_proj")
    assert pattern.match("model.layers.3.mlp.down_proj")
    assert not pattern.match("model.vision_tower.encoder.layers.0.self_attn.q_proj")
    assert not pattern.match("model.audio_tower.layers.0.mlp.up_proj")
    assert not pattern.match("model.visual.blocks.0.attn.qkv")
    assert not pattern.match("model.layers.0.linear_attn.in_proj_qkvz")


def test_adapters_are_attached_to_language_projections(checkpoint):
    model, _ = peft_model(checkpoint)
    adapted = [
        n
        for n, _ in model.named_modules()
        if n.endswith(train.LORA_MODULES) and hasattr(_, "lora_A")
    ]
    assert adapted
    assert all("vision" not in n and "audio" not in n for n in adapted)
    trainable = [n for n, p in model.named_parameters() if p.requires_grad]
    assert trainable and all("lora_" in n for n in trainable)


def test_chunked_loss_equals_the_models_own_loss(checkpoint, rows, name, monkeypatch):
    """Including Gemma's final logit softcapping, with chunks smaller than the
    number of targets."""
    model, tok = peft_model(checkpoint)
    ids, labels = window(*train.tokenize(tok, rows[name]), 300)
    monkeypatch.setattr(train, "LOSS_CHUNK", 7)
    with torch.no_grad():
        ours, count = train.loss_and_count(model, ids, labels)
        reference = model(input_ids=ids.unsqueeze(0), labels=labels.unsqueeze(0)).loss
    assert count == int((labels[1:] != train.IGNORE).sum()) > 0
    assert torch.allclose(ours, reference, atol=1e-5)


def test_chunked_loss_gradients_equal_full_loss_gradients(checkpoint, rows, name, monkeypatch):
    model, tok = peft_model(checkpoint)
    ids, labels = window(*train.tokenize(tok, rows[name]), 200)
    # Non-zero B so the adapter gradients are not trivially zero.
    for n, p in model.named_parameters():
        if "lora_B" in n:
            torch.nn.init.normal_(p, std=0.02)
    params = [p for p in model.parameters() if p.requires_grad]

    monkeypatch.setattr(train, "LOSS_CHUNK", 5)
    loss, _ = train.loss_and_count(model, ids, labels)
    ours = torch.autograd.grad(loss, params)
    reference = model(input_ids=ids.unsqueeze(0), labels=labels.unsqueeze(0)).loss
    theirs = torch.autograd.grad(reference, params)
    for a, b in zip(ours, theirs, strict=True):
        assert torch.allclose(a, b, atol=1e-6, rtol=1e-4)


def test_sweep_does_not_change_the_adapter(checkpoint, rows, name):
    model, tok = peft_model(checkpoint)
    ids, labels = train.tokenize(tok, rows[name])
    before = {n: p.detach().clone() for n, p in model.named_parameters() if p.requires_grad}
    cfg = train.Config(precision="lora")
    results = train.sweep(model, cfg, ids, labels, [64, 128, 10**7], repeats=2)
    assert [r["status"] for r in results] == ["ok", "ok", "not enough tokens"]
    assert all(r["tokens_per_second"] > 0 for r in results[:2])
    for n, p in model.named_parameters():
        if p.requires_grad:
            assert torch.equal(p, before[n]), n


def test_out_of_memory_in_the_sweep_is_recorded_not_raised(checkpoint, rows, name, monkeypatch):
    model, tok = peft_model(checkpoint)
    ids, labels = train.tokenize(tok, rows[name])

    def oom(self, ids, labels):
        raise torch.OutOfMemoryError("CUDA out of memory")

    monkeypatch.setattr(train.Trainer, "step", oom)
    result = train.sweep(model, train.Config(precision="lora"), ids, labels, [64])[0]
    assert result == {"length": 64, "status": "out of memory", "error": "CUDA out of memory"}


def test_gemma_per_layer_embeddings_are_looked_up_in_host_memory(tmp_path, tokenizer, rows):
    torch.manual_seed(0)
    tiny_gemma(len(tokenizer)).save_pretrained(tmp_path / "model")
    tokenizer.save_pretrained(tmp_path / "model")
    plain, tok = peft_model(tmp_path / "model", host_tables=False)
    hosted, _ = peft_model(tmp_path / "model")

    tables = [m for n, m in hosted.named_modules() if n.endswith(placement.HOST_TABLES)]
    assert len(tables) == 1 and isinstance(tables[0], placement.HostLookup)
    # Not a registered submodule: no parameters, nothing in the state dict.
    assert not any(
        n.endswith(placement.HOST_TABLES[0] + ".weight") for n, _ in hosted.named_parameters()
    )
    assert not [k for k in hosted.state_dict() if placement.HOST_TABLES[0] in k]
    ids, labels = window(*train.tokenize(tok, rows["gemma-4-e4b"]), 200)
    with torch.no_grad():
        assert torch.equal(
            train.loss_and_count(plain, ids, labels)[0],
            train.loss_and_count(hosted, ids, labels)[0],
        )


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
    metrics = train.run(args)

    out = tmp_path / "out"
    saved = json.loads((out / "metrics.json").read_text())
    assert saved["train_examples"]["total"] == saved["train_examples"]["kept"] == 1
    assert [r["status"] for r in saved["sweep"]] == ["ok", "ok"]
    assert saved["max_len"] == 100000
    assert all(e["finite"] and not e["skipped"] for e in metrics["train_log"])
    assert [e["step"] for e in saved["train_log"]] == list(range(1, 9))
    assert saved["skipped_steps"] == 0
    assert set(saved["parameter_dtypes"]) == {"torch.float32"}
    assert "error" not in saved and "train_error" not in saved
    losses = [e["loss"] for e in metrics["train_log"]]
    assert losses[-1] < losses[0]
    # Before training the adapter is the identity (B = 0); after, it helps on
    # the trained example.
    before, after = saved["eval_before"], saved["eval_after"]
    assert before["with_adapter"] == pytest.approx(before["without_adapter"])
    assert after["with_adapter"] < after["without_adapter"]
    assert after["without_adapter"] == pytest.approx(before["without_adapter"], rel=1e-5)
    assert (out / "adapter" / "adapter_config.json").exists()
    assert (out / "adapter" / "adapter_model.safetensors").exists()


def test_examples_longer_than_the_cap_are_dropped_not_truncated(tmp_path, checkpoint, rows, name):
    args = run_args(
        tmp_path, checkpoint, rows, name, "--steps", "1", "--max-len", "128", "--sweep", "64"
    )
    metrics = train.run(args)
    assert metrics["max_len"] == 128
    assert metrics["train_examples"]["kept"] == 0
    assert metrics["train_log"] == []


def test_cap_falls_to_the_longest_length_that_fitted(tmp_path, checkpoint, rows, name, monkeypatch):
    real_step = train.Trainer.step

    def step(self, ids, labels):
        if len(ids) > 100:
            raise torch.OutOfMemoryError("CUDA out of memory")
        return real_step(self, ids, labels)

    monkeypatch.setattr(train.Trainer, "step", step)
    args = run_args(tmp_path, checkpoint, rows, name, "--steps", "1", "--sweep", "64,96,128,256")
    metrics = train.run(args)
    assert [r["status"] for r in metrics["sweep"]] == ["ok", "ok", "out of memory", "out of memory"]
    assert metrics["max_len"] == 96


def test_out_of_memory_in_training_is_recorded_and_the_rest_still_runs(
    tmp_path, checkpoint, rows, name, monkeypatch
):
    real_step = train.Trainer.step
    calls = []

    def step(self, ids, labels):
        if self.base_lr and len(calls) == 2:
            raise torch.OutOfMemoryError("CUDA out of memory")
        if self.base_lr:
            calls.append(len(ids))
        return real_step(self, ids, labels)

    monkeypatch.setattr(train.Trainer, "step", step)
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
    metrics = train.run(args)
    assert len(metrics["train_log"]) == 2
    assert metrics["train_error"] == "CUDA out of memory"
    assert "eval_after" in metrics
    assert (tmp_path / "out" / "adapter" / "adapter_config.json").exists()


def test_a_failure_keeps_what_was_measured(tmp_path, checkpoint, rows, name, monkeypatch):
    def broken(*args, **kwargs):
        raise ValueError("broken evaluation")

    monkeypatch.setattr(train, "evaluate", broken)
    args = run_args(
        tmp_path, checkpoint, rows, name, "--eval", str(tmp_path / "train.jsonl"), "--sweep", "64"
    )
    with pytest.raises(ValueError):
        train.run(args)
    saved = json.loads((tmp_path / "out" / "metrics.json").read_text())
    assert saved["sweep"][0]["status"] == "ok"
    assert saved["error"] == "ValueError: broken evaluation"


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
    results = [{"length": n, "status": s} for n, s in statuses.items()]
    assert train.training_cap(results, 1000) == cap


def test_frozen_matrices_to_keeps_1d_and_trainable_weights_in_float32():
    model = torch.nn.Sequential(
        torch.nn.Embedding(8, 4), torch.nn.LayerNorm(4), torch.nn.Linear(4, 4)
    )
    for p in model[:2].parameters():
        p.requires_grad_(False)
    train.frozen_matrices_to(model, torch.float16)
    assert model[0].weight.dtype == torch.float16
    assert model[1].weight.dtype == torch.float32
    assert model[2].weight.dtype == torch.float32
