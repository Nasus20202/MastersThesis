from finetune import render
from finetune.students import Student
from finetune.trajectories import Trajectory

STUDENT = Student("gemma", "x", "rev", "served", in_study=True)

TRAJ = Trajectory(
    source="a",
    scenario_id="s",
    condition="prompt",
    model="m",
    peak_context_tokens=1,
    tools=[],
    messages=[
        {"role": "user", "content": "u"},
        {"role": "assistant", "content": "a1"},
        {"role": "tool", "content": "t"},
        {"role": "assistant", "content": "a2"},
    ],
)


class FakeTokenizer:
    bos_token = "<bos>"

    def apply_chat_template(self, messages, add_generation_prompt, **_):
        text = "<bos>" + "".join(f"[{m['content']}]" for m in messages)
        return text + ("<gen>" if add_generation_prompt else "")


def test_compare_ignores_bos_and_flags_non_prefix_prompts(monkeypatch):
    monkeypatch.setattr(render, "tokenizer", lambda student: FakeTokenizer())
    served = {1: "[u]<gen>", 3: "[u][a1][t]<gen>"}
    monkeypatch.setattr(render, "served_prompt", lambda url, model, traj, upto: served[upto])

    checks = render.compare("http://x", STUDENT, TRAJ)

    assert [c.turn for c in checks] == [1, 3]
    assert all(c.served_equals_hf_prompt for c in checks)
    # The generation prompt is not part of the full render, so neither turn
    # prompt is a prefix of it.
    assert not any(c.served_is_prefix_of_full for c in checks)


def test_compare_reports_first_difference(monkeypatch):
    monkeypatch.setattr(render, "tokenizer", lambda student: FakeTokenizer())
    monkeypatch.setattr(render, "served_prompt", lambda url, model, traj, upto: "[u]<other>")

    checks = render.compare("http://x", STUDENT, TRAJ)

    assert not checks[0].served_equals_hf_prompt
    assert '"at": 4' in checks[0].first_difference


def test_template_render_equals_llama_cpp_served_prompt(
    student_name, student, template_tokenizer, attempt
):
    """Golden prompts recorded from llama.cpp b10964 (/apply-template) for the
    fixture attempt: the training template must reproduce them exactly."""
    import json

    from tests.conftest import FIXTURES

    golden = json.loads((FIXTURES / f"served-{student_name}.json").read_text())
    for turn, served in golden["prompts"].items():
        hf = render.render(template_tokenizer, attempt, upto=int(turn))
        assert hf.removeprefix(render.bos_text(student)) == served


def test_qwen_serving_is_append_only(attempt):
    """Every Qwen served prompt is a prefix of the full render: one training
    sequence per trajectory sees exactly the serving context."""
    import json

    from tests.conftest import FIXTURES, TemplateTokenizer

    full = render.render(TemplateTokenizer("qwen"), attempt)
    golden = json.loads((FIXTURES / "served-qwen35-4b.json").read_text())
    assert all(full.startswith(p) for p in golden["prompts"].values())
