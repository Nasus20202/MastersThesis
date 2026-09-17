package prompt

import (
	"context"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type promptTestClient struct {
	messages []inference.Message
}

func (c *promptTestClient) Chat(_ context.Context, messages []inference.Message, _ []inference.Tool, _ inference.Options) (inference.Result, error) {
	c.messages = append([]inference.Message(nil), messages...)
	return inference.Result{
		Message:      inference.Message{Role: "assistant", Content: "Done."},
		FinishReason: "stop",
	}, nil
}

type promptTestShell struct{}

func (promptTestShell) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}

func TestRunUsesDetailedSystemPrompt(t *testing.T) {
	client := &promptTestClient{}
	agent, err := New(client, promptTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Restore the application.")
	require.NoError(t, err)

	assert.Equal(t, "prompt", result.Condition)
	assert.Equal(t, "Restore the application.", result.Task)
	require.Len(t, client.messages, 2)
	assert.Equal(t, "system", client.messages[0].Role)
	assert.Equal(t, systemPrompt, client.messages[0].Content)
	assert.Equal(t, "user", client.messages[1].Role)
	assert.Equal(t, "Restore the application.", client.messages[1].Content)
}

func TestRunRejectsInvalidAgentAndTask(t *testing.T) {
	var nilAgent *Agent
	_, err := nilAgent.Run(context.Background(), "task")
	assert.EqualError(t, err, "prompt agent is not initialized")

	agent, err := New(&promptTestClient{}, promptTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	require.NoError(t, err)
	_, err = agent.Run(context.Background(), " ")
	assert.EqualError(t, err, "agent task is required")
}
