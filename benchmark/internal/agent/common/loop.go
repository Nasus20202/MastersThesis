// Package common contains the shared model/tool loop used by benchmark
// conditions.
package common

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const (
	TerminationCompleted    = "completed"
	TerminationTurnLimit    = "turn_limit"
	TerminationToolLimit    = "tool_limit"
	TerminationInference    = "inference_error"
	TerminationCancellation = "cancellation"
)

// Config controls the safety limits and generation settings for one model
// loop. The limits apply to one call to Run.
type Config struct {
	MaxTurns     int      `json:"max_turns"`
	MaxToolCalls int      `json:"max_tool_calls"`
	Temperature  *float64 `json:"temperature,omitempty"`
	MaxTokens    *int     `json:"max_tokens,omitempty"`
}

// DefaultConfig returns conservative limits suitable for a benchmark smoke
// test. Callers should persist the selected values with the run result.
func DefaultConfig() Config {
	return Config{MaxTurns: 12, MaxToolCalls: 24}
}

type Loop struct {
	client      inference.Client
	tools       map[string]Tool
	definitions []inference.Tool
	config      Config
	metadata    inference.Metadata
}

func NewLoop(client inference.Client, tools []Tool, config Config) (*Loop, error) {
	if client == nil {
		return nil, errors.New("agent llama client is required")
	}
	if len(tools) == 0 {
		return nil, errors.New("agent requires at least one tool")
	}
	if config.MaxTurns < 1 {
		return nil, errors.New("agent maximum turns must be at least 1")
	}
	if config.MaxToolCalls < 1 {
		return nil, errors.New("agent maximum tool calls must be at least 1")
	}
	if config.Temperature != nil && *config.Temperature < 0 {
		return nil, errors.New("agent temperature must not be negative")
	}
	if config.MaxTokens != nil && *config.MaxTokens < 1 {
		return nil, errors.New("agent maximum tokens must be at least 1")
	}

	toolMap := make(map[string]Tool, len(tools))
	definitions := make([]inference.Tool, 0, len(tools))
	for index, tool := range tools {
		if tool == nil {
			return nil, fmt.Errorf("agent tool %d is nil", index+1)
		}
		definition := tool.Definition()
		name := strings.TrimSpace(definition.Name)
		if name == "" {
			return nil, fmt.Errorf("agent tool %d has no name", index+1)
		}
		if _, exists := toolMap[name]; exists {
			return nil, fmt.Errorf("agent has duplicate tool %q", name)
		}
		toolMap[name] = tool
		definitions = append(definitions, definition)
	}

	metadata := inference.Metadata{}
	if provider, ok := client.(inference.MetadataProvider); ok {
		metadata = provider.Metadata()
		metadata.RuntimeSettings = cloneRuntimeSettings(metadata.RuntimeSettings)
	}
	return &Loop{client: client, tools: toolMap, definitions: definitions, config: config, metadata: metadata}, nil
}

// ResponseEvidence preserves the raw response and the client round-trip
// duration for later result persistence.
type ResponseEvidence struct {
	Response        inference.Result `json:"response"`
	DurationSeconds float64          `json:"duration_seconds"`
}

// Result is the complete model-loop evidence produced by Run.
type Result struct {
	Condition       string              `json:"condition,omitempty"`
	Task            string              `json:"task"`
	Inference       inference.Metadata  `json:"inference"`
	LoopConfig      Config              `json:"loop_config"`
	Tools           []inference.Tool    `json:"tools"`
	Messages        []inference.Message `json:"messages"`
	Responses       []ResponseEvidence  `json:"responses"`
	ToolCalls       []ToolCallEvidence  `json:"tool_calls"`
	Turns           int                 `json:"turns"`
	ToolCallCount   int                 `json:"tool_call_count"`
	Termination     string              `json:"termination"`
	Error           string              `json:"error,omitempty"`
	DurationSeconds float64             `json:"duration_seconds"`
}

