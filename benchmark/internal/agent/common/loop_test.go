package common

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type loopClient struct {
	results  []inference.Result
	errors   []error
	requests []loopRequest
	metadata inference.Metadata
}

type loopRequest struct {
	messages []inference.Message
	tools    []inference.Tool
	options  inference.Options
}

func (c *loopClient) Chat(_ context.Context, messages []inference.Message, tools []inference.Tool, options inference.Options) (inference.Result, error) {
	c.requests = append(c.requests, loopRequest{
		messages: cloneMessages(messages),
		tools:    append([]inference.Tool(nil), tools...),
		options:  options,
	})
	index := len(c.requests) - 1
	var result inference.Result
	if index < len(c.results) {
		result = c.results[index]
	}
	var err error
	if index < len(c.errors) {
		err = c.errors[index]
	}
	return result, err
}

func (c *loopClient) Metadata() inference.Metadata { return c.metadata }

type blockingLoopClient struct{}

func (blockingLoopClient) Chat(ctx context.Context, _ []inference.Message, _ []inference.Tool, _ inference.Options) (inference.Result, error) {
	<-ctx.Done()
	return inference.Result{}, ctx.Err()
}

type loopTool struct {
	definition inference.Tool
	result     ToolResult
	calls      []inference.ToolCall
	cancel     context.CancelFunc
}

func (t *loopTool) Definition() inference.Tool { return t.definition }

func (t *loopTool) Execute(_ context.Context, call inference.ToolCall) ToolResult {
	t.calls = append(t.calls, call)
	if t.cancel != nil {
		t.cancel()
	}
	return t.result
}

type blockingTool struct {
	definition inference.Tool
}

func (t *blockingTool) Definition() inference.Tool { return t.definition }

func (t *blockingTool) Execute(ctx context.Context, _ inference.ToolCall) ToolResult {
	<-ctx.Done()
	return ToolResult{Error: ctx.Err()}
}

func TestDefaultConfig(t *testing.T) {
	assert.Equal(t, Config{MaxTurns: 25, MaxToolCalls: 50, ToolTimeoutSeconds: 60, TimeoutSeconds: 300}, DefaultConfig())
}

