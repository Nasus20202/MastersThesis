package baseline

import (
	"context"
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChatClient struct {
	responses []llama.ChatResponse
	requests  []fakeRequest
	index     int
}

type fakeRequest struct {
	Messages []inference.Message
	Tools    []inference.Tool
	Options  inference.Options
}

func (c *fakeChatClient) Chat(_ context.Context, messages []inference.Message, tools []inference.Tool, options inference.Options) (inference.Result, error) {
	c.requests = append(c.requests, fakeRequest{Messages: messages, Tools: tools, Options: options})
	response := c.responses[c.index]
	c.index++
	choice := response.Choices[0]
	return inference.Result{
		ID:           response.ID,
		Message:      genericMessage(choice.Message),
		FinishReason: choice.FinishReason,
	}, nil
}

func genericMessage(message llama.Message) inference.Message {
	calls := make([]inference.ToolCall, len(message.ToolCalls))
	for index, call := range message.ToolCalls {
		calls[index] = inference.ToolCall{
			ID:        call.ID,
			Type:      call.Type,
			Name:      call.Function.Name,
			Arguments: call.Function.Arguments,
		}
	}
	return inference.Message{
		Role:       message.Role,
		Content:    message.Content,
		Name:       message.Name,
		ToolCallID: message.ToolCallID,
		ToolCalls:  calls,
	}
}

type fakeShell struct {
	specs   []command.Spec
	results []command.Result
	errors  []error
}

func (s *fakeShell) Exec(_ context.Context, spec command.Spec) (command.Result, error) {
	s.specs = append(s.specs, spec)
	index := len(s.specs) - 1
	var result command.Result
	if index < len(s.results) {
		result = s.results[index]
	}
	var err error
	if index < len(s.errors) {
		err = s.errors[index]
	}
	return result, err
}

func TestLoopExecutesBashCallsAndContinuesWithToolOutput(t *testing.T) {
	client := &fakeChatClient{responses: []llama.ChatResponse{
		{Choices: []llama.Choice{{Message: llama.Message{
			Role:    "assistant",
			Content: "I will inspect the workload.",
			ToolCalls: []llama.ToolCall{{
				ID:   "call-1",
				Type: "function",
				Function: llama.ToolCallFunction{
					Name:      "bash",
					Arguments: `{"command":"printf healthy"}`,
				},
			}},
		}}}},
		{Choices: []llama.Choice{{Message: llama.Message{
			Role:    "assistant",
			Content: "The workload is healthy.",
		}}}},
	}}
	shell := &fakeShell{results: []command.Result{{
		Stdout:   "healthy",
		ExitCode: 0,
		Duration: 5 * time.Millisecond,
	}}}
	loop, err := New(client, shell, common.Config{MaxTurns: 3, MaxToolCalls: 2})
	require.NoError(t, err)

	result, err := loop.Run(context.Background(), "Restore the application.")
	require.NoError(t, err)
	assert.Equal(t, common.TerminationCompleted, result.Termination)
	assert.Equal(t, 2, result.Turns)
	assert.Equal(t, 1, result.ToolCallCount)
	assert.Equal(t, "Restore the application.", result.Task)
	assert.Len(t, result.Messages, 4)
	assert.Equal(t, "tool", result.Messages[2].Role)
	assert.Contains(t, result.Messages[2].Content, "healthy")
	assert.Equal(t, "The workload is healthy.", result.Messages[3].Content)

	require.Len(t, shell.specs, 1)
	assert.Equal(t, command.Spec{Program: "bash", Args: []string{"-lc", "printf healthy"}}, shell.specs[0])
	require.Len(t, result.ToolCalls, 1)
	details, ok := result.ToolCalls[0].Details.(common.CommandEvidence)
	require.True(t, ok)
	assert.Equal(t, "printf healthy", details.Command)
	assert.Equal(t, "healthy", details.Stdout)
	assert.Equal(t, 0.005, details.DurationSeconds)

	require.Len(t, client.requests, 2)
	assert.Equal(t, result.Messages[:1], client.requests[0].Messages)
	assert.Equal(t, result.Messages[:3], client.requests[1].Messages)
	require.Len(t, client.requests[0].Tools, 1)
	assert.Equal(t, common.BashTool(), client.requests[0].Tools[0])
}

func TestLoopReturnsToolErrorsToModelWithoutCallingShell(t *testing.T) {
	client := &fakeChatClient{responses: []llama.ChatResponse{
		{Choices: []llama.Choice{{Message: llama.Message{
			Role: "assistant",
			ToolCalls: []llama.ToolCall{
				{ID: "unsupported", Type: "function", Function: llama.ToolCallFunction{Name: "kubectl", Arguments: `{}`}},
				{ID: "malformed", Type: "function", Function: llama.ToolCallFunction{Name: "bash", Arguments: `{`}},
				{ID: "wrong-type", Type: "custom", Function: llama.ToolCallFunction{Name: "bash", Arguments: `{}`}},
			},
		}}}},
		{Choices: []llama.Choice{{Message: llama.Message{Role: "assistant", Content: "Finished."}}}},
	}}
	shell := &fakeShell{}
	loop, err := New(client, shell, common.Config{MaxTurns: 2, MaxToolCalls: 3})
	require.NoError(t, err)

	result, err := loop.Run(context.Background(), "Inspect the workload.")
	require.NoError(t, err)
	assert.Equal(t, common.TerminationCompleted, result.Termination)
	assert.Empty(t, shell.specs)
	assert.Len(t, result.ToolCalls, 3)
	assert.Contains(t, result.ToolCalls[0].Error, `unsupported tool "kubectl"`)
	assert.Contains(t, result.ToolCalls[1].Error, "malformed bash arguments")
	assert.Contains(t, result.ToolCalls[2].Error, `unsupported tool call type "custom"`)
	assert.Equal(t, 3, len(client.requests[1].Messages)-2)
	for _, message := range client.requests[1].Messages[2:] {
		assert.Equal(t, "tool", message.Role)
		assert.NotEmpty(t, message.Content)
	}
}

func TestLoopStopsAtConfiguredLimits(t *testing.T) {
	toolResponse := llama.ChatResponse{Choices: []llama.Choice{{Message: llama.Message{
		Role: "assistant",
		ToolCalls: []llama.ToolCall{{
			ID: "call-1", Type: "function",
			Function: llama.ToolCallFunction{Name: "bash", Arguments: `{"command":"true"}`},
		}},
	}}}}

	t.Run("turn limit", func(t *testing.T) {
		client := &fakeChatClient{responses: []llama.ChatResponse{toolResponse}}
		shell := &fakeShell{results: []command.Result{{ExitCode: 0}}}
		loop, err := New(client, shell, common.Config{MaxTurns: 1, MaxToolCalls: 2})
		require.NoError(t, err)

		result, err := loop.Run(context.Background(), "Repair the workload.")
		require.NoError(t, err)
		assert.Equal(t, common.TerminationTurnLimit, result.Termination)
		assert.Len(t, shell.specs, 1)
	})

	t.Run("tool limit", func(t *testing.T) {
		client := &fakeChatClient{responses: []llama.ChatResponse{toolResponse}}
		shell := &fakeShell{}
		loop, err := New(client, shell, common.Config{MaxTurns: 2, MaxToolCalls: 0})
		assert.Error(t, err)
		assert.Nil(t, loop)

		loop, err = New(client, shell, common.Config{MaxTurns: 1, MaxToolCalls: 1})
		require.NoError(t, err)
		result, err := loop.Run(context.Background(), "Repair the workload.")
		require.NoError(t, err)
		assert.Equal(t, common.TerminationTurnLimit, result.Termination)

		client = &fakeChatClient{responses: []llama.ChatResponse{{Choices: []llama.Choice{{Message: llama.Message{
			Role: "assistant",
			ToolCalls: []llama.ToolCall{
				{ID: "one", Type: "function", Function: llama.ToolCallFunction{Name: "bash", Arguments: `{"command":"true"}`}},
				{ID: "two", Type: "function", Function: llama.ToolCallFunction{Name: "bash", Arguments: `{"command":"true"}`}},
			},
		}}}}}}
		shell = &fakeShell{}
		loop, err = New(client, shell, common.Config{MaxTurns: 2, MaxToolCalls: 1})
		require.NoError(t, err)
		result, err = loop.Run(context.Background(), "Repair the workload.")
		require.NoError(t, err)
		assert.Equal(t, common.TerminationToolLimit, result.Termination)
		assert.Empty(t, shell.specs)
	})
}

func TestLoopHonorsCancellation(t *testing.T) {
	client := &fakeChatClient{}
	shell := &fakeShell{}
	loop, err := New(client, shell, common.DefaultConfig())
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := loop.Run(ctx, "Inspect the workload.")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, common.TerminationCancellation, result.Termination)
	assert.Equal(t, context.Canceled.Error(), result.Error)
	assert.Empty(t, client.requests)
}

func TestNewLoopValidatesConfiguration(t *testing.T) {
	client := &fakeChatClient{}
	shell := &fakeShell{}
	negative := -1.0
	zero := 0
	for name, config := range map[string]common.Config{
		"missing turns":   {MaxToolCalls: 1},
		"missing tools":   {MaxTurns: 1},
		"negative temp":   {MaxTurns: 1, MaxToolCalls: 1, Temperature: &negative},
		"zero max tokens": {MaxTurns: 1, MaxToolCalls: 1, MaxTokens: &zero},
	} {
		t.Run(name, func(t *testing.T) {
			loop, err := New(client, shell, config)
			assert.Error(t, err)
			assert.Nil(t, loop)
		})
	}
}
