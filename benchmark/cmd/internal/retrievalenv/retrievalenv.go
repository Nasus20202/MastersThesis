// Package retrievalenv validates the retrieval configuration and opens the
// retrieval index pinned by retrieval.env, for the retrieval CLI and the RAG
// condition.
package retrievalenv

import (
	"context"
	"errors"
	"fmt"
	"os"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

type Settings struct {
	Mode     retrieval.Mode
	Chunking retrieval.Chunking
	TopK     int
	MaxBytes int
	IndexDir string
}

func Load(values benchmarkconfig.RetrievalConfig) (Settings, error) {
	if values.TopK == nil || values.MaxBytes == nil || values.IndexDir == "" {
		return Settings{}, errors.New("retrieval.top_k, retrieval.max_bytes and retrieval.index_dir are required")
	}
	result := Settings{TopK: *values.TopK, MaxBytes: *values.MaxBytes, IndexDir: values.IndexDir}
	if result.TopK < 1 || result.MaxBytes < 1 {
		return Settings{}, errors.New("retrieval.top_k and retrieval.max_bytes must be positive")
	}
	var err error
	if result.Mode, err = retrieval.ParseMode(values.Mode); err != nil {
		return Settings{}, err
	}
	if result.Chunking, err = retrieval.ParseChunking(values.Chunking); err != nil {
		return Settings{}, err
	}
	return result, nil
}

// OpenIndex rejects an index built from another corpus revision or embedding
// model than the current ones.
func OpenIndex(ctx context.Context, path string, embedding inference.Metadata) (*retrieval.Index, error) {
	index, err := retrieval.OpenIndex(ctx, path)
	if err != nil {
		return nil, err
	}
	built := index.Metadata()
	if pinned := os.Getenv("CORPUS_REVISION"); built.CorpusRevision != pinned {
		err = fmt.Errorf("index %s was built from corpus %s, want %s", path, built.CorpusRevision, pinned)
	} else if built.Embedding.Artifact != embedding.Artifact || built.Embedding.SHA256 != embedding.SHA256 {
		err = fmt.Errorf("index %s was embedded with %s, the embedding client uses %s", path, built.Embedding.Artifact, embedding.Artifact)
	}
	if err != nil {
		_ = index.Close()
		return nil, fmt.Errorf("%w; rebuild it with make retrieval-index", err)
	}
	return index, nil
}
