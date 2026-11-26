"""LoRA training of one model with Hugging Face transformers + PEFT.

One run, in order (``run``):

1. **Load** the pinned checkpoint with a fresh LoRA adapter onto the GPU
   (``finetune.training.lora``).
2. **Sweep**: peak GPU memory and speed of a training step at each sequence
   length, with learning rate 0, so the adapter is unchanged.
3. **Cap**: training uses ``--max-len``, or the longest swept length that
   fitted if a longer one ran out of memory. Longer trajectories are dropped,
   never cut.
4. **Evaluate** the target-token loss on the held-out trajectories, with and
   without the adapter.
5. **Train** ``--epochs`` shuffled passes over the kept trajectories, one
   trajectory per step (or a fixed ``--steps``, for short checks).
   After each epoch but the last, the held-out loss is measured again, to
   show whether later epochs over-train (a diagnostic, not a selection).
   The adapter of each such epoch is saved too (``adapter-epoch-N/``), so a
   run that is cut off keeps its finished epochs.
6. **Evaluate** again and save the adapter.

The loss is the mean cross-entropy over target tokens (``finetune.data.rows``),
computed from the last hidden state in chunks of target positions, each chunk
recomputed in the backward pass, so the full logits matrix (sequence x
vocabulary, 262K for Gemma) is never held in memory.

Outputs in ``--out``: ``metrics.json``, rewritten after every phase so that a
failure keeps what was measured, ``adapter-epoch-N/`` and ``adapter/``.
"""

from __future__ import annotations

import argparse
import logging
import random
import resource
import time
import traceback
from collections.abc import Callable
from pathlib import Path

import torch

from finetune import logs
from finetune.config import STUDENTS
from finetune.training.environment import environment
from finetune.training.evaluate import evaluate
from finetune.training.gpu import allocated_gib, is_oom, oom_message, peak_gib, reset_peak
from finetune.training.lora import adapted_modules, load, trainable_parameters
from finetune.training.models import (
    TRAINING,
    EpochEvaluation,
    ExampleCount,
    Job,
    Metrics,
    Precision,
    Settings,
    Step,
    Tokens,
)
from finetune.training.sweep import sweep, training_cap
from finetune.training.tokens import read_examples
from finetune.training.trainer import Trainer, train_steps

log = logging.getLogger(__name__)


def run(job: Job, saved: Callable[[], None] = lambda: None) -> Metrics:
    """One training run (module docstring); ``saved()`` after each epoch's
    adapter is written, to persist it (the Modal volume commit)."""
    random.seed(job.settings.seed)
    torch.manual_seed(job.settings.seed)
    job.out.mkdir(parents=True, exist_ok=True)
    metrics = Metrics(job, environment())
    with logs.LogFile(job.out / "log.txt"):
        log.info("job %s", job)
        log.info("environment %s", metrics.environment)
        try:
            _run(job, metrics, saved)
        except BaseException as error:
            metrics.error = "".join(traceback.format_exception_only(error)).strip()[-2000:]
            log.exception("run failed")
            raise
        finally:
            metrics.host_peak_gib = host_peak_gib()
            metrics.save()
    return metrics


def host_peak_gib() -> float:
    """Peak resident memory of this process (Linux reports KiB)."""
    return resource.getrusage(resource.RUSAGE_SELF).ru_maxrss / 2**20


