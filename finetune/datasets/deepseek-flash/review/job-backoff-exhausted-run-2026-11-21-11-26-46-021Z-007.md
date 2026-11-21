---
attempt: finetune/datasets/deepseek-flash/runs/run-2026-11-21-11-26-46-021Z/job-backoff-exhausted/prompt/007.json
verdict: accepted
reason: "first-pass review by a Haiku model: no defect found (not a human review)"
---

# job-backoff-exhausted · deepseek-flash · run-2026-11-21-11-26-46-021Z · attempt 7

## Task

A Kubernetes Job has failed and is not completing. Restore the Job so that it completes successfully.

Documented in:

- content/en/docs/concepts/workloads/controllers/job.md: Handling Pod and container failures > Pod backoff failure policy

## Write commands

1. `kubectl patch job app --type=merge -p '{"spec":{"template":{"spec":{"containers":[{"name":"app","args":["exit 0"]}]}}}}' --dry-run=server 2>`
2. `kubectl delete job app --cascade=foreground --wait=true`
3. `kubectl apply -f /tmp/job-fixed.json`

## Final report

The Job now completes successfully.

**Evidence of the fault**

- `Job/app` (default ns) was `Failed` with condition `Failed=True reason=BackoffLimitExceeded` ("Job has reached the specified backoff limit"), `failed: 1`, `completions 0/1`.
- Its pod `app-qtlrt` was `Error`, exit code `1`.
- The only cause was the pod template itself: container `app` ran `sh -c "exit 1"`, i.e. an unconditional failure, so with `backoffLimit: 0` the Job could never succeed. Infrastructure was healthy (scheduled, image `nginx:1.31.5` pulled, container started).

**Repair**

- Job pod templates are immutable (confirmed: a server-side dry-run patch of `spec.template` was rejected with `spec.template: Invalid value ... field is immutable`), so the owning Job had to be rebuilt.
- Saved the Job spec, changed only the container args from `exit 1` to `exit 0`, then deleted the failed Job (`--cascade=foreground`, which also removed its failed pod) and re-applied it under the same name.
- All other settings were kept as-is: image `nginx:1.31.5`, command `sh -c`, `completions: 1`, `parallelism: 1`, `backoffLimit: 0`, `restartPolicy: Never`, labels, namespace `default`.

**Verification (fresh observations)**

- `kubectl wait --for=condition=Complete job/app` returned "condition met".
- `Job/app`: `STATUS Complete`, `COMPLETIONS 1/1`, conditions `SuccessCriteriaMet=True (CompletionsReached)` and `Complete=True (CompletionsReached)`.
- Pod `app-z6d8z`: `phase=Succeeded`, `exitCode=0`.
- Spec check: `image=nginx:1.31.5`, `cmd=["sh","-c"]`, `args=["exit 0"]`, `completions/parallelism=1/1`, `backoffLimit=0` — no unrelated configuration changed.
