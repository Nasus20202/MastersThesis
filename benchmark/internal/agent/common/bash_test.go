package common

import (
	"context"
	"encoding/json"
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

func TestBashToolDefinitionAndConstructor(t *testing.T) {
	assert.Equal(t, inference.Tool{
		Name:        "bash",
		Description: "Execute a shell command inside the isolated benchmark sandbox.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute"}},"required":["command"],"additionalProperties":false}`),
	}, BashTool())

	shell := &bashTestShell{}
	tool, err := NewBashTool(shell)
	require.NoError(t, err)
	assert.Equal(t, BashTool(), tool.Definition())

	tool, err = NewBashTool(nil)
	assert.Nil(t, tool)
	assert.EqualError(t, err, "bash shell is required")
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

func TestFormatCommandResultShowsCommandExitWithoutSandboxError(t *testing.T) {
	content := formatCommandResult(command.Result{
		Stdout:   "output",
		Stderr:   "diagnostic",
		ExitCode: 7,
	}, errors.New(`execute sandbox command: run "docker": exit status 7`))
	assert.True(t, strings.Contains(content, "exit_code: 7"))
	assert.Contains(t, content, "stdout:\noutput")
	assert.Contains(t, content, "stderr:\ndiagnostic")
	assert.NotContains(t, content, "execute sandbox command")
}

func TestBashToolKeepsFailedExitInEvidenceWithoutMisleadingAgent(t *testing.T) {
	shell := &bashTestShell{
		result: command.Result{ExitCode: 1},
		err:    errors.New(`execute sandbox command: run "docker": exit status 1`),
	}
	tool, err := NewBashTool(shell)
	require.NoError(t, err)

	result := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"command":"grep -i config-reader"}`})
	assert.Equal(t, "exit_code: 1\nstdout:\n\nstderr:\n\n", result.Content)
	assert.ErrorContains(t, result.Error, "execute sandbox command")
	details, ok := result.Details.(CommandEvidence)
	require.True(t, ok)
	assert.Equal(t, 1, details.ExitCode)
}

func TestFormatCommandResultShowsActualExecutionError(t *testing.T) {
	content := formatCommandResult(command.Result{ExitCode: -1}, errors.New("sandbox unavailable"))
	assert.Contains(t, content, "exit_code: -1")
	assert.Contains(t, content, "error: sandbox unavailable")
}

func TestFormatCommandResultBoundsModelVisibleOutputButKeepsEvidenceRaw(t *testing.T) {
	stdout := strings.Repeat("x", maxModelVisibleCommandOutputBytes*2)
	content := formatCommandResult(command.Result{Stdout: stdout}, nil)

	assert.LessOrEqual(t, len(content), maxModelVisibleCommandOutputBytes)
	assert.Contains(t, content, truncatedCommandOutputMarker)
	assert.Contains(t, content, "exit_code: 0")
}

func TestBashToolPreservesFullOutputInEvidence(t *testing.T) {
	stdout := strings.Repeat("x", maxModelVisibleCommandOutputBytes*2)
	shell := &bashTestShell{result: command.Result{Stdout: stdout}}
	tool, err := NewBashTool(shell)
	require.NoError(t, err)

	result := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"command":"kubectl get -o yaml"}`})
	details, ok := result.Details.(CommandEvidence)
	require.True(t, ok)
	assert.Equal(t, stdout, details.Stdout)
	assert.LessOrEqual(t, len(result.Content), maxModelVisibleCommandOutputBytes)
}
