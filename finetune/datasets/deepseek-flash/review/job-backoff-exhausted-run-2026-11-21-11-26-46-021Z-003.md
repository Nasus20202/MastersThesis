---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/job-backoff-exhausted/prompt/003.json
verdict: accepted
reason: 'first review of a replacement attempt against the raw record by a Haiku model: ok: Recreate justified: "kubectl patch" of spec.template rejected as immutable; recreated spec matches original; Complete 1/1 verified.'
---

# job-backoff-exhausted · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 3

## Task

A Kubernetes Job has failed and is not completing. Restore the Job so that it completes successfully.

Documented in:

- content/en/docs/concepts/workloads/controllers/job.md: Handling Pod and container failures > Pod backoff failure policy

## Write commands

1. `kubectl patch job app -n default --type=json -p '[{"op":"replace","path":"/spec/template/spec/containers/0/args","value":["exit 0"]}]' 2>`
2. `write_file {"content": "apiVersion: batch/v1\nkind: Job\nmetadata:\n name: app\n namespace: default\n labels:\n batch.kubernetes.io/job-name: app\n job-name: app\nspec:\n backoffLimit: 0\n completionMode: NonIndexed\n completions: 1\n manualSelector: false\n parallelism: 1\n podReplacementPolicy: TerminatingOrFailed\n suspend: false\n template:\n metadata:\n labels:\n batch.kubernetes.io/job-name: app\n job-name: app\n spec:\n containers:\n - args:\n - exit 0\n command:\n - sh\n - -c\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n name: app\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n dnsPolicy: ClusterFirst\n restartPolicy: Never\n schedulerName: default-scheduler\n securityContext: {}\n terminationGracePeriodSeconds: 30\n", "path": "/tmp/job-fixed.yaml"}`
3. `kubectl delete job app -n default --cascade=foreground --wait=true --timeout=60s 2>`
4. `kubectl apply -f job-fixed.yaml 2>`

## Final report

The Job is now completing successfully.

## Diagnosis

`kubectl get jobs -A` showed `default/app` in `Failed` status with `0/1` completions, condition `Failed / BackoffLimitExceeded`.

Evidence gathered:

- **Job spec**: `backoffLimit: 0`, `completions: 1`, `restartPolicy: Never`, image `nginx:1.31.5`, command `sh -c` with arg `exit 1`.
- **Pod `app-vg574`**: terminated with `Reason: Error, Exit Code: 1`, no logs, image pulled fine, pod scheduled fine — so the failure was deterministic application behavior, not scheduling, image, or resource pressure.
- **Events**: the same job _name_ had an earlier pod `app-q22gn` that ran and produced `Job completed`; one second later a new pod `app-vg574` was created that immediately failed and exhausted the backoff. This showed the working job had been replaced by one with a broken container command.
- **Immutability check**: patching `spec.template` was rejected (`field is immutable`), so the pod template could not be corrected in place.

## Fix applied

Deleted the failed Job (removing its failed pod) and recreated it with the identical spec — same name/namespace/labels, `image: nginx:1.31.5`, `command: ["sh","-c"]`, `backoffLimit: 0`, `completions: 1`, `parallelism: 1`, `restartPolicy: Never` — changing only the container arg from `exit 1` to `exit 0`, the single defect making the Job unable to succeed.

## Verification (fresh observations)

- Job status: `app  Complete  1/1` (duration 4s), stable on a re-check 15s later.
- Job conditions: `SuccessCriteriaMet=True (CompletionsReached)`, `Complete=True (CompletionsReached)`.
- Counters: `succeeded=1 failed=<none> ready=0` — requested completions (1) matched by successful completions.
- Pod `app-kdcnh`: `phase=Succeeded`, container `state=Completed`, `exitCode=0`.
- Event log for the job ends with `SuccessfulCreate → Created pod: app-kdcnh` and `Completed / Job completed`.

All other settings and limits (backoff limit of 0, single completion, image, restart policy) were left intact.
