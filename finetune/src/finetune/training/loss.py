"""Mean cross-entropy over target tokens.

Computed from the last hidden state in chunks of target positions, each chunk
recomputed in the backward pass, so the full logits matrix (sequence x
vocabulary, 262K for Gemma) is never held in memory.
"""

from __future__ import annotations

import torch
import torch.nn.functional as F
from torch.utils.checkpoint import checkpoint

from finetune.training.tokens import IGNORE

LOSS_CHUNK = 256


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
