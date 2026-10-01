package evaluation

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

func TestParseQueriesStripsListMarkers(t *testing.T) {
	queries, err := parseQueries("1. pod pending memory request\n\n- \"node allocatable\"\n* `kubectl describe pod events`\n4) extra")
	require.NoError(t, err)
	assert.Equal(t, []string{"pod pending memory request", "node allocatable", "kubectl describe pod events"}, queries)

	_, err = parseQueries("only one\n")
	assert.ErrorContains(t, err, "returned 1 queries")
}

type recordingClient struct {
	messages []inference.Message
	options  inference.Options
}

func (c *recordingClient) Chat(_ context.Context, messages []inference.Message, _ []inference.Tool, options inference.Options) (inference.Result, error) {
	c.messages, c.options = messages, options
	return inference.Result{Message: inference.Message{Content: "a\nb\nc", ReasoningContent: "thinking"}}, nil
}

func TestGenerateQueriesSendsOnlyTheProbePrompt(t *testing.T) {
	client := &recordingClient{}
	probe := Probe{ID: "K01", Prompt: "Why Pending?", Primary: []string{"content/en/docs/a.md"}}

	generated, err := GenerateQueries(context.Background(), client, []Probe{probe}, 0, 42)
	require.NoError(t, err)
	assert.Equal(t, []ProbeQueries{{Probe: probe, Queries: []string{"a", "b", "c"}, Response: "a\nb\nc", Reasoning: "thinking"}}, generated)
	require.Len(t, client.messages, 1)
	assert.Contains(t, client.messages[0].Content, "Problem:\nWhy Pending?")
	assert.NotContains(t, client.messages[0].Content, "content/en/docs")
	assert.Equal(t, 0.0, *client.options.Temperature)
	assert.Equal(t, 42, *client.options.Seed)
}

func TestQuerySetIsFrozenOnceWritten(t *testing.T) {
	path := filepath.Join(t.TempDir(), "study", "queries.json")
	set := QuerySet{Seed: 42, Probes: []ProbeQueries{{Probe: Probe{ID: "K01"}, Queries: []string{"a"}}}}

	require.NoError(t, WriteQuerySet(path, set))
	loaded, digest, err := LoadQuerySet(path)
	require.NoError(t, err)
	assert.Equal(t, set, loaded)
	assert.Len(t, digest, 64)
	assert.ErrorContains(t, WriteQuerySet(path, set), "frozen")
}
