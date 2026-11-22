"""Target spans on real renders of a real attempt, for both students."""

from finetune.data import render, rows


def build(student, template_tokenizer, attempt):
    return rows.example(template_tokenizer, student, attempt)


def targets_text(row):
    return [row.text[a:b] for a, b in row.targets]


def test_every_assistant_output_is_a_target(student, template_tokenizer, attempt):
    row = build(student, template_tokenizer, attempt)
    joined = "".join(targets_text(row))

    for message in attempt.messages:
        if message["role"] != "assistant":
            continue
        if message.get("reasoning_content"):
            assert message["reasoning_content"].strip() in joined
        if message["content"].strip():
            assert message["content"].strip() in joined
        for call in message.get("tool_calls", []):
            command = call["function"]["arguments"]["command"]
            assert command in joined


def test_context_is_never_a_target(student, template_tokenizer, attempt):
    row = build(student, template_tokenizer, attempt)
    joined = "".join(targets_text(row))

    system, task = attempt.messages[0]["content"], attempt.messages[1]["content"]
    assert system[:200] not in joined
    assert task not in joined
    for message in attempt.messages:
        if message["role"] == "tool" and len(message["content"]) > 40:
            assert message["content"].strip() not in joined


def test_spans_are_ordered_and_disjoint(student, template_tokenizer, attempt):
    row = build(student, template_tokenizer, attempt)
    spans = row.targets
    assert spans
    assert all(a < b for a, b in spans)
    assert all(prev[1] <= cur[0] for prev, cur in zip(spans, spans[1:], strict=False))


def test_turn_ends_are_targets_so_the_model_learns_to_stop(student, template_tokenizer, attempt):
    row = build(student, template_tokenizer, attempt)
    texts = targets_text(row)
    if student.family == "qwen":
        assert all(t.endswith("<|im_end|>") for t in texts)
    else:
        assert texts[-1].endswith("<turn|>")
        # After a tool call the model stops by emitting <|tool_response>.
        assert sum(t.endswith("<|tool_response>") for t in texts) == sum(
            len(m.get("tool_calls", [])) for m in attempt.messages
        )


def test_generation_prompt_is_not_a_target(student, template_tokenizer, attempt):
    """Every served prompt ends right where a target starts (Qwen exactly)."""
    row = build(student, template_tokenizer, attempt)
    starts = {a for a, _ in row.targets}
    first_turn = next(i for i, m in enumerate(attempt.messages) if m["role"] == "assistant")
    prompt = render.render(template_tokenizer, attempt, upto=first_turn)
    if student.family == "qwen":
        assert len(prompt) in starts
    else:
        # Gemma's first served prompt ends with "<|turn>model\n"; the
        # thought channel the model writes next is a target.
        assert prompt.endswith("<|turn>model\n")
        assert len(prompt) in starts


def test_mask_context_drops_the_spans_inside_the_context():
    text = "promptturnmodel answer"

    assert rows.mask_context(text, "prompt", [(0, 6), (6, len(text))]) == [(6, len(text))]


def test_mask_context_needs_the_context_at_the_start():
    try:
        rows.mask_context("abc", "bc", [(0, 1)])
    except ValueError as error:
        assert "not a prefix" in str(error)
    else:
        raise AssertionError("a context that does not start the text was accepted")


def test_mask_context_tolerates_only_the_turn_terminator_of_the_prefix():
    # A template closes the turn that ends the context where the full render continues it.
    text = "task<turn|>model<|channel>thought\nfix it"
    context = "task<turn|>model<turn|>"

    assert rows.mask_context(text, context, [(0, 4), (11, 16)]) == []
    assert rows.mask_context(text, context, [(17, len(text))]) == [(17, len(text))]

    try:
        rows.mask_context("task" + "x" * 40 + "rest", "task" + "y" * 40 + "rest", [(0, 1)])
    except ValueError as error:
        assert "not a prefix" in str(error)
    else:
        raise AssertionError("a context that differs well before its end was accepted")
