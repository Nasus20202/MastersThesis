# Benchmark Module — Agent Instructions

Go-specific conventions for `benchmark/`. See the repository root
[AGENTS.md](../AGENTS.md) for the research workflow, approval gate, and
documentation policy, which also apply here.

## Architecture

- `cmd/benchmark` is the CLI: flags, environment variables, and config file
  reading live there. `internal/` is the library: it receives already-resolved
  values as parameters and never reads env vars or config files itself.
- Domain YAML (scenario/validation definitions) is not "config" — loading and
  validating it is `internal/scenario` and `internal/validation`'s job.
- External integrations (cluster, sandbox, inference, ...) live under
  `internal/integrations`. Define the integration's contract as an interface
  plus a `Factory` type in the integration's own package (e.g.
  `sandbox.Sandbox`), then add each concrete implementation in its own
  subpackage (e.g. `sandbox/docker`).

## Code style

- Prefer readable, simple solutions over clever ones.
- Every package has a one-sentence `// Package <name> ...` doc comment on its
  primary file.
- Add comments only when they give value (a "why", an invariant, a gotcha),
  and keep them short: don't restate what the code already says. Keep a
  comment only for ordering, units, a sentinel return, an edge case or a
  rationale — not comments like `// Sum is the total of values.` or
  `// Snapshot returns a loaded run snapshot.`
- `go fmt ./...` and `go vet ./...` (`make format` / `make lint`) must pass
  before committing.

## Dependencies

- Get explicit approval before adding a new dependency to `go.mod`.

## Testing

- New code needs tests, but keep them concise — don't over-verbose test
  setup, table cases, or assertions.
- Every test must exercise some logic and fail under a small change to the
  code it covers. Do not add tests that only assert a constant against
  itself, a constructor's trivial output, or embedded prompt/content text.

## After implementing

- Once a change works and is tested, review it for refactor/simplification
  opportunities before considering it done.
