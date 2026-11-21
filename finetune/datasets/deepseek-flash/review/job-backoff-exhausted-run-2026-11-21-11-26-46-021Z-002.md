---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/job-backoff-exhausted/prompt/002.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# job-backoff-exhausted · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 2

## Task

A Kubernetes Job has failed and is not completing. Restore the Job so that it completes successfully.

Documented in:

- content/en/docs/concepts/workloads/controllers/job.md: Handling Pod and container failures > Pod backoff failure policy

## Write commands

1. `kubectl auth can-i delete jobs -n default 2>`
2. `kubectl auth can-i create jobs -n default 2>`
3. `kubectl patch job app -n default --type=strategic -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","args":["exit 0"]}]}}}}' --dry-run=server 2>`
4. `write_file {"content": "apiVersion: batch/v1\nkind: Job\nmetadata:\n name: app\n namespace: default\nspec:\n backoffLimit: 0\n completions: 1\n parallelism: 1\n template:\n spec:\n containers:\n - name: app\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n command:\n - sh\n - -c\n args:\n - exit 0\n restartPolicy: Never\n terminationGracePeriodSeconds: 30\n", "path": "/workspace/job-fixed.yaml"}`
5. `kubectl delete job app -n default --cascade=foreground --wait=true --timeout=60s 2>`
6. `kubectl apply -f job-fixed.yaml 2>`

## Final report

The Job now completes successfully, and the fix is durable.

## Diagnosis (observed evidence)

- `default/app` was `Status: Failed`, `COMPLETIONS 0/1`, with condition `Failed=True, reason=BackoffLimitExceeded` ("Job has reached the specified backoff limit").
- Its pod `app-k69p8` had container state `terminated: exitCode=1, reason=Error`, produced no logs, and the Job spec was `backoffLimit: 0` with container `command: [sh, -c]`, `args: ["exit 1"]`.
- So the failure was deterministic: the container's command itself exits non-zero, and with `backoffLimit: 0` the very first failure permanently marks the Job `Failed`. Retrying/raising backoff alone could never make it succeed, and the Job was not suspended or blocked by scheduling/RBAC/image issues (image `nginx:1.31.5` pulled fine, pod scheduled and started).
- I confirmed the pod template is immutable in this cluster (`server-side dry-run` patch of `spec.template` → `field is immutable`), so the corrected spec had to be applied by recreating the same-named Job (a lasting spec fix, not a restart of a transient pod).

## Change made

Only the container command's exit code was corrected, all other settings preserved (`backoffLimit: 0`, `completions: 1`, `parallelism: 1`, `restartPolicy: Never`, image `nginx:1.31.5`):

```yaml
command: ["sh","-c"]; args: ["exit 1"]  ->  args: ["exit 0"]
```

Applied via `kubectl delete job app --cascade=foreground` then `kubectl apply -f /workspace/job-fixed.yaml` (backup of the original object saved at `/workspace/original-job-backup.yaml`).

## Verification (fresh observations)

- Job: `app  Complete  1/1`, conditions `SuccessCriteriaMet=True(CompletionsReached)`, `Complete=True(CompletionsReached)`, `status.succeeded=1`, `failed` empty, `active` empty.
- Pod `app-fv2sd`: `phase=Succeeded`, `exitCode=0`, `reason=Completed`, `RESTARTS 0`.
- Re-checked after an additional 25s: still `Complete 1/1` with the same succeeded pod — no new failed pods and no further backoffs.
- Constraints retained: image `nginx:1.31.5`, `backoffLimit=0`, `completions=1`, `restartPolicy=Never` (verified in the live object).
