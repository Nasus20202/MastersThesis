# Gateway probe cleanup

A scenario fix for `gateway-route-misconfig`, made overnight on the researcher's
request and not yet reviewed. The rule behind it is D-053: scenarios must not
leak their problem to the agent.

## Question

Agents in `gateway-route-misconfig` often saw a Pod named `route-probe` when
they listed the cluster. That Pod is the grader's request probe from the setup
checks.

1. How often did the agent see it?
2. How can it be removed without changing what the grade measures?

## Method

**Counting.** Every committed attempt of the scenario was searched for
`route-probe` in a tool output.

**Cause.** The setup checks (`verify_clean` and `verify_fault`) start a busybox
Pod that runs `sleep 3600`. On exit they delete it with `--wait=false`.
`sleep` ignores SIGTERM, so the Pod stays Terminating for its 30 s grace
period, and the agent starts inside that window.

**Fix.**

- Both scripts now delete the probe with `--now` and wait for the deletion
  before they exit. They also delete the probe's events, which otherwise stay
  visible to `kubectl get events` for about an hour.
- The faster cleanup exposed a grading race, described in the results. To
  close it, the request check now passes only after 8 successes in a row,
  about 2 s apart; any failure resets the count. The time limit grows from 90 s
  to 120 s.

**Validation.** `make benchmark-validate` on the scenario's three declared
repairs, in five versions of the scripts: unchanged, waiting cleanup only,
waiting cleanup plus a route-generation wait (dropped afterwards), waiting
cleanup plus 8 successes, and the final version that also deletes the probe's
events. The validation runs are not kept as evidence: CI runs the same
validation on every push.

**Agent check.** Two runs of Qwen3.5 4B, prompt agent, on the scenario,
searched for `route-probe` in every tool output:

| Run (in [raw/](raw/))          | Scripts                        | Attempts                                     |
| ------------------------------ | ------------------------------ | -------------------------------------------- |
| `run-2026-11-18-03-50-20-765Z` | pod cleanup only               | 2 attempts and 1 cluster error, then stopped |
| `run-2026-11-18-04-09-28-337Z` | final (pod and events cleanup) | 2                                            |

## Results

**The agent saw the grader's probe in 54 of 213 attempts.**

| Model       | Attempts that saw it | Turn limit when seen | Turn limit when not seen |
| ----------- | -------------------: | -------------------: | -----------------------: |
| Qwen3.5 4B  |             28 of 43 |             24 of 28 |                 10 of 15 |
| Qwen3.5 9B  |             10 of 39 |              6 of 10 |                 10 of 29 |
| Gemma 4 E4B |            14 of 112 |                    0 |                        0 |
| Gemma 4 E2B |              2 of 19 |                    0 |                        0 |

- It was usually seen in the first tool call, in the output of
  `kubectl get pods`.
- Qwen3.5 4B named the Pod in its own commands in 11 attempts, for example
  `kubectl exec route-probe -- …` and `kubectl get pod route-probe -o yaml`.
  The Pod then disappeared.
- Qwen attempts that saw the probe ended at the turn limit more often. This is
  an association, not a measured effect: seeing it depends on how early the
  agent lists Pods.

**The faster cleanup exposed a grading race.** In the declared repair
`partial-repair-wrong-path`, the route is fixed and then its path is set to
`/wrong`. The expected score is 0.333, because the request should fail.

| Scripts                    | `partial-repair-wrong-path` score |
| -------------------------- | --------------------------------: |
| Unchanged                  |                  0.333 (expected) |
| Waiting cleanup only       |                             1.000 |
| Plus route-generation wait |                      1.000, 1.000 |
| Plus 8 successes (final)   |                      0.333, 0.333 |

- **Why it scored 1.000.** The old check passed on the first successful
  request, and Traefik still served the previous route for a few seconds after
  the patch.
- **Why the old scripts got it right.** The grading check first deleted the
  Terminating probe left by the setup check. That took about 30 s and hid the race.
- **The generation wait did not help.** The HTTPRoute reports the new
  generation as Accepted before Traefik serves it.
- **The final version passes validation.** With 8 successes it passes all six
  attempts with the expected scores: 1, 0.333 and 0 for full repair, wrong path
  and broken. The final scripts with the events cleanup pass all three cases
  again.

**Deleting the Pod alone was not enough.** In the run with only the Pod
cleanup, the Pod was gone, but `kubectl get events` still showed
`pod/route-probe` Scheduled and Pulling. The agent queried `route-probe` at
turn 45.

**With the final scripts, neither attempt saw the probe**, although they listed
Pods or events 5 and 7 times.

| Attempt | Ended at       | Turns | Score |
| ------- | -------------- | ----: | ----: |
| 1       | its own report |    32 |     1 |
| 2       | the turn limit |    50 |     1 |

Removing the probe does not by itself stop Qwen's Gateway turn-limit endings;
see the [stopping-rule diagnosis](../stopping-rule/README.md#why-the-failures-persist).

## Other agent-visible hints

A scan of the setup steps of all 46 development scenarios found no other
scenario that leaves a grader object in the cluster. It found a related hint:
six faults inject values whose names state the fault, and the agent sees them
in the cluster.

| Scenario                  | Injected value                     |
| ------------------------- | ---------------------------------- |
| `image-pull-failure`      | image `nginx:does-not-exist`       |
| `gateway-route-misconfig` | parentRef `app-gateway-missing`    |
| `missing-rbac-binding`    | RoleBinding subject `wrong-reader` |
| `configmap-missing-key`   | ConfigMap key `missing`            |
| `secret-ref-mismatch`     | Secret key `missing-key`           |
| `ingress-path-routing`    | Ingress path `/wrong`              |

They were left unchanged. Every earlier study scored these scenarios with these
names, so renaming them would break comparability, and the size of the hint was
not measured. The held-out final set (#247) can avoid such names from the start;
realistic typos or stale names would test diagnosis rather than reading.

## Limitations

- Earlier results on this scenario were graded with the old check. A repair
  the agent breaks within the last seconds before grading could have passed
  then. The validation case shows that this is possible, not that it happened.
- 8 successes about 2 s apart assume that Traefik serves a stale route for less
  than about 16 s; validation and the agent check observed that.
- `gateway-httproute-basic` and `ingress-path-routing` run the same
  single-success request check, but only in grading, so their agents never see
  the probe. Their declared repairs all validate with the expected scores
  without changes. This does not
  exclude the race: none of their partial repairs breaks a route that has just
  worked. They are left unchanged.
