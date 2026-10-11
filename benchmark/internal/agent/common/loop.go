// Package common contains the shared model/tool loop used by benchmark
// conditions.
package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// Config controls the safety limits and generation settings for one model
// loop. The limits apply to one call to Run. FileTools adds read_file,
// write_file and edit_file next to bash (NewSandboxTools).
type Config struct {
	MaxTurns           int      `json:"max_turns"`
	MaxToolCalls       int      `json:"max_tool_calls"`
	ToolTimeoutSeconds float64  `json:"tool_timeout_seconds"`
	TimeoutSeconds     float64  `json:"timeout_seconds"`
	Temperature        *float64 `json:"temperature,omitempty"`
	MaxTokens          *int     `json:"max_tokens,omitempty"`
	MaxOutputBytes     int      `json:"max_output_bytes"`
	FileTools          bool     `json:"file_tools,omitempty"`
}

// DefaultConfig returns the approved development benchmark limits. Callers
// should persist the selected values with the run result.
func DefaultConfig() Config {
	return Config{
		MaxTurns:           25,
		MaxToolCalls:       50,
		ToolTimeoutSeconds: 60,
		TimeoutSeconds:     300,
		MaxOutputBytes:     8192,
		FileTools:          true,
	}
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
	if config.ToolTimeoutSeconds == 0 {
		config.ToolTimeoutSeconds = DefaultConfig().ToolTimeoutSeconds
	}
	if config.ToolTimeoutSeconds < 0 {
		return nil, errors.New("agent tool timeout must not be negative")
	}
	if config.TimeoutSeconds == 0 {
		config.TimeoutSeconds = DefaultConfig().TimeoutSeconds
	}
	if config.TimeoutSeconds < 0 {
		return nil, errors.New("agent timeout must not be negative")
	}
	if config.MaxOutputBytes < 0 {
		return nil, errors.New("agent maximum output bytes must not be negative")
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
		metadata.RuntimeSettings = maps.Clone(metadata.RuntimeSettings)
	}
	return &Loop{client: client, tools: toolMap, definitions: definitions, config: config, metadata: metadata}, nil
}

func (l *Loop) Run(ctx context.Context, task string) (Result, error) {
	return l.run(ctx, task, []inference.Message{{Role: "user", Content: task}})
}

func (l *Loop) run(ctx context.Context, task string, initialMessages []inference.Message) (result Result, err error) {
	if strings.TrimSpace(task) == "" {
		return Result{}, errors.New("agent task is required")
	}

	started := time.Now()
	defer func() { result.DurationSeconds = time.Since(started).Seconds() }()

	runCtx, cancel := context.WithTimeout(ctx, time.Duration(l.config.TimeoutSeconds*float64(time.Second)))
	defer cancel()

	result.Task = task
	result.Inference = l.metadata
	result.LoopConfig = l.config
	result.Tools = slices.Clone(l.definitions)
	result.Messages = cloneMessages(initialMessages)
	logger := slog.With("component", "agent")
	for result.Turns < l.config.MaxTurns {
		if ctxErr := runCtx.Err(); ctxErr != nil {
			return result.finishContext(ctxErr)
		}

		turn := result.Turns + 1
		response, chatErr := l.chat(runCtx, logger, turn, &result)
		if chatErr != nil {
			if ctxErr := runCtx.Err(); ctxErr != nil {
				return result.finishContext(ctxErr)
			}
			return result.finish(TerminationInference, chatErr)
		}
		message := response.Message
		result.Messages = append(result.Messages, message)
		if len(message.ToolCalls) == 0 {
			return result.finish(terminationForFinishReason(response.FinishReason), nil)
		}

		if result.ToolCallCount+len(message.ToolCalls) > l.config.MaxToolCalls {
			logger.InfoContext(runCtx, "tool calls rejected",
				"turn", turn,
				"reason", "maximum tool call limit reached",
				"requested_count", len(message.ToolCalls),
				"remaining_count", l.config.MaxToolCalls-result.ToolCallCount,
			)
			for _, call := range message.ToolCalls {
				result.ToolCalls = append(result.ToolCalls, ToolCallEvidence{
					Call:  call,
					Error: "maximum tool call limit reached",
				})
			}
			return result.finish(TerminationToolLimit, nil)
		}

		for _, call := range message.ToolCalls {
			l.callTool(runCtx, logger, turn, &result, call)
			if ctxErr := runCtx.Err(); ctxErr != nil {
				return result.finishContext(ctxErr)
			}
		}
	}

	return result.finish(TerminationTurnLimit, nil)
}