func TestNewLoopValidatesDependenciesAndConfiguration(t *testing.T) {
	client := &loopClient{}
	tool := &loopTool{definition: inference.Tool{Name: "test"}}
	temperature := 0.2
	maxTokens := 10

	tests := map[string]struct {
		client inference.Client
		tools  []Tool
		config Config
		want   string
	}{
		"missing client": {
			tools:  []Tool{tool},
			config: Config{MaxTurns: 1, MaxToolCalls: 1},
			want:   "agent llama client is required",
		},
		"missing tools": {
			client: client,
			config: Config{MaxTurns: 1, MaxToolCalls: 1},
			want:   "agent requires at least one tool",
		},
		"zero turns": {
			client: client,
			tools:  []Tool{tool},
			config: Config{MaxToolCalls: 1},
			want:   "agent maximum turns must be at least 1",
		},
		"zero tool calls": {
			client: client,
			tools:  []Tool{tool},
			config: Config{MaxTurns: 1},
			want:   "agent maximum tool calls must be at least 1",
		},
		"negative temperature": {
			client: client,
			tools:  []Tool{tool},
			config: Config{MaxTurns: 1, MaxToolCalls: 1, Temperature: func() *float64 { value := -1.0; return &value }()},
			want:   "agent temperature must not be negative",
		},
		"zero max tokens": {
			client: client,
			tools:  []Tool{tool},
			config: Config{MaxTurns: 1, MaxToolCalls: 1, MaxTokens: func() *int { value := 0; return &value }()},
			want:   "agent maximum tokens must be at least 1",
		},
		"negative tool timeout": {
			client: client,
			tools:  []Tool{tool},
			config: Config{MaxTurns: 1, MaxToolCalls: 1, ToolTimeoutSeconds: -1},
			want:   "agent tool timeout must not be negative",
		},
		"negative timeout": {
			client: client,
			tools:  []Tool{tool},
			config: Config{MaxTurns: 1, MaxToolCalls: 1, TimeoutSeconds: -1},
			want:   "agent timeout must not be negative",
		},
		"nil tool": {
			client: client,
			tools:  []Tool{nil},
			config: Config{MaxTurns: 1, MaxToolCalls: 1},
			want:   "agent tool 1 is nil",
		},
		"unnamed tool": {
			client: client,
			tools:  []Tool{&loopTool{}},
			config: Config{MaxTurns: 1, MaxToolCalls: 1},
			want:   "agent tool 1 has no name",
		},
		"duplicate tool": {
			client: client,
			tools: []Tool{
				&loopTool{definition: inference.Tool{Name: "test"}},
				&loopTool{definition: inference.Tool{Name: "test"}},
			},
			config: Config{MaxTurns: 1, MaxToolCalls: 1},
			want:   `agent has duplicate tool "test"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			loop, err := NewLoop(test.client, test.tools, test.config)
			require.Nil(t, loop)
			require.EqualError(t, err, test.want)
		})
	}

	loop, err := NewLoop(client, []Tool{tool}, Config{
		MaxTurns:     2,
		MaxToolCalls: 3,
		Temperature:  &temperature,
		MaxTokens:    &maxTokens,
	})
	require.NoError(t, err)
	assert.NotNil(t, loop)
}

func TestLoopCompletesAndPassesConfiguredRequest(t *testing.T) {
	temperature := 0.4
	maxTokens := 32
	client := &loopClient{results: []inference.Result{{
		ID:           "response-1",
		Message:      inference.Message{Role: "assistant", Content: "done"},
		FinishReason: "stop",
		Usage:        &inference.Usage{TotalTokens: 7},
		Timings:      &inference.Timings{PredictedMS: 12.5},
	}}, metadata: inference.Metadata{
		Provider: "llama.cpp",
		Model:    "gemma-test",
		Artifact: "repo@revision/model.gguf",
		RuntimeSettings: map[string]string{
			"LLAMA_CONTEXT_SIZE": "32768",
		},
	}}
	tool := &loopTool{definition: inference.Tool{Name: "inspect"}}
	loop, err := NewLoop(client, []Tool{tool}, Config{
		MaxTurns:     2,
		MaxToolCalls: 2,
		Temperature:  &temperature,
		MaxTokens:    &maxTokens,
	})
	require.NoError(t, err)

	result, err := loop.Run(context.Background(), "Inspect the workload.")
	require.NoError(t, err)
	assert.Equal(t, TerminationCompleted, result.Termination)
	assert.Equal(t, "Inspect the workload.", result.Task)
	assert.Equal(t, 1, result.Turns)
	assert.Empty(t, result.ToolCalls)
	assert.Len(t, result.Messages, 2)
	assert.Equal(t, "done", result.Messages[1].Content)
	assert.Equal(t, client.results[0], result.Responses[0].Response)
	assert.Equal(t, client.metadata, result.Inference)
	assert.Equal(t, loop.config, result.LoopConfig)
	assert.Equal(t, []inference.Tool{tool.definition}, result.Tools)
	assert.GreaterOrEqual(t, result.DurationSeconds, float64(0))

	require.Len(t, client.requests, 1)
	assert.Equal(t, result.Messages[:1], client.requests[0].messages)
	assert.Equal(t, []inference.Tool{tool.definition}, client.requests[0].tools)
	assert.Equal(t, &temperature, client.requests[0].options.Temperature)
	assert.Equal(t, &maxTokens, client.requests[0].options.MaxTokens)
}

func TestLoopPreservesConfiguredSystemPromptInTranscriptAndRequest(t *testing.T) {
	client := &loopClient{results: []inference.Result{{
		Message:      inference.Message{Role: "assistant", Content: "done"},
		FinishReason: "stop",
	}}}
	tool := &loopTool{definition: inference.Tool{Name: "inspect"}}
	loop, err := NewLoopWithSystemPrompt(client, []Tool{tool}, Config{MaxTurns: 1, MaxToolCalls: 1}, "short system context")
	require.NoError(t, err)

	result, err := loop.Run(context.Background(), "Inspect the environment.")
	require.NoError(t, err)
	require.Len(t, result.Messages, 3)
	assert.Equal(t, inference.Message{Role: "system", Content: "short system context"}, result.Messages[0])
	assert.Equal(t, inference.Message{Role: "user", Content: "Inspect the environment."}, result.Messages[1])
	assert.Equal(t, result.Messages[:2], client.requests[0].messages)
}

func TestLoopLogsInferenceMessagesResponsesToolCallsAndMetrics(t *testing.T) {
	var logs bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	client := &loopClient{results: []inference.Result{
		{
			ID: "response-1",
			Message: inference.Message{
				Role:    "assistant",
				Content: "I will inspect the workload.",
				ToolCalls: []inference.ToolCall{{
					ID: "call-1", Type: "function", Name: "inspect", Arguments: `{}`,
				}},
			},
			FinishReason: "tool_calls",
			Usage:        &inference.Usage{PromptTokens: 10, CompletionTokens: 4, TotalTokens: 14},
			Timings:      &inference.Timings{PromptPerSecond: 20.5, PredictedPerSecond: 30.5},
		},
		{ID: "response-2", Message: inference.Message{Role: "assistant", Content: "The workload is healthy."}, FinishReason: "stop"},
	}}
	tool := &loopTool{definition: inference.Tool{Name: "inspect"}, result: ToolResult{Content: "healthy"}}
	loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 2, MaxToolCalls: 1})
	require.NoError(t, err)

	_, err = loop.Run(context.Background(), "Inspect the workload.")
	require.NoError(t, err)

	records := make([]map[string]any, 0)
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		var record map[string]any
		require.NoError(t, json.Unmarshal([]byte(line), &record))
		records = append(records, record)
	}
	require.Len(t, records, 10)

	sent := make([]map[string]any, 0, 2)
	responses := make([]map[string]any, 0, 2)
	debugSent := make([]map[string]any, 0, 2)
	debugResponses := make([]map[string]any, 0, 2)
	byMessage := make(map[string]map[string]any, len(records))
	for _, record := range records {
		message := record["msg"].(string)
		if record["level"] == "DEBUG" {
			switch message {
			case "inference message sent":
				debugSent = append(debugSent, record)
			case "inference response received":
				debugResponses = append(debugResponses, record)
			}
			continue
		}
		switch message {
		case "inference message sent":
			sent = append(sent, record)
		case "inference response received":
			responses = append(responses, record)
		default:
			byMessage[message] = record
		}
	}
	require.Len(t, sent, 2)
	require.Len(t, responses, 2)
	require.Len(t, debugSent, 2)
	require.Len(t, debugResponses, 2)
	assert.Equal(t, float64(len("Inspect the workload.")), sent[0]["message_content_bytes"])
	assert.NotContains(t, sent[0], "message_content")
	assert.Equal(t, "Inspect the workload.", debugSent[0]["message_content"])
	assert.Equal(t, float64(len("The workload is healthy.")), responses[1]["response_content_bytes"])
	assert.NotContains(t, responses[1], "response_content")
	assert.Equal(t, "The workload is healthy.", debugResponses[1]["response_content"])
	assert.Equal(t, float64(20.5), responses[0]["prompt_tokens_per_second"])
	assert.Equal(t, float64(30.5), responses[0]["predicted_tokens_per_second"])
	assert.Equal(t, "inspect", byMessage["tool call started"]["tool"])
	assert.Equal(t, "{}", byMessage["tool call started"]["arguments"])
	assert.Equal(t, true, byMessage["tool call completed"]["success"])
}

func TestLoopExecutesToolCallsAndAppendsEvidence(t *testing.T) {
	client := &loopClient{results: []inference.Result{
		{Message: inference.Message{
			Role: "assistant",
			ToolCalls: []inference.ToolCall{{
				ID: "call-1", Type: "function", Name: "inspect", Arguments: `{}`,
			}},
		}},
		{Message: inference.Message{Role: "assistant", Content: "finished"}},
	}}
	tool := &loopTool{
		definition: inference.Tool{Name: "inspect"},
		result: ToolResult{
			Content: "workload is healthy",
			Details: map[string]string{"source": "sandbox"},
		},
	}
	loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 2, MaxToolCalls: 2})
	require.NoError(t, err)

	result, err := loop.Run(context.Background(), "Inspect the workload.")
	require.NoError(t, err)
	assert.Equal(t, TerminationCompleted, result.Termination)
	assert.Equal(t, 2, result.Turns)
	assert.Equal(t, 1, result.ToolCallCount)
	assert.Equal(t, client.results[1].Message, result.Messages[3])
	assert.Equal(t, "tool", result.Messages[2].Role)
	assert.Equal(t, "call-1", result.Messages[2].ToolCallID)
	assert.Equal(t, "workload is healthy", result.Messages[2].Content)
	require.Len(t, result.ToolCalls, 1)
	assert.Equal(t, client.results[0].Message.ToolCalls[0], result.ToolCalls[0].Call)
	assert.Equal(t, "workload is healthy", result.ToolCalls[0].Content)
	assert.Equal(t, tool.result.Details, result.ToolCalls[0].Details)
	assert.Empty(t, result.ToolCalls[0].Error)
	assert.Equal(t, client.results[0].Message.ToolCalls, tool.calls)
	assert.Equal(t, result.Messages[:1], client.requests[0].messages)
	assert.Equal(t, result.Messages[:3], client.requests[1].messages)
}

func TestLoopHandlesUnsupportedCallsAndToolErrors(t *testing.T) {
	client := &loopClient{results: []inference.Result{
		{Message: inference.Message{Role: "assistant", ToolCalls: []inference.ToolCall{
			{ID: "unknown", Type: "function", Name: "missing"},
			{ID: "wrong-type", Type: "custom", Name: "inspect"},
			{ID: "failed", Type: "function", Name: "inspect"},
		}}},
		{Message: inference.Message{Role: "assistant", Content: "finished"}},
	}}
	tool := &loopTool{
		definition: inference.Tool{Name: "inspect"},
		result:     ToolResult{Error: errors.New("tool failed")},
	}
	loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 2, MaxToolCalls: 3})
	require.NoError(t, err)

	result, err := loop.Run(context.Background(), "Inspect the workload.")
	require.NoError(t, err)
	assert.Equal(t, TerminationCompleted, result.Termination)
	assert.Len(t, result.ToolCalls, 3)
	assert.Contains(t, result.ToolCalls[0].Error, `unsupported tool "missing"`)
	assert.Contains(t, result.ToolCalls[1].Error, `unsupported tool call type "custom"`)
	assert.Empty(t, result.ToolCalls[2].Content)
	assert.Equal(t, "tool failed", result.ToolCalls[2].Error)
	assert.Equal(t, "tool failed", result.Messages[4].Content)
	assert.Equal(t, []inference.ToolCall{client.results[0].Message.ToolCalls[2]}, tool.calls)
}

func TestLoopStopsAtTurnAndToolLimits(t *testing.T) {
	toolResponse := inference.Result{Message: inference.Message{Role: "assistant", ToolCalls: []inference.ToolCall{
		{ID: "one", Type: "function", Name: "inspect"},
		{ID: "two", Type: "function", Name: "inspect"},
	}}}

	t.Run("turn limit", func(t *testing.T) {
		client := &loopClient{results: []inference.Result{toolResponse}}
		tool := &loopTool{definition: inference.Tool{Name: "inspect"}}
		loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 1, MaxToolCalls: 2})
		require.NoError(t, err)

		result, err := loop.Run(context.Background(), "Inspect the workload.")
		require.NoError(t, err)
		assert.Equal(t, TerminationTurnLimit, result.Termination)
		assert.Equal(t, 2, result.ToolCallCount)
		assert.Len(t, tool.calls, 2)
	})

	t.Run("tool limit records pending calls", func(t *testing.T) {
		client := &loopClient{results: []inference.Result{toolResponse}}
		tool := &loopTool{definition: inference.Tool{Name: "inspect"}}
		loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 2, MaxToolCalls: 1})
		require.NoError(t, err)

		result, err := loop.Run(context.Background(), "Inspect the workload.")
		require.NoError(t, err)
		assert.Equal(t, TerminationToolLimit, result.Termination)
		assert.Equal(t, 0, result.ToolCallCount)
		assert.Len(t, result.ToolCalls, 2)
		assert.Empty(t, tool.calls)
		assert.Equal(t, "maximum tool call limit reached", result.ToolCalls[0].Error)
	})
}

func TestLoopClassifiesFinishReasonsAndTimeout(t *testing.T) {
	tool := &loopTool{definition: inference.Tool{Name: "inspect"}}

	for reason, want := range map[string]string{
		"length":         TerminationTokenLimit,
		"content_filter": TerminationFinishReason,
	} {
		t.Run(reason, func(t *testing.T) {
			client := &loopClient{results: []inference.Result{{
				Message:      inference.Message{Role: "assistant", Content: "partial"},
				FinishReason: reason,
			}}}
			loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 1, MaxToolCalls: 1})
			require.NoError(t, err)

			result, err := loop.Run(context.Background(), "Inspect the workload.")
			require.NoError(t, err)
			assert.Equal(t, want, result.Termination)
		})
	}

	loop, err := NewLoop(blockingLoopClient{}, []Tool{tool}, Config{
		MaxTurns:       1,
		MaxToolCalls:   1,
		TimeoutSeconds: 0.01,
	})
	require.NoError(t, err)

	started := time.Now()
	result, err := loop.Run(context.Background(), "Inspect the workload.")
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, TerminationTimeout, result.Termination)
	assert.Less(t, time.Since(started), time.Second)
}

func TestLoopReturnsInferenceErrorsAndCancellation(t *testing.T) {
	t.Run("inference error", func(t *testing.T) {
		client := &loopClient{errors: []error{errors.New("server unavailable")}}
		tool := &loopTool{definition: inference.Tool{Name: "inspect"}}
		loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 1, MaxToolCalls: 1})
		require.NoError(t, err)

		result, err := loop.Run(context.Background(), "Inspect the workload.")
		assert.EqualError(t, err, "server unavailable")
		assert.Equal(t, TerminationInference, result.Termination)
		assert.Equal(t, "server unavailable", result.Error)
		assert.Len(t, result.Responses, 1)
	})

	t.Run("already cancelled", func(t *testing.T) {
		client := &loopClient{}
		tool := &loopTool{definition: inference.Tool{Name: "inspect"}}
		loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 1, MaxToolCalls: 1})
		require.NoError(t, err)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := loop.Run(ctx, "Inspect the workload.")
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, TerminationCancellation, result.Termination)
		assert.Empty(t, client.requests)
	})

	t.Run("cancelled by tool", func(t *testing.T) {
		client := &loopClient{results: []inference.Result{{Message: inference.Message{Role: "assistant", ToolCalls: []inference.ToolCall{{
			ID: "call-1", Type: "function", Name: "inspect",
		}}}}}}
		ctx, cancel := context.WithCancel(context.Background())
		tool := &loopTool{definition: inference.Tool{Name: "inspect"}, cancel: cancel}
		loop, err := NewLoop(client, []Tool{tool}, Config{MaxTurns: 1, MaxToolCalls: 1})
		require.NoError(t, err)

		result, err := loop.Run(ctx, "Inspect the workload.")
		assert.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, TerminationCancellation, result.Termination)
		assert.Equal(t, context.Canceled.Error(), result.Error)
	})
}

func TestLoopRejectsEmptyTasks(t *testing.T) {
	tool := &loopTool{definition: inference.Tool{Name: "inspect"}}
	loop, err := NewLoop(&loopClient{}, []Tool{tool}, Config{MaxTurns: 1, MaxToolCalls: 1})
	require.NoError(t, err)

	result, err := loop.Run(context.Background(), " ")
	assert.EqualError(t, err, "agent task is required")
	assert.Empty(t, result)
}

func TestLoopTimesOutIndividualToolCallAndContinues(t *testing.T) {
	client := &loopClient{results: []inference.Result{
		{Message: inference.Message{Role: "assistant", ToolCalls: []inference.ToolCall{{
			ID: "call-1", Type: "function", Name: "inspect",
		}}}},
		{Message: inference.Message{Role: "assistant", Content: "finished"}, FinishReason: "stop"},
	}}
	tool := &blockingTool{definition: inference.Tool{Name: "inspect"}}
	loop, err := NewLoop(client, []Tool{tool}, Config{
		MaxTurns:           2,
		MaxToolCalls:       1,
		ToolTimeoutSeconds: 0.01,
		TimeoutSeconds:     1,
	})
	require.NoError(t, err)

	started := time.Now()
	result, err := loop.Run(context.Background(), "Inspect the workload.")
	require.NoError(t, err)
	assert.Equal(t, TerminationCompleted, result.Termination)
	assert.Equal(t, 1, result.ToolCallCount)
	require.Len(t, result.ToolCalls, 1)
	assert.ErrorContains(t, errors.New(result.ToolCalls[0].Error), context.DeadlineExceeded.Error())
	assert.Less(t, time.Since(started), time.Second)
}
