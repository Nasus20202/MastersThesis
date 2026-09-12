package command

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSpecValidateRejectsBlankProgram(t *testing.T) {
	assert.Error(t, (Spec{Program: "  "}).Validate())
}

func TestLocalExecutorCapturesOutput(t *testing.T) {
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf stdout; printf stderr >&2"},
	})
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "stdout", result.Stdout)
	assert.Equal(t, "stderr", result.Stderr)
}

func TestLocalExecutorReturnsExitCodeAndOutput(t *testing.T) {
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf stdout; printf stderr >&2; exit 7"},
	})
	assert.Error(t, err)
	assert.Equal(t, 7, result.ExitCode)
	assert.Equal(t, "stdout", result.Stdout)
	assert.Equal(t, "stderr", result.Stderr)
	var executionErr *ExecutionError
	assert.True(t, errors.As(err, &executionErr))
	assert.Equal(t, result, executionErr.Result)
}

func TestLocalExecutorLogsArgumentsOnCompletion(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(previousLogger)

	args := []string{"-c", "printf test"}
	_, err := (LocalExecutor{}).Run(context.Background(), Spec{Program: "sh", Args: args})

	require.NoError(t, err)
	assert.Contains(t, logs.String(), `"msg":"command completed"`)
	assert.Contains(t, logs.String(), `"args":["-c","printf test"]`)
}

func TestLocalExecutorLogsArgumentsOnFailure(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(previousLogger)

	args := []string{"-c", "exit 7"}
	_, err := (LocalExecutor{}).Run(context.Background(), Spec{Program: "sh", Args: args})

	require.Error(t, err)
	assert.Contains(t, logs.String(), `"msg":"command exited with non-zero status"`)
	assert.Contains(t, logs.String(), `"args":["-c","exit 7"]`)
}

func TestLocalExecutorUsesDirectoryAndEnvironment(t *testing.T) {
	directory := t.TempDir()
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf '%s:%s' \"$PWD\" \"$TEST_VALUE\""},
		Dir:     directory,
		Env:     map[string]string{"TEST_VALUE": "configured"},
	})
	require.NoError(t, err)
	assert.Equal(t, directory+":configured", strings.TrimSpace(result.Stdout))
}

func TestSpecValidateRejectsInvalidEnvironment(t *testing.T) {
	assert.Error(t, (Spec{Program: "sh", Env: map[string]string{"BAD=KEY": "value"}}).Validate())
	assert.Error(t, (Spec{Program: "sh", Env: map[string]string{"KEY": "bad\x00value"}}).Validate())
}

func TestLocalExecutorHonorsContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	result, err := (LocalExecutor{}).Run(ctx, Spec{
		Program: "sh",
		Args:    []string{"-c", "sleep 1"},
	})
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, -1, result.ExitCode)
}

func TestLocalExecutorPreservesInheritedEnvironment(t *testing.T) {
	result, err := (LocalExecutor{Environment: []string{"COMMAND_TEST_INHERITED=present"}}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf '%s' \"$COMMAND_TEST_INHERITED\""},
	})
	require.NoError(t, err)
	assert.Equal(t, "present", result.Stdout)
}
