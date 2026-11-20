"""LoRA training of one model with Hugging Face transformers + PEFT.

One run, in order (``run``):

1. **Load** the pinned checkpoint with a fresh LoRA adapter, split over all
   visible GPUs (``finetune.placement``). Weights are float16 with float16
   autocast and loss scaling (the T4 has no bfloat16); ``qlora`` stores the
   decoder projections in 4-bit NF4 (bitsandbytes) and the norms in float32.
   On the CPU (tests) everything is float32.
2. **Sweep**: peak GPU memory and speed of a training step at each sequence
   length, with learning rate 0, so the adapter is unchanged.
3. **Cap**: training uses ``--max-len``, or the longest swept length that
   fitted if a longer one ran out of memory. Longer trajectories are dropped,
   never cut.
4. **Evaluate** the target-token loss on the held-out trajectories, with and
   without the adapter.
5. **Train** ``--steps`` steps, one trajectory per step, over shuffled epochs.
6. **Evaluate** again and save the adapter.

The loss is the mean cross-entropy over target tokens (``finetune.dataset``),
computed from the last hidden state in chunks of target positions, each chunk
recomputed in the backward pass, so the full logits matrix (sequence x
vocabulary, 262K for Gemma) is never held in memory.

Outputs in ``--out``: ``metrics.json``, rewritten after every phase so that a
failure keeps what was measured, and ``adapter/``.
"""

from __future__ import annotations

import argparse
import contextlib
import gc
import json
import logging
import platform
import random
import time
import traceback
from collections.abc import Callable
from dataclasses import asdict, dataclass
from importlib import metadata
from pathlib import Path

import torch
import torch.nn.functional as F
from torch.utils.checkpoint import checkpoint

from finetune import placement
from finetune.students import STUDENTS

LORA_MODULES = ("q_proj", "k_proj", "v_proj", "o_proj", "gate_proj", "up_proj", "down_proj")
# Language-model projections only: vision and audio towers stay frozen and
# without adapters (the served GGUF has no towers).
LORA_TARGET_REGEX = r"^(?!.*(?:vision|audio|visual)).*\.(?:" + "|".join(LORA_MODULES) + r")$"
LOSS_CHUNK = 256
IGNORE = -100
SWEEP = [4096, 8192, 12288, 16384]
MAX_LEN = 16384

log = logging.getLogger(__name__)

Example = tuple[list[int], list[int]]  # token ids, labels


@dataclass(frozen=True)
class Config:
    precision: str  # "lora" or "qlora"
    rank: int = 16
    alpha: int = 16
    learning_rate: float = 1e-4
    warmup_steps: int = 3
    seed: int = 0


# ---------------------------------------------------------------- data


def tokenize(tok, row: dict) -> Example:
    """Token ids and labels; labels are IGNORE outside target spans."""
    enc = tok(row["text"], add_special_tokens=False, return_offsets_mapping=True)
    ids = enc["input_ids"]
    labels = [IGNORE] * len(ids)
    spans = sorted(tuple(s) for s in row["targets"])
    j = 0
    for i, (start, end) in enumerate(enc["offset_mapping"]):
        if end <= start:
            continue
        while j < len(spans) and spans[j][1] <= start:
            j += 1
        if j < len(spans) and spans[j][0] <= start < spans[j][1]:
            labels[i] = ids[i]
    return ids, labels


def read_examples(tok, path: Path) -> list[Example]:
    rows = [json.loads(line) for line in path.read_text().splitlines() if line.strip()]
    return [tokenize(tok, row) for row in rows]


# ---------------------------------------------------------------- model


