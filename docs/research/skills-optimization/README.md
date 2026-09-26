# Skills optimization

## Status

Before this study the frozen skill condition was `S1`, the routing skill bundle committed in the repository, and the frozen prompt was `P1` (D-023). This study tested whether pruning skills the agent never loaded and moving repair procedures into symptom-routed subsystem skills improves repair outcomes over `S1` and the prompt agent, and selected the configuration to freeze.

`S10` is selected and frozen as the development skill condition and is now the repository default. On the **18-scenario subset**, where `S1` was measured, `S10` reached **0.777 macro score and 39/54 full successes** against 0.571 and 25/54 for `S1` and 0.649 and 30/54 for the best prompt. On the **full 46-scenario benchmark**, where `S1` was not run, `S10` reached **0.710 and 90/138** against 0.634 and 78/138 for the prompt agent and 0.696 and 85/138 for the strongest earlier candidate (`S8`). The `S11` enforced-load variant was slightly worse and was not selected. The unused `bash`, `playbook`, `connectivity` and `pod-security` skills are removed.

## Question

Does pruning unused skills and moving repair procedures into symptom-routed subsystem skills improve repair outcomes over the previous frozen skill set and the prompt agent, and which observed configuration should be frozen?

The design follows three observations from the earlier rounds: always-loaded content is used reliably while content behind `load_skill` and `load_reference` is not; the agent loads the always-loaded `troubleshooting` skill in nearly every attempt; and the earlier area content was descriptive (API reference) rather than procedural (a repair sequence). The working hypothesis was that routing to procedural subsystem skills raises their use and lifts the procedure-heavy scenarios.

## Method

### Scenarios and measures

The full benchmark is expensive to run repeatedly, so prompt and skill tuning used a fixed **18-scenario stratified subset** covering every scenario area and a spread of difficulty bands, while the final comparisons used **all 46 scenarios**. The subset is listed in [subset.txt](subset.txt) and was checked with the model-free validation runner. Runs on the subset use 18 scenarios; full-set runs use 46.

All runs held the model, runtime, scenarios, scoring, sandbox and agent limits fixed and changed only the prompt or skill files through benchmark config overlays. Each condition ran three independent attempts per scenario. The primary measure is the **macro-average normalized score**: the mean score per scenario, then the unweighted mean across scenarios. Full success requires every deterministic grading criterion to pass. Tables also report turn-limit terminations, mean tool calls and mean agent duration. Attempts that end in an error are kept in the denominator with their recorded score.

### Candidates

Each candidate is a complete file or skill tree under [`candidates/`](candidates/). None adds scenario answers or hidden benchmark knowledge. The candidates are listed as one complete series rather than split by round; the `Evaluated` column records which scope the candidate was actually run on.

Prompt candidates:

| ID   | Change from the frozen prompt                                                                                                              | Evaluated     |
| ---- | ------------------------------------------------------------------------------------------------------------------------------------------ | ------------- |
| `P1` | Frozen prompt from D-023 (constraints + decision points).                                                                                  | subset        |
| `P2` | Adds concrete final-state verification: confirm the reported symptom directly and compare requested with updated/ready/available replicas. | subset + full |
| `P3` | Adds explicit discovery discipline and bounded waits shorter than the tool deadline.                                                       | subset        |

Skill candidates:

| ID    | Change from the previous candidate                                                                                                                                               | Skill files                                               | Evaluated     |
| ----- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------- | ------------- |
| `S1`  | Frozen routing prompt and skill bundle (previous frozen set).                                                                                                                    | base + `bash`                                             | subset        |
| `S2`  | Load references only for the diagnosed subsystem, batch loads, add functional verification and an `Outcome:` report.                                                             | base + `bash`                                             | subset + full |
| `S3`  | `S2` plus reference additions for NetworkPolicy peers, Gateway attachment and image pull-policy staleness.                                                                       | base + `bash`                                             | subset        |
| `S4`  | `S2` plus generic repair discipline and an explicit symptom-to-reference routing table.                                                                                          | base                                                      | full          |
| `S5`  | `S4` plus always-loaded symptom checks in the `troubleshooting` skill.                                                                                                           | base                                                      | full          |
| `S6`  | `S5` plus the `S3` reference additions.                                                                                                                                          | base                                                      | run           |
| `S7`  | `S6` plus an always-loaded `playbook` skill.                                                                                                                                     | base + `bash`, `playbook`                                 | run           |
| `S8`  | `S7` plus targeted `connectivity` and `pod-security` skills and the outcome proofs and common-wrong-fix rules in the always-loaded prompt. Strongest earlier candidate.          | base + `bash`, `playbook`, `connectivity`, `pod-security` | full          |
| `S9`  | Prune the never- or barely-loaded skills (`bash`, `playbook`, `connectivity`, `pod-security`), fold their rules into the prompt, add a routing block and three subsystem skills. | base + `scheduling`, `storage`, `workloads`               | subset        |
| `S10` | Move the symptom-to-skill routing into the always-loaded `troubleshooting` skill and add all nine subsystem skills. Selected and frozen.                                         | base + nine subsystem skills                              | subset + full |
| `S11` | `S10` plus a hard requirement to load the matching subsystem skill.                                                                                                              | base + nine subsystem skills                              | full          |

The **base** bundle is `troubleshooting`, `kubernetes` (with references) and `kubectl`. The subsystem skills (`scheduling`, `storage`, `workloads`, `authorization`, `networking`, `configuration`, `governance`, `security`, `resources`) are procedural "how" guidance keyed to a symptom, distinct from the descriptive `kubernetes` references. They describe class-wide repair procedures and deliberately contain no scenario-specific fix.

### Configuration

The model is Gemma 4 E4B QAT `UD-Q4_K_XL`, served by the pinned llama.cpp Vulkan image with multi-token prediction. Attempts allow 25 turns, 50 tool calls, 60 seconds per tool call and 600 seconds total, with model-visible Bash output capped at 8 KiB. Exact reference configurations are in `raw/*/configuration/`.

## Results

### Subset comparison

Three repeats per scenario (54 attempts per condition). The `S10` and `S11` rows are the full runs restricted to the 18 subset scenarios.

| Condition            | Macro score |     Full success | Turn limits | Mean tools | Mean duration |
| -------------------- | ----------: | ---------------: | ----------: | ---------: | ------------: |
| baseline             |       0.398 |     19/54 (0.35) |           0 |        5.6 |           72s |
| prompt `P1` (frozen) |       0.603 |     28/54 (0.52) |           1 |        9.3 |          100s |
| prompt `P2`          |       0.649 |     30/54 (0.56) |           0 |        9.2 |          111s |
| prompt `P3`          |       0.609 |     27/54 (0.50) |           0 |        9.3 |          102s |
| skill `S1` (frozen)  |       0.571 |     25/54 (0.46) |           4 |       14.1 |          158s |
| skill `S2`           |       0.725 |     34/54 (0.63) |           2 |       12.9 |          136s |
| skill `S3`           |       0.628 |     29/54 (0.54) |           3 |       14.1 |          152s |
| skill `S9`           |       0.680 |     35/54 (0.65) |           1 |       12.4 |          135s |
| skill `S10`          |   **0.777** | **39/54 (0.72)** |           3 |       13.1 |          144s |
| skill `S11`          |       0.631 |     30/54 (0.56) |           1 |       13.1 |          143s |

The previous frozen set `S1` was the weakest skill condition on the subset (0.571), below the prompt; this motivated the redesign. `P2` improves on the frozen prompt and `P3` is indistinguishable from it, so the discovery/bounded-wait direction added nothing measurable. `S2` improved on the frozen skill by cutting reference-loading overhead; `S9`, the pruned bundle with a routing block and three subsystem skills, reached 35/54. The full nine-skill `S10` then led the subset at 0.777 and 39/54, +0.206 macro and +14 full successes over `S1`, and ahead of both prompts. `S11`'s harder wording lost 0.146 macro and nine full successes to `S10` on the same scenarios.

