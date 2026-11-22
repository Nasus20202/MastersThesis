"""Training examples: one full chat-template render per trajectory, with the
character spans the loss is computed on.

Targets are the text the model generates when serving: everything inside an
assistant turn except the generation prompt and tool results. Prompt, tool
definitions, the task and tool results are context only.

- Qwen3.5: each ``<|im_start|>assistant\\n<think>\\n ... <|im_end|>`` turn,
  starting after ``<think>\\n`` (part of the serving generation prompt).
- Gemma 4: each ``<|turn>model\\n ... <turn|>`` turn, minus the tool-result
  blocks ``<|tool_response>...<tool_response|>``. The opening
  ``<|tool_response>`` stays a target: it is the model's stop after a tool call.

Gemma's serving prompt closes the previous model turn when the assistant wrote
text next to a tool call, which a single render does not; the technical check
showed that training on full renders still transfers to the served model
.
"""

from __future__ import annotations

import dataclasses
import re
from pathlib import Path

from finetune.config import Family, Student
from finetune.data import render
from finetune.data.models import Row, Trajectory

QWEN_TURN = re.compile(r"<\|im_start\|>assistant\n(?:<think>\n)?(.*?<\|im_end\|>)", re.DOTALL)


GEMMA_TURN = re.compile(r"<\|turn>model\n(.*?(?:<turn\|>|$))", re.DOTALL)


GEMMA_TOOL_RESULT = re.compile(r"<\|tool_response>(.*?<tool_response\|>)", re.DOTALL)


def target_spans(family: Family, text: str) -> list[tuple[int, int]]:
    if family == Family.QWEN:
        return [m.span(1) for m in QWEN_TURN.finditer(text)]
    spans = []
    for turn in GEMMA_TURN.finditer(text):
        start, end = turn.span(1)
        for result in GEMMA_TOOL_RESULT.finditer(text, start, end):
            spans.append((start, result.start(1)))
            start = result.end(1)
        spans.append((start, end))
    return [(a, b) for a, b in spans if b > a]


# A template closes the model turn that ends a prefix render where the full
# render continues it: Gemma ends the context with ``<turn|>`` and the full
# conversation with ``<|channel>thought``. The two agree up to that terminator.
TURN_TAIL = 16


def mask_context(text: str, context: str, spans: list[tuple[int, int]]) -> list[tuple[int, int]]:
    """The spans after ``context``, for a conversation whose first messages the
    loss must not see (the prefix of a continuation).

    The boundary is where the full render leaves the prefix render, not the
    latter's length: the template's turn terminator sits on the context side.
    """
    boundary = next(
        (i for i, (a, b) in enumerate(zip(text, context, strict=False)) if a != b),
        min(len(text), len(context)),
    )
    if boundary == 0 or len(context) - boundary > TURN_TAIL:
        raise ValueError("the context is not a prefix of the rendered conversation")
    return [(a, b) for a, b in spans if a >= boundary]


def read_rows(path: Path) -> list[Row]:
    return [Row.from_json(line) for line in path.read_text().splitlines() if line.strip()]


def example(tok, student: Student, traj: Trajectory) -> Row:
    text = render.render(tok, traj)
    return Row(traj.source, traj.scenario_id, text, target_spans(student.family, text))


def replace_source(row: Row, source: str) -> Row:
    """The row with its source set to the raw record's path in the repository."""
    return dataclasses.replace(row, source=source)
