---
name: bash
description: "Bash shell syntax and processing: quoting, pipelines, exit status, files, and processes."
---

# Bash

Use this skill for shell knowledge: how expansions, pipelines, streams, and exit statuses behave. What to investigate and in which order is decided by the `troubleshooting` skill.

## Working directory and environment

Command resolution and relative paths depend on the working directory and `PATH`. The environment may carry credentials, so read only the variables a command depends on and never print secret-bearing environments in full:

```bash
pwd
printf 'PATH=%s\n' "$PATH"
command -v kubectl
```

## Quoting, arguments, and filenames

Unquoted expansions split on whitespace and glob characters. Quote every expansion unless splitting or globbing is intentional. Arrays hold argument lists without re-parsing; never pass untrusted text to `eval` or a second shell:

```bash
shopt -s nullglob
files=(/var/log/app/*.log)
for file in "${files[@]}"; do
  grep -n -- 'error' "$file"
done
```

Filenames can contain spaces, newlines, and leading dashes. `find -print0` with `while IFS= read -r -d ''` or `xargs -0 -r` preserves NUL-delimited names, and `--` marks the end of options before paths:

```bash
find /var/log/app -type f -print0 |
  while IFS= read -r -d '' file; do
    sed -n '1,80p' -- "$file"
  done
```

```bash
find /var/log/app -type f -print0 |
  xargs -0 -r grep -nH -- 'error'
```

`grep -nH`, `awk`, and `sed` reduce output while retaining line numbers and filenames. Any reduction (`sort`, `uniq -c`, `awk` aggregation) hides distinctions, so confirm the hidden detail is irrelevant before summarizing.

## Exit status, streams, and pipelines

Every command returns an exit status: `0` for success, nonzero otherwise. `grep` returns `1` for no match and `2` for an error; the two meanings differ. `cmd >out 2>&1` merges stderr into stdout, while `cmd 2>&1 >out` does not. `$(...)` captures stdout; stderr must be redirected separately when it needs collecting:

```bash
if output=$(some-command --json 2>command.err); then
  printf '%s\n' "$output"
else
  status=$?
  printf 'command failed with status %d\n' "$status" >&2
  cat command.err >&2
fi
```

In a pipeline, each stage has its own status; only the last is in `$?` unless `pipefail` is set, and `PIPESTATUS` holds every stage status. `set -o pipefail` makes a pipeline report an intermediate failure; `set -e` terminates the shell on nonzero status, including expected ones such as `grep` finding no match, so it does not belong on ad-hoc probes.

## File and process facts

`stat` reports type, owner, permissions, size, and modification time; `file` identifies content type; `readlink -f` resolves the absolute path through links:

```bash
stat -- path
file -- path
readlink -f -- path
```

`ps` reports process identity, parent, state, elapsed time, and command line; `/proc/$pid/exe` resolves the running binary. `/proc/$pid/environ` holds NUL-separated variables and may contain secrets, so read it only when necessary and redact values. An absent process and an unhealthy running process are different facts:

```bash
ps -eo pid,ppid,stat,etime,args --forest
ps -p "$pid" -o pid,ppid,stat,etime,args
readlink -f "/proc/$pid/exe"
```

## Atomic replacement mechanics

Writing directly into a watched file exposes partial content. The atomic pattern writes a temporary file in the same directory, validates it, keeps a backup, and renames it over the target, since `mv` within a filesystem is atomic. `trap` removes the temporary file on exit:

```bash
tmp=$(mktemp "./config.XXXXXX")
trap 'rm -f -- "$tmp"' EXIT
generate-config >"$tmp"
validate-config "$tmp"
cp -a -- config config.backup
mv -f -- "$tmp" config
trap - EXIT
```

Broad globs, unquoted variables, and `rm -rf` as uncertainty shortcuts widen the blast radius beyond the intended path.