### Full-set comparison

All rows are 46 scenarios with three attempts (138 attempts). Baseline, `P2` and `S2` come from the first full comparison.

| Condition   | Macro score |      Full success | Turn limits | Mean tools | Mean duration |
| ----------- | ----------: | ----------------: | ----------: | ---------: | ------------: |
| baseline    |       0.442 |     53/138 (0.38) |           0 |        5.2 |           65s |
| prompt `P2` |       0.634 |     78/138 (0.57) |           0 |        9.0 |          102s |
| skill `S2`  |       0.637 |     78/138 (0.57) |           0 |       12.3 |          127s |
| skill `S4`  |       0.671 |     83/138 (0.60) |           1 |       11.2 |          127s |
| skill `S5`  |       0.678 |     85/138 (0.62) |           2 |       10.7 |          121s |
| skill `S8`  |       0.696 |     85/138 (0.62) |           2 |       11.6 |          132s |
| skill `S11` |       0.700 |     88/138 (0.64) |           3 |       12.4 |          133s |
| skill `S10` |   **0.710** | **90/138 (0.65)** |           5 |       11.9 |          129s |

The selected `S10` leads the prompt agent by 0.076 macro score and twelve full successes, and the strongest earlier candidate `S8` by 0.014 and five full successes. `S11` was marginally worse than `S10`, so the softer routing instruction is kept. `S1` has no full-set result. None of the complete runs reaches the target of about 80% full success.

### Skill loading

The routing change is visible in the `load_skill` counts. The base skills are loaded in most attempts in every condition; the subsystem skills are essentially absent before `S9` and become common in `S10`/`S11`. Counts are load calls over all attempts of the run (`S2`, `S8`, `S10` and `S11` each 138 attempts; `S9` 54 subset attempts).

| Skill             | `S2` | `S8` | `S9` (subset) | `S10` | `S11` |
| ----------------- | ---: | ---: | ------------: | ----: | ----: |
| `troubleshooting` |  134 |  112 |            46 |   117 |   121 |
| `kubectl`         |  100 |   56 |            25 |    62 |    60 |
| `kubernetes`      |  136 |   38 |            18 |    22 |    16 |
| `workloads`       |    — |    — |             7 |    48 |    50 |
| `networking`      |    — |    1 |             1 |    24 |    26 |
| `scheduling`      |    — |    — |             — |    10 |     9 |
| `configuration`   |    — |    — |             — |     9 |    14 |
| `security`        |    — |    — |             — |     9 |     7 |
| `resources`       |    — |    — |             — |     7 |     8 |
| `authorization`   |    — |    1 |             — |     6 |     4 |
| `storage`         |    — |    — |             1 |     6 |     8 |
| `governance`      |    — |    — |             — |     2 |     5 |
| `connectivity`    |    — |   13 |             — |     — |     — |
| `pod-security`    |    — |    2 |             — |     — |     — |

The `S8`-only skills `connectivity` and `pod-security` were each loaded in a minority of attempts; `playbook` and `bash` were available in `S8`/`S2` but loaded **zero** times and were removed. Routing through the always-loaded `troubleshooting` skill moved the three core subsystem skills from 7 loads over 54 attempts (`S9`) to 48 over 138 attempts (`S10`) for `workloads`, and made all nine reachable. `S11`'s harder wording raised some counts (`configuration` 9→14, `governance` 2→5) without improving the score.

### Effect of loading a subsystem skill

To test whether the subsystem skills help, each skill's loaded attempts are compared against the prompt agent (`P2`) on the **same scenarios**. Macro is the mean of the per-scenario means over the attempts in scope; the prompt column is the prompt's macro over all its attempts on those scenarios.

