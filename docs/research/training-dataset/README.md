# Training datasets from two teachers

Ticket #25 builds the training data for the fine-tuning of #26 (D-054): successful agent trajectories on the 38 training scenarios, collected from two teachers, filtered, reviewed and stored with their provenance. This study records how the data were collected and what they contain. The datasets are in [`finetune/datasets/`](../../../finetune/datasets/); the code is in [`finetune/src/finetune/data/`](../../../finetune/src/finetune/data/).

## Summary

- **DeepSeek V4.1 Flash covers all 38 training scenarios; Qwen3.5 9B covers 37.** After review the datasets hold 185 DeepSeek and 156 Qwen trajectories. DeepSeek has 5 per scenario on 33 scenarios and 4 on the other 5. Qwen has 5 on 21 scenarios, 4 on 10, 3 on 2, 2 on 1, 1 on 3 (`gateway-httproute-basic`, `networkpolicy-egress-dns`, `networkpolicy-cross-namespace`) and none on `readonly-rootfs`.
- **Review removed 18 DeepSeek and 76 Qwen attempts.** The grader's full score does not catch final reports that claim more than the commands show, forceful or recreating writes, unexplained values, or changes outside the task. Qwen has far more of these than DeepSeek, so its data shrank by a third.
- **DeepSeek is the cheaper and more reliable teacher.** 304 attempts cost about $1 off-peak. Qwen needed three Kaggle passes (about 7 hours of Kaggle session time).
- **DeepSeek leaves reasoning empty on some turns.** At the default effort, 27% of assistant turns of the sweep trajectories carry no reasoning (Qwen: none). The raw records keep reasoning as returned; whether to drop or truncate affected examples is decided later.
- **Most examples fit the 16,384-token cap.** The freeze leaves out 4 DeepSeek and 3 Qwen examples longer than the cap, never cutting them: 181 and 153 examples remain for each of the four students.

## Question

1. Can each teacher provide up to 5 clean trajectories on every training scenario under the D-054 filter?
2. What do the datasets contain, and what does reviewing them show?

## Method

**Teachers.** Both run the frozen prompt agent on the 38 training scenarios (the 46 development scenarios minus the 8 validation scenarios of D-055; the held-out set is not used). DeepSeek V4.1 Flash (`deepseek-flash`, revision 1017) runs through its official API with thinking mode at the default effort. Qwen3.5 9B (`Qwen3.5-9B-Q4_K_M`) runs on Kaggle through the repository's "Benchmark on Kaggle" workflow. 8 attempts per scenario; extra attempts were added where a scenario stayed below 5.

**Filter and selection (D-054).** An attempt is kept when it has a full score, ended by the agent itself, did not overflow the context and made no malformed tool call. Two attempts are the same trajectory when their sequences of write commands are equal after whitespace normalisation; of those the one with fewer tokens stays. Up to 5 trajectories per scenario are kept, fewest tokens first. Infrastructure failures (a cluster that did not start, a lost model server) were run again; they are not scored attempts.

**Storage.** Every attempt of a teacher's runs is stored under `finetune/datasets/<teacher>/runs/`, the kept and the unkept ones, with each run's `run.json` and `results.json`. The manifest lists the kept attempts with run, attempt, tokens and review status. The selection can therefore be redone with another cap or filter without the benchmark results.

**Review.** Each kept attempt has a sheet in `<teacher>/review/` with the task, the documented repair, the write commands and the final report; the verdict (`accepted` or `rejected`) is set in the sheet. Every sheet is reviewed for both teachers. A rejected attempt is replaced by the next candidate of its scenario, which gets its own sheet. The freeze refuses to run while a sheet is pending or rejected.

**Freeze.** `python -m finetune.data.freeze --name <teacher>` renders the examples for each of the four students (D-053) with their chat templates, leaves out examples longer than `training.max_len`, and records counts per scenario, scenarios without examples and a hash.

## Results

### Collection and review