def load(
    checkpoint_path: str,
    revision: str | None,
    cfg: Config,
    max_len: int = 0,
    host_tables: bool = True,
    dtype: torch.dtype | None = None,
):
    """The model with a fresh LoRA adapter, its tokenizer and its GPU placement
    (planned for sequences of ``max_len`` tokens). Weights are float16 on the
    GPUs and float32 on the CPU unless ``dtype`` is given."""
    from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
    from transformers import AutoConfig, AutoModelForCausalLM, AutoTokenizer, BitsAndBytesConfig

    cuda = torch.cuda.is_available()
    dtype = dtype or (torch.float16 if cuda else torch.float32)
    kwargs: dict = {"revision": revision, "dtype": dtype}
    if cfg.precision == "qlora":
        kwargs["quantization_config"] = BitsAndBytesConfig(
            load_in_4bit=True,
            bnb_4bit_quant_type="nf4",
            bnb_4bit_use_double_quant=True,
            bnb_4bit_compute_dtype=dtype,
            llm_int8_skip_modules=placement.NOT_QUANTIZED,
        )
    device_map: dict[str, int] = {}
    if cuda:
        config = AutoConfig.from_pretrained(checkpoint_path, revision=revision)
        with torch.device("meta"):
            skeleton = AutoModelForCausalLM.from_config(config)
        device_map = placement.device_map(
            skeleton, cfg.precision, torch.cuda.device_count(), max_len
        )
        log.info("placement for %d tokens: %s", max_len, placement.summary(device_map))
        # Host tables load onto GPU 0 and leave it right after loading.
        kwargs["device_map"] = device_map
    log.info("loading %s (%s, %s)", checkpoint_path, cfg.precision, dtype)
    model = AutoModelForCausalLM.from_pretrained(checkpoint_path, **kwargs)
    log.info("loaded; GPU memory %s", gib(allocated_gib()))
    if host_tables and (names := placement.tables_to_host(model)):
        log.info("host tables %s; GPU memory %s", names, gib(allocated_gib()))
    if cfg.precision == "qlora":
        model = prepare_model_for_kbit_training(
            model,
            use_gradient_checkpointing=True,
            gradient_checkpointing_kwargs={"use_reentrant": False},
        )
        frozen_matrices_to(model, dtype)
    model.gradient_checkpointing_enable(gradient_checkpointing_kwargs={"use_reentrant": False})
    model.enable_input_require_grads()
    model = get_peft_model(
        model,
        LoraConfig(
            r=cfg.rank,
            lora_alpha=cfg.alpha,
            lora_dropout=0.0,
            bias="none",
            target_modules=LORA_TARGET_REGEX,
            task_type="CAUSAL_LM",
        ),
    )
    tok = AutoTokenizer.from_pretrained(checkpoint_path, revision=revision)
    return model, tok, device_map


def frozen_matrices_to(model, dtype: torch.dtype) -> None:
    """Casts frozen float32 matrices (embeddings, output head, towers) to
    ``dtype``. The 4-bit preparation casts every weight that is not 4-bit to
    float32; only the small 1-D weights (norms, biases) need it, and Qwen3.5
    9B's embeddings and head alone would take 8 GB in float32."""
    for p in model.parameters():
        if not p.requires_grad and p.dtype == torch.float32 and p.dim() >= 2:
            p.data = p.data.to(dtype)


def input_device(model) -> torch.device:
    return model.get_input_embeddings().weight.device


def adapted_modules(model) -> list[str]:
    return sorted(
        {n.rsplit(".lora_A", 1)[0] for n, _ in model.named_parameters() if ".lora_A" in n}
    )


def trainable_parameters(model) -> int:
    return sum(p.numel() for p in model.parameters() if p.requires_grad)


# ---------------------------------------------------------------- loss


def _softcap(model) -> float | None:
    config = model.config
    for c in (config, getattr(config, "text_config", None)):
        cap = getattr(c, "final_logit_softcapping", None) if c is not None else None
        if cap:
            return float(cap)
    return None


def _chunk_loss(head, softcap, hidden, labels):
    logits = head(hidden).float()
    if softcap:
        logits = torch.tanh(logits / softcap) * softcap
    return F.cross_entropy(logits, labels.to(logits.device), reduction="sum")


def loss_and_count(model, ids: torch.Tensor, labels: torch.Tensor) -> tuple[torch.Tensor, int]:
    """Mean next-token loss over target tokens, and their number."""
    base = model.get_base_model()
    head = base.get_output_embeddings()
    hidden = base.model(input_ids=ids.unsqueeze(0)).last_hidden_state[0]
    shifted = labels[1:].to(hidden.device)
    positions = (shifted != IGNORE).nonzero(as_tuple=True)[0]
    hidden = hidden[:-1][positions]
    targets = shifted[positions]
    softcap = _softcap(model)
    # Zero but attached to the graph, so a window without targets still
    # backpropagates (memory sweep) instead of failing.
    total = hidden.sum().float() * 0.0
    for start in range(0, len(positions), LOSS_CHUNK):
        h = hidden[start : start + LOSS_CHUNK]
        t = targets[start : start + LOSS_CHUNK]
        if torch.is_grad_enabled():
            total = total + checkpoint(_chunk_loss, head, softcap, h, t, use_reentrant=False)
        else:
            total = total + _chunk_loss(head, softcap, h, t)
    count = len(positions)
    return total / max(count, 1), count


# ---------------------------------------------------------------- GPU memory


def _autocast():
    if torch.cuda.is_available():
        return torch.autocast("cuda", dtype=torch.float16)
    return contextlib.nullcontext()