| Subsystem skill | Attempts where loaded | Macro (loaded) | Full (loaded) | Prompt macro, same scenarios | Δ macro |
| --------------- | --------------------: | -------------: | ------------: | ---------------------------: | ------: |
| `workloads`     |                    48 |          0.714 |         30/48 |                        0.610 |  +0.104 |
| `networking`    |                    24 |          0.791 |         17/24 |                        0.652 |  +0.139 |
| `scheduling`    |                    10 |          0.600 |          7/10 |                        0.633 |  −0.033 |
| `configuration` |                     9 |          0.667 |           5/9 |                        0.556 |  +0.111 |
| `security`      |                     9 |          0.625 |           4/9 |                        0.500 |  +0.125 |
| `resources`     |                     7 |          0.933 |           6/7 |                        0.556 |  +0.378 |
| `authorization` |                     6 |          0.389 |           1/6 |                        0.167 |  +0.222 |
| `storage`       |                     6 |          0.556 |           3/6 |                        0.704 |  −0.148 |
| `governance`    |                     2 |          0.625 |           1/2 |                        0.750 |  −0.125 |

Six of the nine skills show a positive delta against the prompt on the scenarios where they were loaded; `scheduling`, `storage` and `governance` are negative but rest on small counts. Aggregated:

| Group (`S10`)             | Attempts | Macro |         Full | Prompt, same scenarios |
| ------------------------- | -------: | ----: | -----------: | ---------------------: |
| loaded ≥1 subsystem skill |       93 | 0.719 | 59/93 (0.63) |         0.606 (64/120) |
| loaded no subsystem skill |       45 | 0.713 | 31/45 (0.69) |          0.690 (54/87) |

The prompt was weakest exactly where the agent chose to load a subsystem skill (0.606 versus 0.690 on the no-load scenarios), and the loaded attempts recovered to 0.719, 0.113 above the prompt on those scenarios. This comparison is confounded: the agent selects skills for cases it perceives as harder, and the skill/scenario sets are not independent, so the positive deltas bound the observed association rather than establish a per-skill causal effect.

### Reference loading

The descriptive `kubernetes` references do not show the same association. In every skill run, attempts that loaded at least one reference scored **lower** than attempts that loaded none.

| Run (`skill`) | Attempts with ≥1 reference | Macro (with) | Full (with) | Macro (without) | Full (without) |
| ------------- | -------------------------: | -----------: | ----------: | --------------: | -------------: |
| `S2`          |                     74/138 |        0.603 |       40/74 |           0.656 |          38/64 |
| `S4`          |                     71/137 |        0.595 |       41/71 |           0.730 |          42/66 |
| `S5`          |                     57/138 |        0.620 |       32/57 |           0.724 |          53/81 |
| `S8`          |                     46/138 |        0.642 |       23/46 |           0.710 |          62/92 |
| `S10`         |                     64/138 |        0.633 |       37/64 |           0.740 |          53/74 |

The reference loads concentrate on the harder cases and do not compensate for them; this is why the subsystem skills, which carry the repair procedure, rather than the references are the routed content.

### Scenario detail

Mean score and full successes over three attempts for the prompt (`P2`), the strongest earlier candidate (`S8`) and the selected set (`S10`).

