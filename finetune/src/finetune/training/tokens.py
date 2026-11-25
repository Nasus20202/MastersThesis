"""Rows as token ids with labels: IGNORE outside the target spans."""

from __future__ import annotations

from pathlib import Path

from finetune.data import rows
from finetune.data.models import Row
from finetune.training.models import Tokens

IGNORE = -100


def tokenize(tok, row: Row) -> Tokens:
    enc = tok(row.text, add_special_tokens=False, return_offsets_mapping=True)
    ids = enc["input_ids"]
    labels = [IGNORE] * len(ids)
    spans = sorted(row.targets)
    j = 0
    for i, (start, end) in enumerate(enc["offset_mapping"]):
        if end <= start:
            continue
        while j < len(spans) and spans[j][1] <= start:
            j += 1
        if j < len(spans) and spans[j][0] <= start < spans[j][1]:
            labels[i] = ids[i]
    return Tokens(ids, labels, row.source)


def read_examples(tok, path: Path) -> list[Tokens]:
    return [tokenize(tok, row) for row in rows.read_rows(path)]
