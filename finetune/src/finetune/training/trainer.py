"""Training steps: AdamW with warm-up and the configured decay, one example per step."""

from __future__ import annotations

import math
import random
import time
from collections.abc import Callable
from dataclasses import replace

import torch

from finetune.training.gpu import autocast, sync
from finetune.training.lora import input_device
from finetune.training.loss import loss_and_count
from finetune.training.models import Settings, Step, Tokens


def lr_scale(step: int, warmup_steps: int, total_steps: int, schedule: str) -> float:
    """Factor on the learning rate at 0-based ``step``: linear warm-up, then
    ``constant`` or ``cosine`` decay to 0 at the last of ``total_steps``."""
    warmup = max(warmup_steps, 1)
    if step < warmup:
        return (step + 1) / warmup
    if schedule == "constant":
        return 1.0
    if schedule != "cosine":
        raise ValueError(f"unknown lr_schedule {schedule!r}")
    progress = (step - warmup + 1) / max(total_steps - warmup, 1)
    return 0.5 * (1.0 + math.cos(math.pi * min(progress, 1.0)))


class Trainer:
    """AdamW with the learning-rate schedule of ``cfg``; one example per step."""

    def __init__(self, model, cfg: Settings, total_steps: int, learning_rate: float | None = None):
        self.model = model
        self.cfg = cfg
        self.total_steps = total_steps
        self.device = input_device(model)
        params = [p for p in model.parameters() if p.requires_grad]
        self.base_lr = cfg.learning_rate if learning_rate is None else learning_rate
        self.optimizer = torch.optim.AdamW(params, lr=self.base_lr, weight_decay=0.0)
        self.steps = 0

    def step(self, example: Tokens) -> Step:
        self.model.train()
        scale = lr_scale(self.steps, self.cfg.warmup_steps, self.total_steps, self.cfg.lr_schedule)
        for group in self.optimizer.param_groups:
            group["lr"] = self.base_lr * scale
        ids_t = torch.tensor(example.ids, device=self.device)
        labels_t = torch.tensor(example.labels, device=self.device)
        sync()
        started = time.perf_counter()
        try:
            with autocast():
                loss, count = loss_and_count(self.model, ids_t, labels_t)
            loss.backward()
            self.optimizer.step()
        finally:
            # Also after running out of memory, so the next attempt starts clean.
            self.optimizer.zero_grad(set_to_none=True)
        sync()
        self.steps += 1
        return Step(
            loss=float(loss.detach()),
            tokens=len(example.ids),
            target_tokens=count,
            seconds=time.perf_counter() - started,
            finite=bool(torch.isfinite(loss.detach())),
        )


def train_steps(
    trainer: Trainer, examples: list[Tokens], steps: int, seed: int, on_step: Callable
) -> None:
    """``steps`` steps over shuffled epochs of ``examples``."""
    order = random.Random(seed)
    done = 0
    while examples and done < steps:
        epoch = list(examples)
        order.shuffle(epoch)
        for example in epoch[: steps - done]:
            done += 1
            on_step(replace(trainer.step(example), step=done))