| Scenario                           | Prompt `P2` | S8 (candidate) | S10 (selected) |
| ---------------------------------- | ----------: | -------------: | -------------: |
| `antiaffinity-unschedulable`       |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `automount-token-disabled`         |        0.22 |     0.22 (0/3) |     0.00 (0/3) |
| `basic-deploy-service`             |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `bind-precreated-pv`               |        1.00 |     1.00 (3/3) |     0.67 (2/3) |
| `configmap-missing-key`            |        0.67 |     0.50 (1/3) |     1.00 (3/3) |
| `configmap-update-needs-reload`    |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `container-crash-loop`             |        0.33 |     0.33 (1/3) |     0.33 (1/3) |
| `daemonset-missing-toleration`     |        0.00 |     0.00 (0/3) |     0.00 (0/3) |
| `deprecated-api-group`             |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `dns-name-resolution`              |        1.00 |     1.00 (3/3) |     0.67 (2/3) |
| `gateway-httproute-basic`          |        0.50 |     0.50 (1/3) |     0.33 (1/3) |
| `gateway-route-misconfig`          |        1.00 |     1.00 (3/3) |     0.67 (2/3) |
| `headless-service-dns`             |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `hostpath-fsgroup`                 |        0.56 |     0.33 (1/3) |     0.78 (2/3) |
| `image-pull-failure`               |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `immutable-configmap`              |        0.44 |     1.00 (3/3) |     0.33 (0/3) |
| `ingress-path-routing`             |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `job-backoff-exhausted`            |        0.67 |     0.67 (2/3) |     0.67 (2/3) |
| `limitrange-defaults`              |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `liveness-restart-loop`            |        0.67 |     1.00 (3/3) |     1.00 (3/3) |
| `missing-rbac-binding`             |        0.11 |     0.67 (2/3) |     0.78 (1/3) |
| `networkpolicy-blocked`            |        1.00 |     0.67 (2/3) |     0.67 (2/3) |
| `networkpolicy-cross-namespace`    |        0.17 |     1.00 (3/3) |     0.17 (0/3) |
| `networkpolicy-egress-dns`         |        0.00 |     0.27 (0/3) |     0.47 (1/3) |
| `node-selector-no-match`           |        1.00 |     1.00 (3/3) |     0.67 (2/3) |
| `nonroot-securitycontext`          |        0.67 |     0.89 (2/3) |     1.00 (3/3) |
| `oom-memory-limit`                 |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `overprivileged-serviceaccount`    |        0.17 |     0.50 (0/3) |     0.33 (0/3) |
| `private-registry-pull-secret`     |        1.00 |     1.00 (3/3) |     1.00 (3/3) |
| `psa-restricted-namespace`         |        0.78 |     0.56 (1/3) |     0.33 (0/3) |
| `pvc-pending`                      |        0.78 |     0.67 (2/3) |     0.67 (2/3) |
| `qos-guaranteed-requests`          |        1.00 |     0.78 (2/3) |     1.00 (3/3) |
| `readiness-probe-breaks-endpoints` |        0.33 |     0.67 (2/3) |     1.00 (3/3) |
| `readonly-rootfs`                  |        0.00 |     0.00 (0/3) |     0.17 (0/3) |
| `resourcequota-blocked`            |        0.75 |     0.50 (1/3) |     0.75 (2/3) |
| `rollback-bad-release`             |        0.33 |     0.50 (0/3) |     0.83 (2/3) |
| `secret-ref-mismatch`              |        0.78 |     0.78 (1/3) |     0.89 (2/3) |
| `service-targetport-mismatch`      |        0.33 |     1.00 (3/3) |     1.00 (3/3) |
| `sidecar-shared-volume`            |        0.58 |     0.25 (0/3) |     0.67 (2/3) |
| `stale-image-ifnotpresent`         |        0.33 |     0.78 (2/3) |     1.00 (3/3) |
| `startup-probe-slow-app`           |        0.33 |     0.00 (0/3) |     0.33 (1/3) |
| `statefulset-headless-service`     |        0.83 |     0.33 (0/3) |     0.83 (2/3) |
| `statefulset-volumeclaim`          |        0.33 |     0.67 (2/3) |     0.33 (1/3) |
| `stuck-terminating-pod`            |        0.33 |     0.33 (1/3) |     0.33 (1/3) |
| `taint-toleration-placement`       |        0.83 |     1.00 (3/3) |     1.00 (3/3) |
| `unschedulable-cpu-request`        |        0.33 |     0.67 (2/3) |     1.00 (3/3) |

