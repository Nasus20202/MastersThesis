package command

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"
)

func TestSpecValidateRejectsBlankProgram(t *testing.T) {
	if err := (Spec{Program: "  "}).Validate(); err == nil {
		t.Fatal("blank program passed validation")
	}
}

func TestLocalExecutorCapturesOutput(t *testing.T) {
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf stdout; printf stderr >&2"},
	})
	if err != nil {
		t.Fatalf("run command: %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "stdout" || result.Stderr != "stderr" {
		t.Fatalf("output = %#v, want stdout/stderr", result)
	}
}

func TestLocalExecutorReturnsExitCodeAndOutput(t *testing.T) {
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf failed >&2; exit 7"},
	})
	if err == nil {
		t.Fatal("non-zero command succeeded")
	}
	if result.ExitCode != 7 {
		t.Fatalf("exit code = %d, want 7", result.ExitCode)
	}
	if result.Stderr != "failed" {
		t.Fatalf("stderr = %q, want failed", result.Stderr)
	}
}

func TestLocalExecutorUsesDirectoryAndEnvironment(t *testing.T) {
	directory := t.TempDir()
	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf '%s:%s' \"$PWD\" \"$TEST_VALUE\""},
		Dir:     directory,
		Env:     map[string]string{"TEST_VALUE": "configured"},
	})
	if err != nil {
		t.Fatalf("run command: %v", err)
	}
	if got, want := strings.TrimSpace(result.Stdout), directory+":configured"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestSpecValidateRejectsInvalidEnvironment(t *testing.T) {
	if err := (Spec{Program: "sh", Env: map[string]string{"BAD=KEY": "value"}}).Validate(); err == nil {
		t.Fatal("invalid environment key passed validation")
	}
	if err := (Spec{Program: "sh", Env: map[string]string{"KEY": "bad\x00value"}}).Validate(); err == nil {
		t.Fatal("NUL environment value passed validation")
	}
}

func TestLocalExecutorHonorsContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	result, err := (LocalExecutor{}).Run(ctx, Spec{
		Program: "sh",
		Args:    []string{"-c", "sleep 1"},
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v, want deadline exceeded", err)
	}
	if result.ExitCode != -1 {
		t.Fatalf("exit code = %d, want -1", result.ExitCode)
	}
}

func TestLocalExecutorPreservesInheritedEnvironment(t *testing.T) {
	t.Setenv("COMMAND_TEST_INHERITED", "present")

	result, err := (LocalExecutor{}).Run(context.Background(), Spec{
		Program: "sh",
		Args:    []string{"-c", "printf '%s' \"$COMMAND_TEST_INHERITED\""},
	})
	if err != nil {
		t.Fatalf("run command: %v", err)
	}
	if result.Stdout != os.Getenv("COMMAND_TEST_INHERITED") {
		t.Fatalf("output = %q, want inherited value", result.Stdout)
	}
}
