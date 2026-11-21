---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/job-backoff-exhausted/prompt/001.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# job-backoff-exhausted · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 1

## Task

A Kubernetes Job has failed and is not completing. Restore the Job so that it completes successfully.

Documented in:

- content/en/docs/concepts/workloads/controllers/job.md: Handling Pod and container failures > Pod backoff failure policy

## Write commands

1. `kubectl patch job app -n default --type=json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/args","value":["exit 0"]}]' 2>`
2. `write_file {"content": "apiVersion: batch/v1\nkind: Job\nmetadata:\n name: app\n namespace: default\nspec:\n backoffLimit: 0\n completionMode: NonIndexed\n completions: 1\n parallelism: 1\n podReplacementPolicy: TerminatingOrFailed\n suspend: false\n template:\n spec:\n containers:\n - args:\n - exit 0\n command:\n - sh\n - -c\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n name: app\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n dnsPolicy: ClusterFirst\n restartPolicy: Never\n schedulerName: default-scheduler\n securityContext: {}\n terminationGracePeriodSeconds: 30\n", "path": "/tmp/app-job.yaml"}`
3. `kubectl delete job app -n default --wait=true --timeout=60s`
4. `kubectl apply -f app-job.yaml`

## Final report

The Job now completes successfully, verified directly against the Job's status.

## Root cause

The Job `default/app` had a container command `sh -c "exit 1"` — a command that always exits non-zero. With `backoffLimit: 0`, the first pod failure immediately exhausted retries, and the Job was permanently marked:

```
Failed  reason=BackoffLimitExceeded  message="Job has reached the specified backoff limit"
failed: 1   succeeded: (none)   pods: app-2kpkx Error
```

No ConfigMap, service, or other manifest referenced an alternate command, so the faulty pod command in the Job itself was the only cause.

## What I changed

`spec.template` of a Job is immutable (`field is immutable` when patching `args`), so the persistent Job had to be recreated with the same name and an identical spec, changing only the exit code:

- `args: ["exit 1"]` → `args: ["exit 0"]`

Everything else was preserved verbatim: `backoffLimit: 0`, `completions: 1`, `parallelism: 1`, `completionMode: NonIndexed`, `podReplacementPolicy: TerminatingOrFailed`, `restartPolicy: Never`, image `nginx:1.31.5`, container name `app`.

```
job.batch "app" deleted from default namespace
job.batch/app created
```

## Verification (fresh observations)

- Desired vs. actual completions: `kubectl get job app` → **STATUS Complete, COMPLETIONS 1/1** (`spec.completions=1`, `status.succeeded=1`, `status.failed` absent, `status.ready=0`, `completionTime` set).
- Terminal condition: `type=Complete, status=True, reason=CompletionsReached, message="Reached expected number of succeeded pods"`.
- Owning pod: `app-pq2rg` → **STATUS Completed**, 0 restarts (exit code 0, so no failure/backoff loop).
- Re-read spec after convergence confirms the corrected command (`cmd=["sh","-c"] args=["exit 0"]`) and that the original limits/config (`backoffLimit=0`, `completions=1`, `parallelism=1`, `restartPolicy=Never`, image `nginx:1.31.5`) are intact — I did not relax the backoff limit or change anything unrelated.

The Job stayed `Complete 1/1` on a re-check 15s later, so this is a lasting fix rather than a transient restart.