### Compute

The study is 12 benchmark invocations; token accounting covers the **1,601 attempts** that report usage (one `S4` attempt reports none). The model processed **98,592,847 tokens**: 93,829,720 prompt tokens, of which **86,832,116 (92.5%) were served from the prompt cache**, plus 4,763,127 generated tokens. The decoding timings, which count only evaluated (non-cached) prompt tokens, report 6,673,872 prompt and 4,562,404 generated tokens, with 4,318,198 draft tokens of which 2,396,873 were accepted (55.5%).

| Run                | Scope                  |  Attempts |  Prompt tokens |  Cached tokens | Generated tokens |   Total tokens |
| ------------------ | ---------------------- | --------: | -------------: | -------------: | ---------------: | -------------: |
| `reference-subset` | subset, 3 conditions   |       162 |      7,207,868 |      6,581,249 |          450,162 |      7,658,030 |
| `round1`           | subset, 2 conditions   |       108 |      5,844,640 |      5,356,502 |          324,507 |      6,169,147 |
| `round2`           | subset, 2 conditions   |       108 |      6,258,326 |      5,783,679 |          337,515 |      6,595,841 |
| `final`            | full, 3 conditions     |       414 |     15,574,438 |     14,242,898 |        1,103,348 |     16,677,786 |
| `s4`               | full, skill            |       137 |      8,703,861 |      8,066,481 |          418,273 |      9,122,134 |
| `s5`               | full, skill            |       138 |      8,485,409 |      7,868,592 |          405,552 |      8,890,961 |
| `s6`               | scenario subset, skill |        28 |      1,724,237 |      1,602,636 |           80,901 |      1,805,138 |
| `s7`               | scenario subset, skill |        38 |      2,549,980 |      2,347,849 |          119,822 |      2,669,802 |
| `s8`               | full, skill            |       138 |     10,166,607 |      9,481,946 |          446,760 |     10,613,367 |
| `s9`               | subset, skill          |        54 |      4,392,665 |      4,106,111 |          181,483 |      4,574,148 |
| `s10`              | full, skill            |       138 |     11,052,493 |     10,291,873 |          445,553 |     11,498,046 |
| `s11`              | full, skill            |       138 |     11,869,196 |     11,102,300 |          449,251 |     12,318,447 |
| **Total**          |                        | **1,601** | **93,829,720** | **86,832,116** |    **4,763,127** | **98,592,847** |

## What worked and what did not

What worked:

- **Routing the skills moved the needle.** Carrying the symptom-to-skill table in the always-loaded `troubleshooting` skill raised subsystem-skill loads from single digits (`S9`) to `workloads` 48, `networking` 24 and `scheduling` 10 (`S10`). The largest gains are on procedure-heavy scenarios: `configmap-missing-key` (+0.50), `statefulset-headless-service` (+0.50), `hostpath-fsgroup` (+0.44), `sidecar-shared-volume` (+0.42), `rollback-bad-release` (+0.33), `readiness-probe-breaks-endpoints` (+0.33) and `unschedulable-cpu-request` (+0.33).
- **A lean prompt with skills does not lose to a heavier prompt.** The selected `S10` keeps the system prompt close to the previous frozen prompt and puts the procedure in skills; it improved on both `S1` and the prompt agent.
- **Loading a subsystem skill is associated with beating the prompt.** On the scenarios where a subsystem skill was loaded, `S10` scored 0.719 versus 0.606 for the prompt; six of nine skills show a positive scenario-matched delta.

What did not:

