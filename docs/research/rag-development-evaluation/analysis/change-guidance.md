# Change guidance: `R4` and `P4`

`R4` adds one paragraph to `R1`; `P4` adds the same paragraph to the prompt
agent and ran in the same invocation. The paragraph: no editor; a merge patch
replaces whole lists, a strategic merge patch merges `containers` and
`volumes` by name; JSON patch shape; read the rejected field before retrying;
a workload is repaired only when `kubectl rollout status` succeeds. Counts use
a text heuristic over the `kubectl` commands in each attempt (46 scenarios ×
3).

| Condition   | Macro | Full success | `kubectl` writes failed | `kubectl edit` | Rollout status (attempts) | Completed without full success | Mean prompt tokens |
| ----------- | ----: | -----------: | ----------------------: | -------------: | ------------------------: | -----------------------------: | -----------------: |
| prompt      | 0.634 |       78/138 |           163/375 (43%) |             19 |                        24 |                             57 |             35,450 |
| `P4`        | 0.699 |       85/138 |           171/380 (45%) |              0 |                        40 |                             48 |             35,376 |
| `R1`        | 0.725 |       91/138 |           186/422 (44%) |             17 |                        17 |                             45 |             44,684 |
| `R4`        | 0.687 |       86/138 |           192/403 (48%) |              0 |                        54 |                             48 |             46,203 |
| skill `S10` | 0.710 |       90/138 |           153/341 (45%) |              2 |                        58 |                             44 |             80,091 |

`kubectl edit` disappeared and rollout checks increased, but failed writes did
not decrease. `R4` − `P4` in the same run: −0.012 [−0.093, +0.070]; `P4` −
prompt across runs: +0.065 [−0.009, +0.138].
