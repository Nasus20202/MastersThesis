package retrieval

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// topicEmbedder embeds text by which topic words it contains, so semantic
// search is predictable without a model.
type topicEmbedder struct {
	calls int
}

var topics = []string{"volume", "network", "schedul"}

func (e *topicEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	e.calls++
	embeddings := make([][]float32, len(texts))
	for index, text := range texts {
		vector := []float32{0.01, 0.01, 0.01}
		for topic, word := range topics {
			if strings.Contains(strings.ToLower(text), word) {
				vector[topic] = 1
			}
		}
		embeddings[index] = vector
	}
	return embeddings, nil
}

var testDocuments = []Document{
	{Path: "docs/storage.md", BlobSHA: "a", Title: "Storage", Text: "## Claims\n\nA PersistentVolumeClaim binds a volume."},
	{Path: "docs/network.md", BlobSHA: "b", Title: "Networking", Text: "## Policies\n\nNetworkPolicy isolates network traffic between Pods."},
	{Path: "docs/scheduler.md", BlobSHA: "c", Title: "Scheduling", Text: "## Taints\n\nThe scheduler avoids nodes with taints."},
}

func buildTestIndex(t *testing.T, embedder *topicEmbedder) *Index {
	t.Helper()
	path := IndexPath(t.TempDir(), Sections)
	metadata, err := BuildIndex(context.Background(), path, testDocuments, Sections, embedder, IndexMetadata{CorpusRevision: "rev"})
	require.NoError(t, err)
	assert.Equal(t, IndexMetadata{CorpusRevision: "rev", Chunking: Sections, MaxChunkBytes: MaxChunkBytes, Documents: 3, Chunks: 3, Dimensions: 3}, metadata)
	assert.NoFileExists(t, path+".partial")

	index, err := OpenIndex(context.Background(), path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = index.Close() })
	assert.Equal(t, metadata, index.Metadata())
	return index
}

func TestSearchModesRankTheMatchingChunkFirst(t *testing.T) {
	embedder := &topicEmbedder{}
	index := buildTestIndex(t, embedder)
	ctx := context.Background()

	lexical, err := index.Search(ctx, nil, Lexical, `taints "OR" nodes*`, 5)
	require.NoError(t, err)
	require.Len(t, lexical, 1)
	assert.Equal(t, "docs/scheduler.md", lexical[0].Chunk.Path)
	assert.Equal(t, "Taints", lexical[0].Chunk.Headings)
	assert.Positive(t, lexical[0].Score)

	semantic, err := index.Search(ctx, embedder, Semantic, "volume problems", 2)
	require.NoError(t, err)
	require.Len(t, semantic, 2)
	assert.Equal(t, "docs/storage.md", semantic[0].Chunk.Path)
	assert.InDelta(t, 1, semantic[0].Score, 0.01)

	hybrid, err := index.Search(ctx, embedder, Hybrid, "network traffic", 3)
	require.NoError(t, err)
	assert.Equal(t, "docs/network.md", hybrid[0].Chunk.Path)
	assert.Equal(t, []int{1, 2, 3}, []int{hybrid[0].Rank, hybrid[1].Rank, hybrid[2].Rank})

	_, err = index.Search(ctx, nil, Semantic, "volume", 1)
	assert.ErrorContains(t, err, "requires an embedder")
}

func TestLexicalQueryQuotesDistinctTerms(t *testing.T) {
	assert.Equal(t, `"pod" OR "pending" OR "near" OR "cpu"`, lexicalQuery(`Pod pending NEAR(cpu) pod*`))
	assert.Empty(t, lexicalQuery(`"*"`))
}

func TestFuseSumsReciprocalRanks(t *testing.T) {
	fused := fuse(3,
		[]scored{{id: 1}, {id: 2}, {id: 3}},
		[]scored{{id: 3}, {id: 2}, {id: 4}},
	)
	assert.Equal(t, []int64{3, 2, 1}, []int64{fused[0].id, fused[1].id, fused[2].id})
	assert.InDelta(t, 1.0/63+1.0/61, fused[0].score, 1e-12)

	tied := fuse(2, []scored{{id: 7}, {id: 5}}, []scored{{id: 5}, {id: 7}})
	assert.Equal(t, []int64{5, 7}, []int64{tied[0].id, tied[1].id})
}

func TestRenderCapsModelVisibleResult(t *testing.T) {
	hits := []Hit{
		{Rank: 1, Chunk: Chunk{Path: "docs/a.md", Title: "A", Body: "first"}},
		{Rank: 2, Chunk: Chunk{Path: "docs/b.md", Title: "B", Headings: "Sub", Body: "second"}},
	}
	rendered, truncated := Render(hits, 1000)
	assert.False(t, truncated)
	assert.Equal(t, "[1] docs/a.md\nA\nfirst\n\n[2] docs/b.md\nB > Sub\nsecond", rendered)

	rendered, truncated = Render(hits, 30)
	assert.True(t, truncated)
	assert.Len(t, rendered, 30)
	assert.True(t, strings.HasSuffix(rendered, "[truncated]"))

	rendered, _ = Render(nil, 30)
	assert.Equal(t, noResults, rendered)
}

func TestOpenIndexRejectsMissingFile(t *testing.T) {
	_, err := OpenIndex(context.Background(), filepath.Join(t.TempDir(), "missing.sqlite"))
	assert.Error(t, err)
}