- **Reference loading did not help.** Attempts that loaded a `kubernetes` reference scored below attempts that did not in every skill run.
- **Enforcing the load did not help.** `S11`, which states that loading the matching subsystem skill is required, scored 0.700/88 versus 0.710/90 for the softer `S10`, and lost nine full successes to `S10` on the subset.
- **Regressions remain.** `S10` loses ground on `networkpolicy-cross-namespace` (1.00 to 0.17) and `immutable-configmap` (1.00 to 0.33), and slips on `dns-name-resolution`, `gateway-route-misconfig`, `node-selector-no-match`, `bind-precreated-pv` and `statefulset-volumeclaim`. With three attempts these are noisy, but they show the subsystem skills are not uniformly better than the previous generic guidance.
- **Unsolved by every condition:** `daemonset-missing-toleration` and `automount-token-disabled` reach no full success, and `readonly-rootfs`, `overprivileged-serviceaccount`, `psa-restricted-namespace`, `gateway-httproute-basic` and `stuck-terminating-pod` remain weak.

## Conclusions

The frozen skill condition `S1` was the weakest skill set on the subset (0.571, 25/54), below the prompt. The redesign moved from reloading descriptive references toward always-loaded procedural subsystem skills routed by symptom. The selected `S10` improved to **0.777 and 39/54** on the subset — +0.206 macro and +14 full successes over `S1` and ahead of both prompt variants — and to **0.710 and 90/138** on the full benchmark, +0.076 and +12 full successes over the prompt agent and +0.014 and +5 over the strongest earlier candidate `S8`. The strongest evidence ties the gain to the routed subsystem skills: the loaded subsystem attempts beat the prompt on the same scenarios, while the descriptive references and the enforced-load variant added nothing.

`S10` is chosen and frozen as the development skill condition, with `bash`, `playbook`, `connectivity` and `pod-security` removed. The target of about 80% full success was not reached; the residual failures are concentrated in scenarios no condition solves and in per-scenario noise.

## Limitations

`S1` was only run on the 18-scenario subset, so the previous frozen set can be compared with `S10` on the subset only; the full-set comparison uses the prompt agent and the strongest candidate `S8`. The attribution to pruning and routing is a hypothesis rather than an isolated effect: the pure prune-only and prompt-routing-only arms were dropped without being run, so the design does not separate the contribution of removing unused skills from that of adding routed subsystem skills. Three attempts per scenario reveal failure modes but cannot establish small differences; `S10` and `S11` are within two full successes on the full set and their ordering is not significant. The per-skill effect table is an association, not a controlled comparison: skills are loaded for cases the agent judges harder, and the loaded skill/scenario sets overlap. Runs were sequential and unpaired, and the baseline and prompt rows come from an earlier invocation rather than a single mixed run. Multi-token prediction was enabled for all runs. Conclusions describe the best observed configuration under this budget, not a globally optimal one.

## Raw evidence

Each directory contains `run.json`, `results.json`, per-attempt records and the configuration used.

| Study                                        | Raw results                                     |
| -------------------------------------------- | ----------------------------------------------- |
| Reference subset (baseline, `P1`, `S1`)      | [`raw/reference-subset/`](raw/reference-subset) |
| Round 1 subset (`P2`, `S2`)                  | [`raw/round1/`](raw/round1)                     |
| Round 2 subset (`P3`, `S3`)                  | [`raw/round2/`](raw/round2)                     |
| First full comparison (baseline, `P2`, `S2`) | [`raw/final/`](raw/final)                       |
| Full skill `S4`                              | [`raw/s4/`](raw/s4)                             |
| Full skill `S5`                              | [`raw/s5/`](raw/s5)                             |
| Skill `S6`                                   | [`raw/s6/`](raw/s6)                             |
| Skill `S7`                                   | [`raw/s7/`](raw/s7)                             |
| Full skill `S8`                              | [`raw/s8/`](raw/s8)                             |
| Subset skill `S9`                            | [`raw/s9/`](raw/s9)                             |
| Full skill `S10` (selected, frozen)          | [`raw/s10/`](raw/s10)                           |
| Full skill `S11`                             | [`raw/s11/`](raw/s11)                           |