def _sync() -> None:
    if torch.cuda.is_available():
        torch.cuda.synchronize()


def peak_gib() -> list[float]:
    """Peak allocated memory per GPU since the last reset."""
    return [torch.cuda.max_memory_allocated(i) / 2**30 for i in range(torch.cuda.device_count())]


def allocated_gib() -> list[float]:
    return [torch.cuda.memory_allocated(i) / 2**30 for i in range(torch.cuda.device_count())]


def gib(values: list[float]) -> str:
    """GiB per GPU, for the log: ``3.1 / 4.2 GiB``."""
    return " / ".join(f"{v:.1f}" for v in values) + " GiB" if values else "no GPU"


def reset_peak() -> None:
    gc.collect()
    if torch.cuda.is_available():
        torch.cuda.empty_cache()
        for i in range(torch.cuda.device_count()):
            torch.cuda.reset_peak_memory_stats(i)


def is_oom(error: BaseException) -> bool:
    return isinstance(error, torch.OutOfMemoryError) or "out of memory" in str(error).lower()


def oom_message(error: BaseException) -> str:
    return str(error).splitlines()[0][:300]


# ---------------------------------------------------------------- steps


class Trainer:
    """AdamW with linear warm-up and float16 loss scaling; one example per step."""

    def __init__(self, model, cfg: Config, learning_rate: float | None = None):
        self.model = model
        self.cfg = cfg
        self.device = input_device(model)
        params = [p for p in model.parameters() if p.requires_grad]
        self.base_lr = cfg.learning_rate if learning_rate is None else learning_rate
        self.optimizer = torch.optim.AdamW(params, lr=self.base_lr, weight_decay=0.0)
        # A small initial scale: the default (65536) overflows in the first
        # steps and skips them, which matters in short runs.
        self.scaler = torch.amp.GradScaler(
            "cuda", init_scale=2.0**12, enabled=torch.cuda.is_available()
        )
        self.steps = 0

    def step(self, ids: list[int], labels: list[int]) -> dict:
        self.model.train()
        warm = min(1.0, (self.steps + 1) / max(self.cfg.warmup_steps, 1))
        for group in self.optimizer.param_groups:
            group["lr"] = self.base_lr * warm
        ids_t = torch.tensor(ids, device=self.device)
        labels_t = torch.tensor(labels, device=self.device)
        _sync()
        started = time.perf_counter()
        try:
            with _autocast():
                loss, count = loss_and_count(self.model, ids_t, labels_t)
            self.scaler.scale(loss).backward()
            scale = self.scaler.get_scale()
            self.scaler.step(self.optimizer)
            self.scaler.update()
        finally:
            # Also after running out of memory, so the next attempt starts clean.
            self.optimizer.zero_grad(set_to_none=True)
        # The scaler lowers its scale only when it skipped the step (inf/NaN).
        skipped = self.scaler.is_enabled() and self.scaler.get_scale() < scale
        _sync()
        self.steps += 1
        return {
            "loss": float(loss.detach()),
            "tokens": len(ids),
            "target_tokens": count,
            "seconds": time.perf_counter() - started,
            "skipped": bool(skipped),
            "finite": bool(torch.isfinite(loss.detach())),
        }


