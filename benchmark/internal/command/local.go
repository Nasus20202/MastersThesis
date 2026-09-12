package command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type LocalExecutor struct {
	Environment []string
}

func (e LocalExecutor) Run(ctx context.Context, spec Spec) (Result, error) {
	if err := spec.Validate(); err != nil {
		return Result{}, err
	}

	process := exec.CommandContext(ctx, spec.Program, spec.Args...)
	process.Dir = spec.Dir
	if e.Environment != nil || len(spec.Env) > 0 {
		process.Env = e.Environment
	}
	if len(spec.Env) > 0 {
		process.Env = mergeEnvironment(e.Environment, spec.Env)
	}
	logger := slog.With("program", spec.Program)
	logger.DebugContext(ctx, "running command",
		"args", spec.Args,
		"dir", spec.Dir,
	)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	process.Stdout = &stdout
	process.Stderr = &stderr

	started := time.Now()
	err := process.Run()
	result := Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode(err),
		Duration: time.Since(started),
	}
	if err == nil {
		logger.InfoContext(ctx, "command completed",
			"args", spec.Args,
			"exit_code", result.ExitCode,
			"duration", result.Duration,
		)
		return result, nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		runErr := fmt.Errorf("run %q: %w", spec.Program, ctxErr)
		logger.ErrorContext(ctx, "command failed",
			"args", spec.Args,
			"exit_code", result.ExitCode,
			"duration", result.Duration,
			"error", runErr,
		)
		return result, &ExecutionError{Spec: spec, Result: result, Err: runErr}
	}
	runErr := fmt.Errorf("run %q: %w", spec.Program, err)
	logger.WarnContext(ctx, "command exited with non-zero status",
		"args", spec.Args,
		"exit_code", result.ExitCode,
		"duration", result.Duration,
		"error", runErr,
	)
	return result, &ExecutionError{Spec: spec, Result: result, Err: runErr}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}

	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		return exitError.ExitCode()
	}
	return -1
}

func mergeEnvironment(base []string, overrides map[string]string) []string {
	result := make([]string, 0, len(base)+len(overrides))
	seen := make(map[string]struct{}, len(base))
	for _, entry := range base {
		key, _, ok := strings.Cut(entry, "=")
		if !ok || key == "" {
			continue
		}
		if _, overridden := overrides[key]; overridden {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, entry)
	}

	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result = append(result, key+"="+overrides[key])
	}
	return result
}
