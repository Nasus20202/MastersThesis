package retrievalenv

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

type unitEmbedder struct{}

func (unitEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for index := range embeddings {
		embeddings[index] = []float32{1, 0}
	}
	return embeddings, nil
}

var testSettings = Settings{MaxChunkBytes: 1536, WindowOverlap: 256, HybridCandidates: 50, RRFK: 60}

func TestOpenIndexRejectsOtherChunkingParameters(t *testing.T) {
	ctx := context.Background()
	path := retrieval.IndexPath(t.TempDir(), retrieval.Sections)
	embedding := inference.Metadata{Artifact: "repo@rev/model.gguf", SHA256: "hash"}
	_, err := retrieval.BuildIndex(ctx, path, []retrieval.Document{{Path: "docs/a.md", Title: "A", Text: "Body"}},
		retrieval.Sections, unitEmbedder{}, retrieval.IndexMetadata{CorpusRevision: "pinned", Embedding: embedding, MaxChunkBytes: 1024})
	require.NoError(t, err)

	t.Setenv("CORPUS_REVISION", "pinned")
	_, err = OpenIndex(ctx, path, embedding, testSettings)
	assert.ErrorContains(t, err, "other chunking parameters")
}

func TestOpenIndexRejectsOtherCorpusOrEmbedding(t *testing.T) {
	ctx := context.Background()
	path := retrieval.IndexPath(t.TempDir(), retrieval.Sections)
	embedding := inference.Metadata{Artifact: "repo@rev/model.gguf", SHA256: "hash"}
	_, err := retrieval.BuildIndex(ctx, path, []retrieval.Document{{Path: "docs/a.md", Title: "A", Text: "Body"}},
		retrieval.Sections, unitEmbedder{}, retrieval.IndexMetadata{CorpusRevision: "pinned", Embedding: embedding, MaxChunkBytes: 1536})
	require.NoError(t, err)

	t.Setenv("CORPUS_REVISION", "pinned")
	index, err := OpenIndex(ctx, path, embedding, testSettings)
	require.NoError(t, err)
	require.NoError(t, index.Close())
	_, err = OpenIndex(ctx, path, inference.Metadata{Artifact: "other@rev/model.gguf", SHA256: "hash"}, testSettings)
	assert.ErrorContains(t, err, "was embedded with")

	t.Setenv("CORPUS_REVISION", "newer")
	_, err = OpenIndex(ctx, path, embedding, testSettings)
	assert.ErrorContains(t, err, "was built from corpus")
}
