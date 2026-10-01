package agent

import (
	"context"
	"fmt"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/llamaenv"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/retrievalenv"
	ragagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/rag"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

// OpenSearch opens the configured retrieval index for the RAG agent and
// describes it for the run metadata. The caller closes search.Index.
func OpenSearch(ctx context.Context, values benchmarkconfig.RetrievalConfig) (*ragagent.Search, results.RetrievalProvenance, error) {
	settings, err := retrievalenv.Load(values)
	if err != nil {
		return nil, results.RetrievalProvenance{}, err
	}
	embedder, err := llamaenv.NewEmbeddingClient()
	if err != nil {
		return nil, results.RetrievalProvenance{}, err
	}
	// Fail before any attempt when the embedding service is unreachable,
	// instead of failing every search.
	if _, err := embedder.Embed(ctx, []string{"embedding service check"}); err != nil {
		return nil, results.RetrievalProvenance{}, fmt.Errorf("embedding service: %w; start it with make embedding-start", err)
	}
	path := retrieval.IndexPath(settings.IndexDir, settings.Chunking)
	digest, err := retrieval.FileSHA256(path)
	if err != nil {
		return nil, results.RetrievalProvenance{}, fmt.Errorf("hash retrieval index: %w", err)
	}
	embedding := embedder.Metadata()
	index, err := retrievalenv.OpenIndex(ctx, path, embedding)
	if err != nil {
		return nil, results.RetrievalProvenance{}, err
	}
	search := &ragagent.Search{Index: index, Embedder: embedder, Mode: settings.Mode, TopK: settings.TopK, MaxBytes: settings.MaxBytes}
	provenance := results.RetrievalProvenance{
		Mode:              string(settings.Mode),
		Chunking:          string(settings.Chunking),
		TopK:              settings.TopK,
		MaxBytes:          settings.MaxBytes,
		IndexSHA256:       digest,
		CorpusRevision:    index.Metadata().CorpusRevision,
		EmbeddingArtifact: embedding.Artifact,
		EmbeddingSHA256:   embedding.SHA256,
	}
	return search, provenance, nil
}
