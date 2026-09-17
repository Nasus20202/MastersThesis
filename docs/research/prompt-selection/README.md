# Prompt selection

This pilot compares a small set of troubleshooting prompts before one is proposed for the benchmark
prompt condition. The approaches respond to failures observed in the
[baseline incident evaluation](../baseline-behaviour/README.md): repairs to
transient resources, changes beyond the needed scope, and claims of success
without final-state verification. This pilot can show how these complete
prompts perform on the development scenarios; it cannot isolate the effect of
individual words or establish a globally best prompt.

## Candidates

| ID                | Approach                                                                      | Prompt                                              | Configuration                                           |
| ----------------- | ----------------------------------------------------------------------------- | --------------------------------------------------- | ------------------------------------------------------- |
| `workflow`        | Current seven-step troubleshooting prompt, copied before the default changes  | [workflow.md](candidates/workflow.md)               | [workflow.yaml](candidates/workflow.yaml)               |
| `hypothesis`      | Test a likely cause with discriminating observations before repair            | [hypothesis.md](candidates/hypothesis.md)           | [hypothesis.yaml](candidates/hypothesis.yaml)           |
| `constraints`     | Establish task constraints and ownership before the smallest justified change | [constraints.md](candidates/constraints.md)         | [constraints.yaml](candidates/constraints.yaml)         |
| `outcome`         | Define observable success first, then repair and verify against it            | [outcome.md](candidates/outcome.md)                 | [outcome.yaml](candidates/outcome.yaml)                 |
| `tool-guidance`   | Explain how to use the Bash tool and `kubectl` observations                   | [tool-guidance.md](candidates/tool-guidance.md)     | [tool-guidance.yaml](candidates/tool-guidance.yaml)     |
| `decision-points` | Use evidence gates to decide when to inspect, change or finish                | [decision-points.md](candidates/decision-points.md) | [decision-points.yaml](candidates/decision-points.yaml) |

The candidates are generic strategies. They contain no
scenario-specific facts, hidden criteria or benchmark answers. Each still asks
the agent to inspect, make a justified repair and verify the result; the
approach and emphasis differ.

## Protocol

Run each unchanged prompt three times on each of the same five development
scenarios: six prompts × five scenarios × three attempts = **90 attempts**.
Keep the model artifact and runtime, sandbox and Bash tool, scenarios and
scoring, agent limits, task format, and result collection fixed. Only the
system prompt changes. Record the exact model profile, configuration, scenario
revision and run identifier for each candidate. Do not edit a prompt between
its three repetitions.

Compare each alternative with `workflow` using per-scenario partial scores
and full-success counts. Inspect raw transcripts for the observed diagnosis,
repair scope, verification, tool use, duration and termination state. Report
negative and inconclusive outcomes as well as improvements. With only three
attempts per scenario, small differences are exploratory rather than reliable
estimates of improvement. Present the evidence and a proposed prompt for
human review; do not freeze a prompt from this pilot alone.

## Execution

Run each candidate with its own overlay, including the copied workflow prompt:

```text
make benchmark AGENT=prompt REPEAT=3 CONFIG=docs/research/prompt-selection/candidates/workflow.yaml
```

Repeat for the other five YAML files in the table. Each overlay resolves its
prompt path relative to the overlay file. Do not use the checked-in default
prompt as a candidate: its text may change. Preserve the raw results under
the existing results root and associate every run ID with the candidate ID
and exact prompt file. Summarize the comparison here after execution.

## Results

Not run yet.

## Findings

Not available yet.

## Prompt sources

The candidate approaches draw on the [baseline incident
evaluation](../baseline-behaviour/README.md). The tool-guidance prompt uses
the benchmark's [Bash tool contract](../../../benchmark/internal/agent/common/bash.go)
and the official [kubectl reference](https://kubernetes.io/docs/reference/kubectl/)
(accessed 2026-09-18). No source text or scenario answer is included in the
prompts.