|                                           | DeepSeek V4.1 Flash                                           | Qwen3.5 9B                               |
| ----------------------------------------- | ------------------------------------------------------------- | ---------------------------------------- |
| Attempts run on training scenarios        | 304 sweep, 10 retried after infrastructure failures, 24 extra | 304 + 48 + 24 across three Kaggle passes |
| Kept before review (up to 5 per scenario) | 190                                                           | 187                                      |
| Attempts rejected in review               | 18                                                            | 76                                       |
| Kept after review                         | 185                                                           | 156                                      |
| Scenarios with 5                          | 33 of 38                                                      | 21 of 38                                 |
| Scenarios without examples                | none                                                          | `readonly-rootfs`                        |
| Cost                                      | about $1.10                                                   | free (Kaggle GPU time)                   |

DeepSeek's cost is measured from the API usage: prompt-cache hit rate 90%, off-peak price (the work ran on a Saturday); at peak price it would be about twice that. Qwen's `networkpolicy-cross-namespace` first had no clean attempt (all 8 hit the turn limit, 4 of them with a full score they never stopped on); the extra passes reached 3 clean attempts. `readonly-rootfs` gave 2 clean attempts in 24 tries, and both were rejected in review.

Haiku models reviewed every sheet; a second pass re-checked the flagged ones against the full raw record, and a rejected attempt was replaced by the next candidate, which was reviewed the same way, until none was pending. A sheet is rejected when the final report states something the record does not support, when the writes change resources the task did not ask about or set values with no explanation, or when it was a forceful or recreating write that a plain command would have handled and the report is also wrong. Attempts whose only issue is a forceful write (`kubectl replace --force`, `delete --force --grace-period=0`) were accepted for now, by the researcher's decision; a text scan found such writes on 31 of 187 Qwen sheets against 2 of 190 DeepSeek sheets. One DeepSeek attempt was rejected on reading: it raised the node kernel's inotify limit through a privileged pod to repair kube-proxy, an effect of running many clusters on one host.

### Reasoning on DeepSeek turns

At the default effort the model returns an empty `reasoning_content` on 460 of 1,689 assistant turns (27%) of the kept sweep trajectories, spread over 170 of 186 attempts. An exploratory run of 5 attempts at maximum effort had reasoning on all 52 turns, but with longer context (11–17k against 8–10k tokens). The default effort is kept so the trajectories are not driven to overthink; the raw records keep the reasoning as returned.

### Length and freeze

Median rendered length is about 8.7k tokens for the Qwen students and 8.5k for the Gemma students. The freeze left out the examples above the 16,384-token cap: 4 of 185 DeepSeek and 3 of 156 Qwen attempts, the same for every student. Examples per student after the freeze: 181 (DeepSeek) and 153 (Qwen). The dataset revision hashes are in each teacher's `freeze.json`.

## Conclusions

- The filter's full score is not enough on its own: review rejected 9% of DeepSeek and 33% of Qwen attempts. Qwen's final reports and writes are the weaker source, and its dataset has thin scenarios (`networkpolicy-egress-dns`, `networkpolicy-cross-namespace`, `gateway-httproute-basic` with 1 each, `readonly-rootfs` with none).
- DeepSeek's data are more complete. The teacher comparison of #26 therefore compares a larger, more even DeepSeek set with a smaller, patchier Qwen set; a difference in the adapters can come from the data volume as well as the teacher.
- The cap of 5 stays. Raising it would add only the longer, more wandering trajectories and favour easy scenarios; it is a candidate if the adapters underfit on the validation scenarios of #26.
- D-054 now says every trajectory of both teachers is reviewed (it said a random 20% of the Qwen ones): the 20% audit could not have removed the defects the full review found. The cap stays at 5; forceful writes are noted and not a reason to reject (proposed, not in the decision log).

## Limitations

- The review is model output, not the researcher's. Two false flags found in a spot check (commands missing from the sheet but present in the raw record) show the reviewers can be wrong in both directions, and the reviewers applied the forceful-write rule unevenly across batches. The accepted sheets were not re-checked; the researcher can overrule any verdict in the sheet and rebuild.
- Only the write commands are listed in a sheet; reads and shell-built files are not. A claim in a final report can be correct and still look unsupported on the sheet.
- DeepSeek trajectories carry the teacher's reasoning only where the model returned it (73% of turns). Qwen's carry reasoning on every turn.
- The 8 validation scenarios and the held-out set are excluded by construction; `finetune/tests/test_dataset.py` checks that no validation scenario is in a manifest.

## Evidence

The collected attempts, the manifests and the review sheets are in [`finetune/datasets/`](../../../finetune/datasets/README.md).
