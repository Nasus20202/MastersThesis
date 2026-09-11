package baseline

import (
	"context"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRejectsMissingDependencies(t *testing.T) {
	_, err := New(nil, nil, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	assert.EqualError(t, err, "bash shell is required")

	_, err = New(nil, &baselineTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	assert.EqualError(t, err, "agent llama client is required")
}

func TestRunRejectsUninitializedAgent(t *testing.T) {
	var nilAgent *Agent
	_, err := nilAgent.Run(context.Background(), "task")
	assert.EqualError(t, err, "baseline agent is not initialized")

	_, err = (&Agent{}).Run(context.Background(), "task")
	assert.EqualError(t, err, "baseline agent is not initialized")
}

func TestRunUsesMinimalToolPromptAndPreservesTaskEvidence(t *testing.T) {
	client := &fakeChatClient{responses: []llama.ChatResponse{{Choices: []llama.Choice{{Message: llama.Message{
		Role:    "assistant",
		Content: "finished",
	}}}}}}
	agent, err := New(client, baselineTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1})
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Restore the application.")
	require.NoError(t, err)
	assert.Equal(t, "Restore the application.", result.Task)
	require.Len(t, client.requests, 1)
	prompt := client.requests[0].Messages[0].Content
	assert.Contains(t, prompt, "Restore the application.")
	assert.Contains(t, prompt, "bash tool")
	assert.NotContains(t, prompt, "kubectl")
}

type baselineTestShell struct{}

func (baselineTestShell) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}
