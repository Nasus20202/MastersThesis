package command

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSpecValidateRejectsBlankProgram(t *testing.T) {
	assert.Error(t, (Spec{Program: "  "}).Validate())
}

func TestLocalExecutorCapturesOutput(t *testing.T) {
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf stdout; printf stderr >&2"},
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "stdout", result.Stdout)
	assert.Equal(t, "stderr", result.Stderr)
}

func TestLocalExecutorReturnsExitCodeAndOutput(t *testing.T) {
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf failed >&2; exit 7"},
	})
	assert.Error(t, err)
	assert.Equal(t, 7, result.ExitCode)
	assert.Equal(t, "failed", result.Stderr)
}

func TestLocalExecutorUsesDirectoryAndEnvironment(t *testing.T) {
	directory := t.TempDir()
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf '%s:%s' \"$PWD\" \"$TEST_VALUE\""},
		Dir:     directory,
		Env:     map[string]string{"TEST_VALUE": "configured"},
	})
	if !assert.NoError(t, err) {
		return
	}
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
	t.Setenv("COMMAND_TEST_INHERITED", "present")

	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf '%s' \"$COMMAND_TEST_INHERITED\""},
	})
	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, os.Getenv("COMMAND_TEST_INHERITED"), result.Stdout)
}
