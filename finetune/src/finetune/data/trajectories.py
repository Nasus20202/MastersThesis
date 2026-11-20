"""Benchmark attempts as chat conversations.

An attempt file (benchmark result JSON) stores the agent's messages with tool
calls flattened to ``{id, type, name, arguments}``. Chat templates expect the
OpenAI shape ``{id, type, function: {name, arguments}}`` with arguments as an
object, so both are converted here. Reasoning stays in ``reasoning_content``.
"""

from __future__ import annotations

import json
from pathlib import Path

from finetune.data.models import Trajectory


def _tool_call(call: dict) -> dict:
    arguments = call["arguments"]
    if isinstance(arguments, str):
        arguments = json.loads(arguments)
    return {
        "id": call["id"],
        "type": "function",
        "function": {"name": call["name"], "arguments": arguments},
    }


def _message(message: dict) -> dict:
    out = {"role": message["role"], "content": message.get("content") or ""}
    if message.get("reasoning_content"):
        out["reasoning_content"] = message["reasoning_content"]
    if message.get("tool_calls"):
        out["tool_calls"] = [_tool_call(c) for c in message["tool_calls"]]
    if message.get("tool_call_id"):
        out["tool_call_id"] = message["tool_call_id"]
    return out


def load(path: Path) -> Trajectory:
    data = json.loads(path.read_text())
    agent = data["agent"]
    return Trajectory(
        source=str(path),
        scenario_id=data["scenario_id"],
        condition=data["condition"],
        model=agent["inference"]["model"],
        peak_context_tokens=agent["token_usage"]["peak_context_tokens"],
        tools=[{"type": "function", "function": t} for t in agent["tools"]],
        messages=[_message(m) for m in agent["messages"]],
    )


def is_clean(path: Path) -> bool:
    """Full score and ended by the agent itself (D-054)."""
    data = json.loads(path.read_text())
    agent = data.get("agent") or {}
    return agent.get("termination") == "completed" and bool(data["grading"]["full_success"])