// chat sends the conversation so far and records the response, its token
// usage and any context overflow.
func (l *Loop) chat(ctx context.Context, logger *slog.Logger, turn int, result *Result) (inference.Result, error) {
	requestMessages := cloneMessages(result.Messages)
	lastMessage := requestMessages[len(requestMessages)-1]
	logger.InfoContext(ctx, "inference message sent",
		"turn", turn,
		"message_count", len(requestMessages),
		"message_role", lastMessage.Role,
		"message_content_bytes", len(lastMessage.Content),
		"message_tool_call_count", len(lastMessage.ToolCalls),
	)
	logger.DebugContext(ctx, "inference message sent", "turn", turn, "message_content", lastMessage.Content)
	started := time.Now()
	response, err := l.client.Chat(ctx, requestMessages, slices.Clone(l.definitions), inference.Options{
		Temperature: l.config.Temperature,
		MaxTokens:   l.config.MaxTokens,
	})
	result.Turns++
	result.Responses = append(result.Responses, ResponseEvidence{
		Response:        response,
		DurationSeconds: time.Since(started).Seconds(),
	})
	result.TokenUsage.add(response.Usage)
	if l.contextOverflow(response, err) {
		result.ContextOverflow = true
	}
	if err != nil {
		logger.ErrorContext(ctx, "inference response failed",
			"turn", turn,
			"duration_seconds", time.Since(started).Seconds(),
			"error", err,
		)
		return response, err
	}
	logInferenceResponse(ctx, logger, turn, response, time.Since(started))
	return response, nil
}

func (l *Loop) callTool(ctx context.Context, logger *slog.Logger, turn int, result *Result, call inference.ToolCall) {
	result.ToolCallCount++
	logger.InfoContext(ctx, "tool call started",
		"turn", turn,
		"tool_call_id", call.ID,
		"tool", call.Name,
		"tool_type", call.Type,
		"arguments", call.Arguments,
	)
	message, evidence := l.executeToolCall(ctx, call)
	result.ToolCalls = append(result.ToolCalls, evidence)
	result.Messages = append(result.Messages, message)
	logger.InfoContext(ctx, "tool call completed",
		"turn", turn,
		"tool_call_id", call.ID,
		"tool", call.Name,
		"duration_seconds", evidence.DurationSeconds,
		"success", evidence.Error == "",
		"error", evidence.Error,
	)
}

func logInferenceResponse(ctx context.Context, logger *slog.Logger, turn int, response inference.Result, duration time.Duration) {
	args := []any{
		"turn", turn,
		"response_id", response.ID,
		"response_role", response.Message.Role,
		"response_content_bytes", len(response.Message.Content),
		"response_tool_call_count", len(response.Message.ToolCalls),
		"finish_reason", response.FinishReason,
		"duration_seconds", duration.Seconds(),
	}
	if response.Usage != nil {
		args = append(args,
			"prompt_tokens", response.Usage.PromptTokens,
			"completion_tokens", response.Usage.CompletionTokens,
			"total_tokens", response.Usage.TotalTokens,
		)
	}
	if response.Timings != nil {
		args = append(args,
			"prompt_tokens_per_second", response.Timings.PromptPerSecond,
			"predicted_tokens_per_second", response.Timings.PredictedPerSecond,
			"prompt_tokens_timed", response.Timings.PromptN,
			"predicted_tokens_timed", response.Timings.PredictedN,
		)
		if response.Timings.DraftN > 0 {
			args = append(args,
				"draft_tokens", response.Timings.DraftN,
				"draft_tokens_accepted", response.Timings.DraftNAccepted,
			)
		}
	}
	logger.InfoContext(ctx, "inference response received", args...)
	logger.DebugContext(ctx, "inference response received",
		append(slices.Clone(args), "response_content", response.Message.Content)...,
	)
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

	toolCtx, cancel := context.WithTimeout(ctx, time.Duration(l.config.ToolTimeoutSeconds*float64(time.Second)))
	defer cancel()

	started := time.Now()
	toolResult := tool.Execute(toolCtx, call)
	if toolResult.Error == nil && toolCtx.Err() != nil {
		toolResult.Error = toolCtx.Err()
	}
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

func (l *Loop) contextOverflow(response inference.Result, chatErr error) bool {
	if errors.Is(chatErr, inference.ErrContextOverflow) {
		return true
	}
	if chatErr != nil || strings.TrimSpace(response.FinishReason) != "length" {
		return false
	}
	return l.config.MaxTokens == nil || (response.Usage != nil && response.Usage.CompletionTokens < *l.config.MaxTokens)
}

func terminationForFinishReason(reason string) string {
	switch strings.TrimSpace(reason) {
	case "", "stop":
		return TerminationCompleted
	case "length":
		return TerminationTokenLimit
	default:
		return TerminationFinishReason
	}
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