def sweep(
    model, cfg: Config, ids: list[int], labels: list[int], lengths: list[int], repeats: int = 3
) -> list[dict]:
    """Peak memory and speed of a training step at each sequence length.

    Runs with learning rate 0 on a separate optimizer, so the adapter is not
    changed by the measurement. Speed is the median of the repeats after the
    first.
    """
    results = []
    for length in lengths:
        if len(ids) < length:
            results.append({"length": length, "status": "not enough tokens"})
            continue
        log.info("sweep: %d tokens", length)
        trainer = Trainer(model, cfg, learning_rate=0.0)
        reset_peak()
        try:
            logs = [trainer.step(ids[:length], labels[:length]) for _ in range(repeats)]
        except (RuntimeError, torch.OutOfMemoryError) as error:
            if not is_oom(error):
                raise
            log.warning("sweep: %d tokens out of memory", length)
            results.append(
                {"length": length, "status": "out of memory", "error": oom_message(error)}
            )
            continue
        finally:
            del trainer
        timed = [entry["seconds"] for entry in logs[1:]] or [logs[0]["seconds"]]
        seconds = sorted(timed)[len(timed) // 2]
        log.info(
            "sweep: %d tokens fit, peak %s, %.1f s/step, %.0f tokens/s",
            length,
            gib(peak_gib()),
            seconds,
            length / seconds,
        )
        results.append(
            {
                "length": length,
                "status": "ok",
                "peak_gib": peak_gib(),
                "seconds_per_step": seconds,
                "tokens_per_second": length / seconds,
                "loss": logs[0]["loss"],
            }
        )
    return results


def training_cap(results: list[dict], max_len: int) -> int:
    """``max_len``, or the longest swept length that fitted below the shortest
    one that ran out of memory (0 if none fitted)."""
    failed = [r["length"] for r in results if r["status"] == "out of memory"]
    if not failed:
        return max_len
    fitting = [r["length"] for r in results if r["status"] == "ok" and r["length"] < min(failed)]
    return min(max_len, max(fitting)) if fitting else 0


@torch.no_grad()
def evaluate(model, examples: list[Example]) -> dict:
    """Target-token loss on held-out examples, with and without the adapter."""
    model.eval()
    device = input_device(model)

    def mean_loss() -> float:
        total = count = 0.0
        for ids, labels in examples:
            with _autocast():
                loss, n = loss_and_count(
                    model, torch.tensor(ids, device=device), torch.tensor(labels, device=device)
                )
            total += float(loss) * n
            count += n
        return total / max(count, 1)

    log.info("evaluating on %d examples", len(examples))
    with_adapter = mean_loss()
    with model.disable_adapter():
        without_adapter = mean_loss()
    log.info(
        "evaluation loss %.4f with the adapter, %.4f without; peak %s",
        with_adapter,
        without_adapter,
        gib(peak_gib()),
    )
    return {
        "examples": len(examples),
        "with_adapter": with_adapter,
        "without_adapter": without_adapter,
    }


def train_steps(
    trainer: Trainer, examples: list[Example], steps: int, seed: int, on_step: Callable
) -> None:
    """``steps`` steps over shuffled epochs of ``examples``."""
    order = random.Random(seed)
    done = 0
    while examples and done < steps:
        epoch = list(examples)
        order.shuffle(epoch)
        for ids, labels in epoch[: steps - done]:
            done += 1
            on_step({"step": done, **trainer.step(ids, labels)})


# ---------------------------------------------------------------- run


def environment() -> dict:
    packages = {}
    for name in (
        "torch",
        "transformers",
        "peft",
        "bitsandbytes",
        "accelerate",
        "flash-linear-attention",
        "triton",
    ):
        with contextlib.suppress(metadata.PackageNotFoundError):
            packages[name] = metadata.version(name)
    env: dict = {"python": platform.python_version(), "packages": packages, "gpus": []}
    if torch.cuda.is_available():
        env["cuda"] = torch.version.cuda
        for i in range(torch.cuda.device_count()):
            props = torch.cuda.get_device_properties(i)
            env["gpus"].append({"name": props.name, "memory_gib": props.total_memory / 2**30})
    return env


def run(args: argparse.Namespace) -> dict:
    """One training run (module docstring); returns ``metrics.json``'s content."""
    cfg = Config(
        precision=args.precision,
        rank=args.rank,
        alpha=args.rank,
        learning_rate=args.learning_rate,
        seed=args.seed,
    )
    random.seed(cfg.seed)
    torch.manual_seed(cfg.seed)
    args.out.mkdir(parents=True, exist_ok=True)
    metrics: dict = {
        "args": {k: str(v) for k, v in vars(args).items()},
        "config": asdict(cfg),
        "environment": environment(),
    }

    def save() -> None:
        (args.out / "metrics.json").write_text(json.dumps(metrics, indent=2))

    try:
        _run(args, cfg, metrics, save)
    except BaseException as error:
        metrics["error"] = "".join(traceback.format_exception_only(error)).strip()[-2000:]
        raise
    finally:
        save()
    return metrics


def _run(args: argparse.Namespace, cfg: Config, metrics: dict, save: Callable) -> None:
    if args.student:
        student = STUDENTS[args.student]
        checkpoint_path, revision = student.checkpoint, student.revision
    else:
        checkpoint_path, revision = str(args.checkpoint), None

    # 1. Load.
    started = time.perf_counter()
    reset_peak()
    model, tok, device_map = load(
        checkpoint_path, revision, cfg, max_len=max([args.max_len, *args.sweep])
    )
    dtypes: dict[str, int] = {}
    for p in model.parameters():
        dtypes[str(p.dtype)] = dtypes.get(str(p.dtype), 0) + p.numel()
    adapted = adapted_modules(model)
    metrics.update(
        {
            "placement": placement.summary(device_map),
            "host_tables": [
                n for n, m in model.named_modules() if isinstance(m, placement.HostLookup)
            ],
            "load_seconds": time.perf_counter() - started,
            "load_peak_gib": peak_gib(),
            "weights_gib": allocated_gib(),
            "parameter_dtypes": dtypes,
            "trainable_parameters": trainable_parameters(model),
            "adapted_module_count": len(adapted),
            "adapted_modules": adapted[:4],
        }
    )
    log.info(
        "ready in %.0f s: %d trainable parameters in %d modules, load peak %s",
        metrics["load_seconds"],
        metrics["trainable_parameters"],
        len(adapted),
        gib(peak_gib()),
    )
    save()

    train = read_examples(tok, args.data)
    evals = read_examples(tok, args.eval) if args.eval else []

    # 2. Sweep, on all training tokens concatenated so every length has real text.
    joined_ids = [t for ids, _ in train for t in ids]
    joined_labels = [t for _, labels in train for t in labels]
    metrics["sweep"] = sweep(model, cfg, joined_ids, joined_labels, sorted(args.sweep))

    # 3. Cap: longer trajectories are dropped, never cut (#24).
    max_len = training_cap(metrics["sweep"], args.max_len)
    kept = [ex for ex in train if len(ex[0]) <= max_len]
    kept_evals = [ex for ex in evals if len(ex[0]) <= max_len]
    metrics["max_len"] = max_len
    for key, total, kept_now in (
        ("train_examples", train, kept),
        ("eval_examples", evals, kept_evals),
    ):
        metrics[key] = {
            "total": len(total),
            "kept": len(kept_now),
            "tokens": sorted(len(ids) for ids, _ in total),
        }
    log.info(
        "cap %d tokens: %d of %d training and %d of %d evaluation trajectories kept",
        max_len,
        len(kept),
        len(train),
        len(kept_evals),
        len(evals),
    )
    save()

    # 4. Evaluate before training.
    if kept_evals:
        reset_peak()
        metrics["eval_before"] = evaluate(model, kept_evals)
        save()

    # 5. Train.
    trainer = Trainer(model, cfg)
    steps: list[dict] = []
    metrics["train_log"] = steps

    def on_step(entry: dict) -> None:
        steps.append(entry)
        log.info(
            "step %d/%d: loss %.4f, %d tokens, %.1f s, peak %s%s",
            entry["step"],
            args.steps,
            entry["loss"],
            entry["tokens"],
            entry["seconds"],
            gib(peak_gib()),
            " (skipped: inf/NaN gradients)" if entry["skipped"] else "",
        )
        save()

    reset_peak()
    try:
        train_steps(trainer, kept, args.steps, cfg.seed, on_step)
    except (RuntimeError, torch.OutOfMemoryError) as error:
        if not is_oom(error):
            raise
        # Recorded, not raised: the steps so far and the evaluation still count.
        metrics["train_error"] = oom_message(error)
        log.warning("training stopped after %d steps: out of memory", len(steps))
    tokens = sum(e["tokens"] for e in steps)
    seconds = sum(e["seconds"] for e in steps)
    metrics.update(
        {
            "train_peak_gib": peak_gib(),
            "skipped_steps": sum(e["skipped"] for e in steps),
            "train_tokens_per_second": tokens / seconds if seconds else None,
            "train_seconds": seconds,
        }
    )
    save()

    # 6. Evaluate after training, save the adapter.
    if kept_evals:
        reset_peak()
        metrics["eval_after"] = evaluate(model, kept_evals)
    model.save_pretrained(args.out / "adapter")
    log.info("adapter saved to %s", args.out / "adapter")


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--student", choices=sorted(STUDENTS))
    source.add_argument("--checkpoint", type=Path, help="local model directory (tests)")
    parser.add_argument("--precision", choices=("lora", "qlora"), required=True)
    parser.add_argument(
        "--data", type=Path, required=True, help="training JSONL (finetune.dataset)"
    )
    parser.add_argument("--eval", type=Path, help="held-out JSONL for the evaluation loss")
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--steps", type=int, default=30)
    parser.add_argument("--max-len", type=int, default=MAX_LEN)
    parser.add_argument("--sweep", type=int_list, default=SWEEP, help="comma-separated lengths")
    parser.add_argument("--rank", type=int, default=16)
    parser.add_argument("--learning-rate", type=float, default=1e-4)
    parser.add_argument("--seed", type=int, default=0)
    return parser.parse_args(argv)


def int_list(text: str) -> list[int]:
    return [int(x) for x in text.split(",")]


def main() -> None:
    logging.basicConfig(format="%(asctime)s %(levelname)s %(message)s", level=logging.INFO)
    run(parse_args())


if __name__ == "__main__":
    main()
