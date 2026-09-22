// Package common contains the shared model/tool loop used by benchmark
// conditions.
package common

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const (
	TerminationCompleted    = "completed"
	TerminationTurnLimit    = "turn_limit"
	TerminationToolLimit    = "tool_limit"
	TerminationTokenLimit   = "token_limit"
	TerminationFinishReason = "finish_reason"
	TerminationInference    = "inference_error"
	TerminationCancellation = "cancellation"
	TerminationTimeout      = "timeout"
)

// Config controls the safety limits and generation settings for one model
// loop. The limits apply to one call to Run.
type Config struct {
	MaxTurns           int      `json:"max_turns"`
	MaxToolCalls       int      `json:"max_tool_calls"`
	ToolTimeoutSeconds float64  `json:"tool_timeout_seconds"`
	TimeoutSeconds     float64  `json:"timeout_seconds"`
	Temperature        *float64 `json:"temperature,omitempty"`
	MaxTokens          *int     `json:"max_tokens,omitempty"`
}

// DefaultConfig returns the approved development benchmark limits. Callers
// should persist the selected values with the run result.
func DefaultConfig() Config {
	return Config{
		MaxTurns:           25,
		MaxToolCalls:       50,
		ToolTimeoutSeconds: 60,
		TimeoutSeconds:     300,
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

// TokenUsage contains the aggregate token usage reported by all model
// responses in one loop run.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
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
	TokenUsage      TokenUsage          `json:"token_usage"`
	ToolCalls       []ToolCallEvidence  `json:"tool_calls"`
	Turns           int                 `json:"turns"`
	ToolCallCount   int                 `json:"tool_call_count"`
	Termination     string              `json:"termination"`
	Error           string              `json:"error,omitempty"`
	DurationSeconds float64             `json:"duration_seconds"`
}

// RunAgent guards a condition's Run preconditions (initialized receiver,
// non-blank task), calls run, and stamps condition/task onto the result.
// Shared by the benchmark conditions, e.g. baseline, prompt, and skill.
func RunAgent(initialized bool, name, condition, task string, run func() (Result, error)) (Result, error) {
	if !initialized {
		return Result{}, fmt.Errorf("%s agent is not initialized", name)
	}
	if strings.TrimSpace(task) == "" {
		return Result{}, errors.New("agent task is required")
	}
	result, err := run()
	result.Condition = condition
	result.Task = task
	return result, err
}

func (l *Loop) Run(ctx context.Context, task string) (Result, error) {
	return l.run(ctx, task, []inference.Message{{Role: "user", Content: task}})
}

func (l *Loop) RunWithSystemPrompt(ctx context.Context, task, systemPrompt string) (Result, error) {
	if strings.TrimSpace(systemPrompt) == "" {
		return Result{}, errors.New("agent system prompt is required")
	}
	return l.run(ctx, task, []inference.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: task},
	})
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
			return l.finishContext(&result, ctxErr)
		}

		requestMessages := cloneMessages(result.Messages)
		requestTools := slices.Clone(l.definitions)
		turn := result.Turns + 1
		lastMessage := requestMessages[len(requestMessages)-1]
		logger.InfoContext(runCtx, "inference message sent",
			"turn", turn,
			"message_count", len(requestMessages),
			"message_role", lastMessage.Role,
			"message_content_bytes", len(lastMessage.Content),
			"message_tool_call_count", len(lastMessage.ToolCalls),
		)
		logger.DebugContext(runCtx, "inference message sent", "turn", turn, "message_content", lastMessage.Content)
		responseStarted := time.Now()
		response, chatErr := l.client.Chat(runCtx, requestMessages, requestTools, inference.Options{
			Temperature: l.config.Temperature,
			MaxTokens:   l.config.MaxTokens,
		})
		result.Turns++
		result.Responses = append(result.Responses, ResponseEvidence{
			Response:        response,
			DurationSeconds: time.Since(responseStarted).Seconds(),
		})
		if response.Usage != nil {
			result.TokenUsage.PromptTokens += response.Usage.PromptTokens
			result.TokenUsage.CompletionTokens += response.Usage.CompletionTokens
			result.TokenUsage.TotalTokens += response.Usage.TotalTokens
		}
		if chatErr != nil {
			logger.ErrorContext(runCtx, "inference response failed",
				"turn", turn,
				"duration_seconds", time.Since(responseStarted).Seconds(),
				"error", chatErr,
			)
			if ctxErr := runCtx.Err(); ctxErr != nil {
				return l.finishContext(&result, ctxErr)
			}
			return l.finish(&result, TerminationInference, chatErr)
		}
		logInferenceResponse(logger, runCtx, turn, response, time.Since(responseStarted))
		message := response.Message
		result.Messages = append(result.Messages, message)
		if len(message.ToolCalls) == 0 {
			return l.finish(&result, terminationForFinishReason(response.FinishReason), nil)
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
			return l.finish(&result, TerminationToolLimit, nil)
		}

		for _, call := range message.ToolCalls {
			result.ToolCallCount++
			logger.InfoContext(runCtx, "tool call started",
				"turn", turn,
				"tool_call_id", call.ID,
				"tool", call.Name,
				"tool_type", call.Type,
				"arguments", call.Arguments,
			)
			toolMessage, evidence := l.executeToolCall(runCtx, call)
			result.ToolCalls = append(result.ToolCalls, evidence)
			result.Messages = append(result.Messages, toolMessage)
			logger.InfoContext(runCtx, "tool call completed",
				"turn", turn,
				"tool_call_id", call.ID,
				"tool", call.Name,
				"duration_seconds", evidence.DurationSeconds,
				"success", evidence.Error == "",
				"error", evidence.Error,
			)
			if ctxErr := runCtx.Err(); ctxErr != nil {
				return l.finishContext(&result, ctxErr)
			}
		}
	}

	return l.finish(&result, TerminationTurnLimit, nil)
}

func logInferenceResponse(logger *slog.Logger, ctx context.Context, turn int, response inference.Result, duration time.Duration) {
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

func (l *Loop) finish(result *Result, termination string, loopErr error) (Result, error) {
	result.Termination = termination
	if loopErr != nil {
		result.Error = loopErr.Error()
	}
	return *result, loopErr
}

func (l *Loop) finishContext(result *Result, ctxErr error) (Result, error) {
	termination := TerminationCancellation
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		termination = TerminationTimeout
	}
	return l.finish(result, termination, ctxErr)
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
