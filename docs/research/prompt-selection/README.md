# Prompt selection

## Method

This pilot compares a small set of troubleshooting prompts before one is proposed for the benchmark
prompt condition. The approaches respond to failures observed in the
[baseline incident evaluation](../baseline-behaviour/README.md): repairs to
transient resources, changes beyond the needed scope, and claims of success
without final-state verification. This pilot can show how these complete
prompts perform on the development scenarios; it cannot isolate the effect of
individual words or establish a globally best prompt.

| ID                | Approach                                                                      | Prompt                                              |
| ----------------- | ----------------------------------------------------------------------------- | --------------------------------------------------- |
| `workflow`        | Current seven-step troubleshooting prompt, copied before the default changes  | [workflow.md](candidates/workflow.md)               |
| `hypothesis`      | Test a likely cause with discriminating observations before repair            | [hypothesis.md](candidates/hypothesis.md)           |
| `constraints`     | Establish task constraints and ownership before the smallest justified change | [constraints.md](candidates/constraints.md)         |
| `outcome`         | Define observable success first, then repair and verify against it            | [outcome.md](candidates/outcome.md)                 |
| `tool-guidance`   | Explain how to use the Bash tool and `kubectl` observations                   | [tool-guidance.md](candidates/tool-guidance.md)     |
| `decision-points` | Use evidence gates to decide when to inspect, change or finish                | [decision-points.md](candidates/decision-points.md) |

