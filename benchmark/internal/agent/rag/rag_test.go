package rag

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ragTestClient struct {
	messages []inference.Message
	tools    []inference.Tool
}

func (c *ragTestClient) Chat(_ context.Context, messages []inference.Message, tools []inference.Tool, _ inference.Options) (inference.Result, error) {
	c.messages = append([]inference.Message(nil), messages...)
	c.tools = append([]inference.Tool(nil), tools...)
	return inference.Result{
		Message:      inference.Message{Role: "assistant", Content: "Done."},
		FinishReason: "stop",
	}, nil
}

type ragTestShell struct{}

func (ragTestShell) Exec(context.Context, command.Spec) (command.Result, error) {
	return command.Result{}, nil
}

// wordEmbedder embeds text by whether it mentions volumes, so semantic
// ranking is predictable without a model.
type wordEmbedder struct{ err error }

func (e wordEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	if e.err != nil {
		return nil, e.err
	}
	embeddings := make([][]float32, len(texts))
	for index, text := range texts {
		embeddings[index] = []float32{0.01, 0.01}
		if strings.Contains(strings.ToLower(text), "volume") {
			embeddings[index][0] = 1
		} else {
			embeddings[index][1] = 1
		}
	}
	return embeddings, nil
}

func testSearch(t *testing.T) *Search {
	t.Helper()
	ctx := context.Background()
	path := retrieval.IndexPath(t.TempDir(), retrieval.Windows)
	documents := []retrieval.Document{
		{Path: "docs/storage.md", BlobSHA: "a", Title: "Storage", Text: "A PersistentVolumeClaim binds a volume."},
		{Path: "docs/network.md", BlobSHA: "b", Title: "Networking", Text: "NetworkPolicy isolates traffic between Pods."},
	}
	_, err := retrieval.BuildIndex(ctx, path, documents, retrieval.Windows, wordEmbedder{}, retrieval.IndexMetadata{CorpusRevision: "pinned", MaxChunkBytes: 1536, WindowOverlap: 256})
	require.NoError(t, err)
	index, err := retrieval.OpenIndex(ctx, path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = index.Close() })
	index.Hybrid = retrieval.HybridParams{Candidates: 50, RRFK: 60}
	return &Search{Index: index, Embedder: wordEmbedder{}, Mode: retrieval.Hybrid, TopK: 1, MaxBytes: 8192}
}

func TestPromptExtendsPromptCondition(t *testing.T) {
	base, err := os.ReadFile(filepath.Join("..", "prompt", "prompt.md"))
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(defaultSystemPrompt, string(base)))
	assert.Contains(t, strings.TrimPrefix(defaultSystemPrompt, string(base)), "search_docs")
}

func TestRunOffersBashAndSearch(t *testing.T) {
	client := &ragTestClient{}
	agent, err := New(client, ragTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, "", testSearch(t))
	require.NoError(t, err)

	result, err := agent.Run(context.Background(), "Restore the application.")
	require.NoError(t, err)

	assert.Equal(t, "rag", result.Condition)
	require.Len(t, client.messages, 2)
	assert.Equal(t, defaultSystemPrompt, client.messages[0].Content)
	assert.Equal(t, "Restore the application.", client.messages[1].Content)
	require.Len(t, client.tools, 2)
	assert.Equal(t, "bash", client.tools[0].Name)
	assert.Equal(t, searchToolName, client.tools[1].Name)
}

func TestNewRequiresSearch(t *testing.T) {
	_, err := New(&ragTestClient{}, ragTestShell{}, common.Config{MaxTurns: 1, MaxToolCalls: 1}, "", nil)
	assert.EqualError(t, err, "rag search index is required")
}

func TestSearchToolReturnsExcerptsAndEvidence(t *testing.T) {
	tool := searchTool{search: testSearch(t)}

	result := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"query":" volume claim "}`})

	require.NoError(t, result.Error)
	assert.True(t, strings.HasPrefix(result.Content, "[1] docs/storage.md\nStorage\n"))
	evidence, ok := result.Details.(SearchEvidence)
	require.True(t, ok)
	assert.Equal(t, "volume claim", evidence.Query)
	assert.False(t, evidence.Truncated)
	require.Len(t, evidence.Hits, 1)
	assert.Equal(t, 1, evidence.Hits[0].Rank)
	assert.Equal(t, "docs/storage.md", evidence.Hits[0].Chunk.Path)
	assert.Positive(t, evidence.Hits[0].Score)
}

func TestSearchToolCapsResult(t *testing.T) {
	search := testSearch(t)
	search.MaxBytes = 32
	result := searchTool{search: search}.Execute(context.Background(), inference.ToolCall{Arguments: `{"query":"volume"}`})

	require.NoError(t, result.Error)
	assert.LessOrEqual(t, len(result.Content), 32)
	assert.True(t, result.Details.(SearchEvidence).Truncated)
}

func TestSearchToolReportsErrors(t *testing.T) {
	search := testSearch(t)
	tool := searchTool{search: search}

	for _, arguments := range []string{`{"query":" "}`, `{"query":`} {
		result := tool.Execute(context.Background(), inference.ToolCall{Arguments: arguments})
		assert.Error(t, result.Error)
		assert.NotEmpty(t, result.Content)
	}

	search.Embedder = wordEmbedder{err: errors.New("embedding service unavailable")}
	result := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"query":"volume"}`})
	assert.ErrorContains(t, result.Error, "embedding service unavailable")
	assert.True(t, strings.HasPrefix(result.Content, "search failed:"))
	assert.Equal(t, "volume", result.Details.(SearchEvidence).Query)
}
