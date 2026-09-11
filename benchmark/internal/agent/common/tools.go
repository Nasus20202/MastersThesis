package common

import (
	"context"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// Tool is one capability exposed to the model. Tool-specific argument
// validation and execution remain with the implementation.
type Tool interface {
	Definition() inference.Tool
	Execute(context.Context, inference.ToolCall) ToolResult
}

// ToolResult contains the model-visible output and optional implementation
// evidence from one tool execution.
type ToolResult struct {
	Content string
	Details any
	Error   error
}

// Shell is the sandbox boundary used by shell-backed tools.
type Shell interface {
	Exec(context.Context, command.Spec) (command.Result, error)
}

// ToolCallEvidence records successful, failed, malformed, and unsupported
// model tool calls.
type ToolCallEvidence struct {
	Call            inference.ToolCall `json:"call"`
	Content         string             `json:"content"`
	Details         any                `json:"details,omitempty"`
	DurationSeconds float64            `json:"duration_seconds"`
	Error           string             `json:"error,omitempty"`
}
