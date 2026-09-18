package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const bashToolName = "bash"

type bashTool struct{ shell Shell }

func NewBashTool(shell Shell) (Tool, error) {
	if shell == nil {
		return nil, errors.New("bash shell is required")
	}
	return bashTool{shell: shell}, nil
}

func BashTool() inference.Tool {
	return inference.Tool{
		Name:        bashToolName,
		Description: "Execute a shell command inside the isolated benchmark sandbox.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"command":{"type":"string","description":"Shell command to execute"}},"required":["command"],"additionalProperties":false}`),
	}
}

func (bashTool) Definition() inference.Tool { return BashTool() }

// CommandEvidence preserves raw shell evidence separately from the model-
// visible formatted output.
type CommandEvidence struct {
	Command         string  `json:"command,omitempty"`
	Stdout          string  `json:"stdout"`
	Stderr          string  `json:"stderr"`
	ExitCode        int     `json:"exit_code"`
	DurationSeconds float64 `json:"duration_seconds"`
}

func (t bashTool) Execute(ctx context.Context, call inference.ToolCall) ToolResult {
	var arguments struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		return ToolResult{Content: fmt.Sprintf("malformed bash arguments: %v", err), Error: fmt.Errorf("malformed bash arguments: %w", err)}
	}
	if strings.TrimSpace(arguments.Command) == "" {
		err := errors.New("bash command is required")
		return ToolResult{Content: err.Error(), Error: err}
	}

	commandResult, execErr := t.shell.Exec(ctx, command.Spec{
		Program: "bash",
		Args:    []string{"-lc", arguments.Command},
	})
	details := CommandEvidence{
		Command:         arguments.Command,
		Stdout:          commandResult.Stdout,
		Stderr:          commandResult.Stderr,
		ExitCode:        commandResult.ExitCode,
		DurationSeconds: commandResult.Duration.Seconds(),
	}
	return ToolResult{
		Content: formatCommandResult(commandResult, execErr),
		Details: details,
		Error:   execErr,
	}
}

func formatCommandResult(result command.Result, execErr error) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "exit_code: %d\n", result.ExitCode)
	builder.WriteString("stdout:\n")
	builder.WriteString(result.Stdout)
	if !strings.HasSuffix(result.Stdout, "\n") {
		builder.WriteByte('\n')
	}
	builder.WriteString("stderr:\n")
	builder.WriteString(result.Stderr)
	if !strings.HasSuffix(result.Stderr, "\n") {
		builder.WriteByte('\n')
	}
	// A nonzero command exit already appears above. The sandbox wraps it as an
	// execution error, but showing that wrapper suggests the tool itself failed.
	if execErr != nil && result.ExitCode <= 0 {
		fmt.Fprintf(&builder, "error: %s\n", execErr)
	}
	return builder.String()
}
