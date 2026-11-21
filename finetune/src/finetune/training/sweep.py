"""Peak memory and speed of a training step at several sequence lengths, and
the training cap that follows from them."""

from __future__ import annotations

import logging

import torch

from finetune.training.gpu import is_oom, oom_message, peak_gib, reset_peak
from finetune.training.models import Settings, SweepPoint, SweepStatus, Tokens
from finetune.training.trainer import Trainer

log = logging.getLogger(__name__)


def sweep(
    model, cfg: Settings, tokens: Tokens, lengths: list[int], repeats: int = 3
) -> list[SweepPoint]:
    """Peak memory and speed of a training step at each sequence length.

    Runs with learning rate 0 on a separate optimizer, so the adapter is not
    changed by the measurement. Speed is the median of the repeats after the
    first.
    """
    results = []
    for length in lengths:
        if len(tokens.ids) < length:
            results.append(SweepPoint(length, SweepStatus.NOT_ENOUGH_TOKENS))
            continue
        log.info("sweep: %d tokens", length)
        trainer = Trainer(model, cfg, repeats, learning_rate=0.0)
        reset_peak()
        try:
            logs = [trainer.step(tokens.window(length)) for _ in range(repeats)]
        except (RuntimeError, torch.OutOfMemoryError) as error:
            if not is_oom(error):
                raise
            log.warning("sweep: %d tokens out of memory", length)
            results.append(SweepPoint(length, SweepStatus.OUT_OF_MEMORY, error=oom_message(error)))
            continue
        finally:
            del trainer
        timed = [entry.seconds for entry in logs[1:]] or [logs[0].seconds]
        seconds = sorted(timed)[len(timed) // 2]
        log.info(
            "sweep: %d tokens fit, peak %.1f GiB, %.1f s/step, %.0f tokens/s",
            length,
            peak_gib(),
            seconds,
            length / seconds,
        )
        results.append(
            SweepPoint(
                length,
                SweepStatus.OK,
                peak_gib=peak_gib(),
                seconds_per_step=seconds,
                tokens_per_second=length / seconds,
                loss=logs[0].loss,
            )
        )
    return results


def training_cap(results: list[SweepPoint], max_len: int) -> int:
    """``max_len``, or the longest swept length that fitted below the shortest
    one that ran out of memory (0 if none fitted)."""
    failed = [r.length for r in results if r.status == SweepStatus.OUT_OF_MEMORY]
    if not failed:
        return max_len
    fitting = [r.length for r in results if r.status == SweepStatus.OK and r.length < min(failed)]
    return min(max_len, max(fitting)) if fitting else 0