def _run(job: Job, metrics: Metrics, saved: Callable[[], None]) -> None:
    cfg = job.settings
    if job.student:
        student = STUDENTS[job.student]
        checkpoint_path, revision = student.checkpoint, student.revision
    else:
        checkpoint_path, revision = str(job.checkpoint), None

    # 1. Load.
    started = time.perf_counter()
    reset_peak()
    model, tok = load(checkpoint_path, revision, cfg)
    dtypes: dict[str, int] = {}
    for p in model.parameters():
        dtypes[str(p.dtype)] = dtypes.get(str(p.dtype), 0) + p.numel()
    metrics.attention = getattr(model.config, "_attn_implementation", None)
    metrics.load_seconds = time.perf_counter() - started
    metrics.load_peak_gib = peak_gib()
    metrics.weights_gib = allocated_gib()
    metrics.parameter_dtypes = dtypes
    metrics.trainable_parameters = trainable_parameters(model)
    metrics.adapted_modules = adapted_modules(model)
    log.info(
        "ready in %.0f s: %d trainable parameters in %d modules, load peak %.1f GiB",
        metrics.load_seconds,
        metrics.trainable_parameters,
        len(metrics.adapted_modules),
        metrics.load_peak_gib,
    )
    metrics.save()

    train = read_examples(tok, job.data)
    evals = read_examples(tok, job.eval) if job.eval else []

    # 2. Sweep, on all training tokens concatenated so every length has real text.
    joined = Tokens([t for e in train for t in e.ids], [t for e in train for t in e.labels])
    metrics.sweep = sweep(model, cfg, joined, sorted(job.sweep))

    # 3. Cap: longer trajectories are dropped, never cut (rows.example).
    metrics.max_len = training_cap(metrics.sweep, job.max_len)
    kept = [e for e in train if len(e.ids) <= metrics.max_len]
    kept_evals = [e for e in evals if len(e.ids) <= metrics.max_len]
    metrics.train_examples = ExampleCount(len(train), len(kept), sorted(len(e.ids) for e in train))
    metrics.eval_examples = ExampleCount(
        len(evals), len(kept_evals), sorted(len(e.ids) for e in evals)
    )
    log.info(
        "cap %d tokens: %d of %d training and %d of %d evaluation trajectories kept",
        metrics.max_len,
        len(kept),
        len(train),
        len(kept_evals),
        len(evals),
    )
    metrics.save()

    # 4. Evaluate before training.
    if kept_evals:
        reset_peak()
        metrics.eval_before = evaluate(model, kept_evals)
        metrics.save()

    # 5. Train.
    metrics.steps = job.steps if job.steps is not None else job.epochs * len(kept)
    trainer = Trainer(model, cfg, metrics.steps)
    log.info("training %d steps, warm-up %d", metrics.steps, trainer.warmup_steps)

    def on_step(entry: Step) -> None:
        metrics.train_log.append(entry)
        log.info(
            "step %d/%d: loss %.4f, %d tokens, %.1f s, peak %.1f GiB",
            entry.step,
            metrics.steps,
            entry.loss,
            entry.tokens,
            entry.seconds,
            peak_gib(),
        )
        metrics.save()

    def on_epoch(epoch: int) -> None:
        model.save_pretrained(job.out / f"adapter-epoch-{epoch}")
        log.info("epoch %d: adapter saved", epoch)
        saved()
        if not kept_evals:
            return
        evaluation = evaluate(model, kept_evals)
        metrics.eval_epochs.append(EpochEvaluation(epoch, trainer.steps, evaluation))
        log.info("epoch %d: held-out loss %.4f", epoch, evaluation.with_adapter)
        metrics.save()

    reset_peak()
    try:
        train_steps(trainer, kept, metrics.steps, cfg.seed, on_step, on_epoch)
    except (RuntimeError, torch.OutOfMemoryError) as error:
        if not is_oom(error):
            raise
        # Recorded, not raised: the steps so far and the evaluation still count.
        metrics.train_error = oom_message(error)
        log.warning("training stopped after %d steps: out of memory", len(metrics.train_log))
    tokens = sum(e.tokens for e in metrics.train_log)
    seconds = sum(e.seconds for e in metrics.train_log)
    metrics.train_peak_gib = peak_gib()
    metrics.train_tokens_per_second = tokens / seconds if seconds else None
    metrics.train_seconds = seconds
    metrics.save()

    # 6. Evaluate after training, save the adapter.
    if kept_evals:
        reset_peak()
        metrics.eval_after = evaluate(model, kept_evals)
    model.save_pretrained(job.out / "adapter")
    log.info("adapter saved to %s", job.out / "adapter")


def parse_args(argv: list[str] | None = None) -> Job:
    parser = argparse.ArgumentParser(
        description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter
    )
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--student", choices=sorted(STUDENTS))
    source.add_argument("--checkpoint", type=Path, help="local model directory (tests)")
    parser.add_argument("--precision", type=Precision, choices=list(Precision), required=True)
    parser.add_argument(
        "--data", type=Path, required=True, help="training JSONL (finetune.data.rows)"
    )
    parser.add_argument("--eval", type=Path, help="held-out JSONL for the evaluation loss")
    parser.add_argument("--out", type=Path, required=True)
    parser.add_argument("--epochs", type=int, default=TRAINING.epochs)
    parser.add_argument("--steps", type=int, help="a fixed step count instead of --epochs")
    parser.add_argument("--max-len", type=int, default=TRAINING.max_len)
    parser.add_argument(
        "--sweep", type=int_list, default=TRAINING.sweep, help="comma-separated lengths"
    )
    parser.add_argument("--rank", type=int, default=TRAINING.rank)
    parser.add_argument("--alpha", type=int, default=TRAINING.alpha)
    parser.add_argument("--learning-rate", type=float, default=TRAINING.learning_rate)
    parser.add_argument("--seed", type=int, default=TRAINING.seed)
    args = parser.parse_args(argv)
    return Job(
        settings=Settings(
            precision=args.precision,
            rank=args.rank,
            alpha=args.alpha,
            learning_rate=args.learning_rate,
            seed=args.seed,
        ),
        data=args.data,
        out=args.out,
        student=args.student,
        checkpoint=args.checkpoint,
        eval=args.eval,
        epochs=args.epochs,
        steps=args.steps,
        max_len=args.max_len,
        sweep=tuple(args.sweep),
    )


def int_list(text: str) -> list[int]:
    return [int(x) for x in text.split(",")]


def main() -> None:
    logs.setup()
    run(parse_args())


if __name__ == "__main__":
    main()