The prompts are generic: none contains a scenario answer or hidden scoring
criterion. Each asks the agent to inspect, repair and verify; the difference
is what it emphasizes. The approaches came from the [baseline incident
evaluation](../baseline-behaviour/README.md). The tool-guidance prompt also
draws on the benchmark's [Bash tool contract](../../../benchmark/internal/agent/common/bash.go)
and the official [kubectl reference](https://kubernetes.io/docs/reference/kubectl/)
(accessed 2026-09-18).

Each prompt ran three times on the same five development scenarios: six
prompts × five scenarios × three attempts = **90 attempts**. The only planned
change between candidates was the system prompt. The pilot held the Gemma 4
E4B QAT Q4_0 model, llama.cpp Vulkan `b10524` runtime, sandbox and Bash tool,
scenarios, scoring, task format and agent limits fixed. Each attempt allowed
25 turns, 50 tool calls, 60 seconds per tool and 300 seconds total. Benchmark
parallelism was 4. The [model profile](../../../benchmark/model-profiles/gemma-4-e4b.env),
[runtime settings](../../../benchmark/config.env) and
[benchmark settings](../../../benchmark/config.yaml) identify the setup.

Each scenario has two equally weighted final-state checks. An attempt scores
0, 0.5 or 1; full success requires both checks to pass. The comparison uses
per-scenario scores and full-success counts, then reads the transcripts for
diagnosis, repair scope, verification, tool use, duration and termination.
Three attempts per scenario can reveal failure modes, but cannot give a
reliable estimate of general improvement.

A follow-up tested two fixed combinations of the strongest ideas:
[constraints + decision points](candidates/constraints-decision-points.md) and
[the same prompt with Bash guidance](candidates/constraints-decision-points-tool-guidance.md).
The five scenarios, model, runtime, scoring, tool and agent limits remained
unchanged. Each combination was scheduled for three attempts per scenario
(30 attempts in the comparison). These are new complete prompts, not isolated
tests of a single sentence. Each combination's evidence is merged into one
15-attempt directory. The attempt records retain their source run IDs, with
the later Bash response-format change was not part of these runs.

## Results

Each prompt ran 15 times: three attempts on each of five incidents. The
`constraints` prompt passed the most final-state checks, 11/15, compared with
7/15 for `workflow`. That lead needs a closer look: three of its successes
ended at the agent deadline, and one RBAC repair granted more access than the
original configuration. The [conclusions](#conclusions) explain these cases.

| Prompt            | Mean score | Both checks passed | Timed out | No tool calls |
| ----------------- | ---------: | -----------------: | --------: | ------------: |
| `workflow`        |       0.50 |               7/15 |         2 |             1 |
| `hypothesis`      |       0.50 |               7/15 |         3 |             3 |
| `constraints`     |       0.73 |              11/15 |         3 |             3 |
| `outcome`         |       0.47 |               7/15 |         3 |             4 |
| `tool-guidance`   |       0.63 |               9/15 |         2 |             0 |
| `decision-points` |       0.63 |               9/15 |         2 |             0 |

“Both checks passed” means both declared grading criteria passed after the agent
stopped. The mean score includes partial successes. Grading also ran after a
timeout, so a passed check does not mean the agent finished or verified its
own answer.

### By incident

The next table places `workflow` beside each alternative for the same
incident. Score is the mean of the two equally weighted checks over three
attempts. Tool calls are averaged over those attempts; duration is the median
agent-loop time in seconds. A timeout means the agent reached 300 seconds.

| Incident                    | Prompt            | Mean score | Both checks passed | Timed out | Mean tool calls | Median seconds |
| --------------------------- | ----------------- | ---------: | -----------------: | --------: | --------------: | -------------: |
| `container-crash-loop`      | `workflow`        |       0.00 |                0/3 |         0 |             5.0 |            197 |
| `container-crash-loop`      | `hypothesis`      |       0.00 |                0/3 |         1 |            12.0 |            244 |
| `container-crash-loop`      | `constraints`     |       0.00 |                0/3 |         0 |             2.3 |             26 |
| `container-crash-loop`      | `outcome`         |       0.00 |                0/3 |         1 |             5.0 |            175 |
| `container-crash-loop`      | `tool-guidance`   |       0.33 |                1/3 |         0 |             6.3 |            201 |
| `container-crash-loop`      | `decision-points` |       0.67 |                2/3 |         0 |             8.3 |            186 |
| `image-pull-failure`        | `workflow`        |       0.67 |                2/3 |         1 |             7.3 |            169 |
| `image-pull-failure`        | `hypothesis`      |       1.00 |                3/3 |         0 |             5.7 |            120 |
| `image-pull-failure`        | `constraints`     |       0.67 |                2/3 |         0 |             5.0 |            124 |
| `image-pull-failure`        | `outcome`         |       0.67 |                2/3 |         0 |             4.3 |            114 |
| `image-pull-failure`        | `tool-guidance`   |       1.00 |                3/3 |         0 |             7.3 |            146 |
| `image-pull-failure`        | `decision-points` |       0.67 |                2/3 |         0 |             4.7 |            139 |
| `missing-rbac-binding`      | `workflow`        |       0.17 |                0/3 |         1 |             9.3 |            274 |
| `missing-rbac-binding`      | `hypothesis`      |       0.33 |                1/3 |         2 |             7.0 |            300 |
| `missing-rbac-binding`      | `constraints`     |       1.00 |                3/3 |         2 |             8.3 |            300 |
| `missing-rbac-binding`      | `outcome`         |       0.33 |                1/3 |         2 |             8.0 |            300 |
| `missing-rbac-binding`      | `tool-guidance`   |       0.17 |                0/3 |         2 |            11.3 |            300 |
| `missing-rbac-binding`      | `decision-points` |       0.50 |                1/3 |         2 |            11.3 |            300 |
| `service-selector-mismatch` | `workflow`        |       1.00 |                3/3 |         0 |             6.0 |            119 |
| `service-selector-mismatch` | `hypothesis`      |       0.67 |                2/3 |         0 |             4.3 |            110 |
| `service-selector-mismatch` | `constraints`     |       1.00 |                3/3 |         1 |             7.3 |            103 |
| `service-selector-mismatch` | `outcome`         |       1.00 |                3/3 |         0 |             6.3 |            103 |
| `service-selector-mismatch` | `tool-guidance`   |       1.00 |                3/3 |         0 |             7.0 |             91 |
| `service-selector-mismatch` | `decision-points` |       0.67 |                2/3 |         0 |             4.3 |             68 |
| `unschedulable-cpu-request` | `workflow`        |       0.67 |                2/3 |         0 |             6.0 |            193 |
| `unschedulable-cpu-request` | `hypothesis`      |       0.50 |                1/3 |         0 |             7.3 |             76 |
| `unschedulable-cpu-request` | `constraints`     |       1.00 |                3/3 |         0 |             6.7 |            149 |
| `unschedulable-cpu-request` | `outcome`         |       0.33 |                1/3 |         0 |             3.0 |             42 |
| `unschedulable-cpu-request` | `tool-guidance`   |       0.67 |                2/3 |         0 |             6.3 |            167 |
| `unschedulable-cpu-request` | `decision-points` |       0.67 |                2/3 |         0 |             6.3 |            185 |

### Raw data

Each run contains its metadata and all 15 attempt records, including failures
and transcripts. The 15 agent timeouts are counted in the tables above; 11
occurred in the RBAC scenario.

| Prompt            | Run and attempt records                                                                   |
| ----------------- | ----------------------------------------------------------------------------------------- |
| `workflow`        | [run-2026-09-18-08-17-02-510Z](raw/workflow/run-2026-09-18-08-17-02-510Z/run.json)        |
| `hypothesis`      | [run-2026-09-18-08-37-02-490Z](raw/hypothesis/run-2026-09-18-08-37-02-490Z/run.json)      |
| `constraints`     | [run-2026-09-18-08-55-58-097Z](raw/constraints/run-2026-09-18-08-55-58-097Z/run.json)     |
| `outcome`         | [run-2026-09-18-09-13-08-391Z](raw/outcome/run-2026-09-18-09-13-08-391Z/run.json)         |
| `tool-guidance`   | [run-2026-09-18-09-29-44-826Z](raw/tool-guidance/run-2026-09-18-09-29-44-826Z/run.json)   |
| `decision-points` | [run-2026-09-18-09-50-17-818Z](raw/decision-points/run-2026-09-18-09-50-17-818Z/run.json) |

### Combination follow-up

Both combinations passed both final-state checks in **11/15 agent attempts**
(mean score 0.73), compared with 7/15 for `workflow`. Neither combination
had a no-tool attempt. The version without Bash guidance timed out once; the
Bash-guided version timed out twice, including one attempt that passed the
checks only after the deadline. Each row below covers three agent attempts.
The `workflow` rows for the same incidents are in the table above.

| Incident                    | Combination             | Mean score | Both checks | Timeouts | Mean tool calls | Median seconds |
| --------------------------- | ----------------------- | ---------: | ----------: | -------: | --------------: | -------------: |
| `container-crash-loop`      | constraints + decisions |       0.67 |         2/3 |        0 |             6.7 |            197 |
|                             | + Bash guidance         |       0.33 |         1/3 |        0 |            10.0 |            102 |
| `image-pull-failure`        | constraints + decisions |       1.00 |         3/3 |        0 |             8.0 |            149 |
|                             | + Bash guidance         |       1.00 |         3/3 |        0 |             7.0 |             97 |
| `missing-rbac-binding`      | constraints + decisions |       0.33 |         1/3 |        1 |             7.3 |            171 |
|                             | + Bash guidance         |       0.67 |         2/3 |        2 |            11.0 |            300 |
| `service-selector-mismatch` | constraints + decisions |       1.00 |         3/3 |        0 |             6.3 |             81 |
|                             | + Bash guidance         |       1.00 |         3/3 |        0 |             7.3 |             63 |
| `unschedulable-cpu-request` | constraints + decisions |       0.67 |         2/3 |        0 |             8.7 |            152 |
|                             | + Bash guidance         |       0.67 |         2/3 |        0 |             7.3 |            103 |

Relative to `workflow`, the combinations gained on crash-loop, image-pull
and RBAC repairs; they matched its service-selector and CPU-request counts.
They traded successes between incidents: the version without Bash guidance
passed one more crash-loop attempt, while the Bash-guided version passed one
more RBAC attempt. Three attempts per incident remain too few to distinguish
the prompts reliably.

The 30 attempts are in the
[constraints + decision points run](raw/constraints-decision-points/run-2026-09-18-22-03-44-795Z/run.json)
and [Bash-guided run](raw/constraints-decision-points-tool-guidance/run-2026-09-18-22-18-27-123Z/run.json).
Each directory contains the attempt transcripts and merged run metadata.

## Conclusions

The four extra successes for `constraints` over `workflow` came from the RBAC
incident (3/3 versus 0/3) and CPU request incident (3/3 versus 2/3). It did
not solve the crash loop. Two of its three RBAC successes reached the time
limit before the agent could finish. In [RBAC attempt 002](raw/constraints/run-2026-09-18-08-55-58-097Z/missing-rbac-binding/prompt/002.json),
the agent first guessed the wrong service account, then created a new Role and
RoleBinding and deleted a Pod. The final Role allowed `get` on every ConfigMap;
the original Role allowed it only for `app-config`. The [grading check](../../../benchmark/scenarios/missing-rbac-binding/scripts/check-least-privilege.sh)
does not test access to other ConfigMaps. That attempt passed the recorded
checks, but its permission change exceeded the original scope.

The crash loop remained difficult. `decision-points` passed 2/3 attempts and
`tool-guidance` passed 1/3; the other four prompts passed none. Successful
[tool-guidance attempt 002](raw/tool-guidance/run-2026-09-18-09-29-44-826Z/container-crash-loop/prompt/002.json)
removed the invalid arguments from the Deployment and checked that all three
updated replicas were ready. In failed [decision-points attempt 002](raw/decision-points/run-2026-09-18-09-50-17-818Z/container-crash-loop/prompt/002.json),
the agent changed the arguments to `["nginx"]` and reported success after
seeing three ready Pods. Grading showed only one of three new replicas updated
and a rollout that never finished. Those ready Pods belonged to the old
ReplicaSet. The agent had checked Pods, but not the requested rollout.

Eleven attempts made no tool call. Some asked the user for names they could
have found with `kubectl`. All eleven came from `workflow`, `hypothesis`,
`constraints` or `outcome`; every `tool-guidance` and `decision-points` attempt
used the tool. Tool use alone did not settle the hard cases: `tool-guidance`
passed none of its three RBAC attempts.

RBAC also accounted for 11 of the 15 agent timeouts. Six timed-out attempts
still passed both final-state checks because grading ran after the deadline.
This explains why a passed check and a completed agent response appear as
separate outcomes above. With three attempts per incident, these differences
are useful for choosing what to test next, but they do not establish a
generally better prompt. The six standalone candidates did not yield a
definitive winner. The follow-up selected the constraints + decision points
combination for `prompt.md`; this remains a small-pilot result rather than
proof of general superiority.

The combinations tie on the recorded checks, but their failure modes matter.
The first combination repaired the existing RoleBinding without broadening
its Role in
[RBAC attempt 001](raw/constraints-decision-points/run-2026-09-18-22-03-44-795Z/missing-rbac-binding/prompt/001.json),
but also deleted a Pod unnecessarily. Another RBAC attempt stopped after a
failed `grep` search and asked the user for resource names that the cluster
could have supplied. The Bash-guided prompt repaired the RoleBinding in one
completed RBAC attempt, but [another attempt](raw/constraints-decision-points-tool-guidance/run-2026-09-18-22-18-27-123Z/missing-rbac-binding/prompt/002.json)
timed out after attempting an over-broad Role edit. Two crash-loop attempts
([one](raw/constraints-decision-points-tool-guidance/run-2026-09-18-22-18-27-123Z/container-crash-loop/prompt/002.json),
[two](raw/constraints-decision-points-tool-guidance/run-2026-09-18-22-18-27-123Z/container-crash-loop/prompt/003.json))
claimed the app was healthy
from three ready old Pods while the new rollout failed both grading checks.
This repeats the final-state verification problem seen in the original pilot.

The equal denominators do not remove the small-sample uncertainty. The
version without Bash guidance had
fewer timeouts and one more completed crash-loop repair; the researcher chose
it for `prompt.md`. This is a selection from a small pilot, not proof that it
is generally better. Neither combination consistently verified the new
rollout before claiming success.