func (l *Loop) Run(ctx context.Context, task string) (result Result, err error) {
	if strings.TrimSpace(task) == "" {
		return Result{}, errors.New("agent task is required")
	}

	started := time.Now()
	defer func() { result.DurationSeconds = time.Since(started).Seconds() }()

	result.Task = task
	result.Inference = l.metadata
	result.LoopConfig = l.config
	result.Tools = slices.Clone(l.definitions)
	result.Messages = []inference.Message{{Role: "user", Content: task}}
	for result.Turns < l.config.MaxTurns {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return l.finish(&result, TerminationCancellation, ctxErr)
		}

		requestMessages := cloneMessages(result.Messages)
		requestTools := slices.Clone(l.definitions)
		responseStarted := time.Now()
		response, chatErr := l.client.Chat(ctx, requestMessages, requestTools, inference.Options{
			Temperature: l.config.Temperature,
			MaxTokens:   l.config.MaxTokens,
		})
		result.Turns++
		result.Responses = append(result.Responses, ResponseEvidence{
			Response:        response,
			DurationSeconds: time.Since(responseStarted).Seconds(),
		})
		if chatErr != nil {
			if ctxErr := ctx.Err(); ctxErr != nil {
				return l.finish(&result, TerminationCancellation, ctxErr)
			}
			return l.finish(&result, TerminationInference, chatErr)
		}
		message := response.Message
		result.Messages = append(result.Messages, message)
		if len(message.ToolCalls) == 0 {
			result.Termination = TerminationCompleted
			return result, nil
		}

		if result.ToolCallCount+len(message.ToolCalls) > l.config.MaxToolCalls {
			for _, call := range message.ToolCalls {
				result.ToolCalls = append(result.ToolCalls, ToolCallEvidence{
					Call:  call,
					Error: "maximum tool call limit reached",
				})
			}
			return l.finish(&result, TerminationToolLimit, nil)
		}

		for _, call := range message.ToolCalls {
			result.ToolCallCount++
			toolMessage, evidence := l.executeToolCall(ctx, call)
			result.ToolCalls = append(result.ToolCalls, evidence)
			result.Messages = append(result.Messages, toolMessage)
			if ctxErr := ctx.Err(); ctxErr != nil {
				return l.finish(&result, TerminationCancellation, ctxErr)
			}
		}
	}

	return l.finish(&result, TerminationTurnLimit, nil)
}

func (l *Loop) executeToolCall(ctx context.Context, call inference.ToolCall) (inference.Message, ToolCallEvidence) {
	evidence := ToolCallEvidence{Call: call}
	if call.Type != "function" {
		evidence.Error = fmt.Sprintf("unsupported tool call type %q", call.Type)
		return toolMessage(call.ID, evidence.Error), evidence
	}

	tool, ok := l.tools[call.Name]
	if !ok {
		evidence.Error = fmt.Sprintf("unsupported tool %q", call.Name)
		return toolMessage(call.ID, evidence.Error), evidence
	}

	started := time.Now()
	toolResult := tool.Execute(ctx, call)
	evidence.DurationSeconds = time.Since(started).Seconds()
	evidence.Content = toolResult.Content
	evidence.Details = toolResult.Details
	if toolResult.Error != nil {
		evidence.Error = toolResult.Error.Error()
	}
	content := toolResult.Content
	if content == "" && toolResult.Error != nil {
		content = toolResult.Error.Error()
	}
	return toolMessage(call.ID, content), evidence
}

func (l *Loop) finish(result *Result, termination string, loopErr error) (Result, error) {
	result.Termination = termination
	if loopErr != nil {
		result.Error = loopErr.Error()
	}
	return *result, loopErr
}

func toolMessage(callID, content string) inference.Message {
	return inference.Message{Role: "tool", ToolCallID: callID, Content: content}
}

func cloneMessages(messages []inference.Message) []inference.Message {
	cloned := make([]inference.Message, len(messages))
	copy(cloned, messages)
	for index := range cloned {
		cloned[index].ToolCalls = slices.Clone(cloned[index].ToolCalls)
	}
	return cloned
}

func cloneRuntimeSettings(settings map[string]string) map[string]string {
	if settings == nil {
		return nil
	}
	cloned := make(map[string]string, len(settings))
	for key, value := range settings {
		cloned[key] = value
	}
	return cloned
}
