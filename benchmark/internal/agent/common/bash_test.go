package common

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type bashTestShell struct {
	spec   command.Spec
	result command.Result
	err    error
}

func (s *bashTestShell) Exec(_ context.Context, spec command.Spec) (command.Result, error) {
	s.spec = spec
	return s.result, s.err
}

func TestBashToolExecutesInTheShellBoundary(t *testing.T) {
	shell := &bashTestShell{result: command.Result{
		Stdout:   "output",
		Stderr:   "diagnostic",
		ExitCode: 7,
	}}
	tool, err := NewBashTool(shell)
	require.NoError(t, err)

	result := tool.Execute(context.Background(), inference.ToolCall{
		Type:      "function",
		Name:      "bash",
		Arguments: `{"command":"printf test"}`,
	})

	assert.Equal(t, command.Spec{Program: "bash", Args: []string{"-lc", "printf test"}}, shell.spec)
	assert.Contains(t, result.Content, "exit_code: 7")
	assert.Contains(t, result.Content, "stdout:\noutput")
	assert.Contains(t, result.Content, "stderr:\ndiagnostic")
	details, ok := result.Details.(CommandEvidence)
	require.True(t, ok)
	assert.Equal(t, "printf test", details.Command)
	assert.Equal(t, "output", details.Stdout)
	assert.Equal(t, "diagnostic", details.Stderr)
	assert.Equal(t, 7, details.ExitCode)
}

func TestBashToolReturnsArgumentAndExecutionErrors(t *testing.T) {
	shell := &bashTestShell{err: errors.New("command failed")}
	tool, err := NewBashTool(shell)
	require.NoError(t, err)

	malformed := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{`})
	assert.Error(t, malformed.Error)
	assert.Contains(t, malformed.Content, "malformed bash arguments")

	missing := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{}`})
	assert.EqualError(t, missing.Error, "bash command is required")
	assert.Equal(t, "bash command is required", missing.Content)

	execution := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"command":"false"}`})
	assert.Error(t, execution.Error)
	assert.Contains(t, execution.Content, "error: command failed")
}

func TestFormatCommandResultIncludesOutputAndError(t *testing.T) {
	content := formatCommandResult(command.Result{
		Stdout:   "output",
		Stderr:   "diagnostic",
		ExitCode: 7,
	}, errors.New("command failed"))
	assert.True(t, strings.Contains(content, "exit_code: 7"))
	assert.Contains(t, content, "stdout:\noutput")
	assert.Contains(t, content, "stderr:\ndiagnostic")
	assert.Contains(t, content, "error: command failed")
}
