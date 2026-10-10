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

	MaxChunkBytes    int
	WindowOverlap    int
	HybridCandidates int
	RRFK             int
}

func Load(values benchmarkconfig.RetrievalConfig) (Settings, error) {
	if values.TopK == nil || values.MaxBytes == nil || values.IndexDir == "" ||
		values.MaxChunkBytes == nil || values.WindowOverlap == nil || values.HybridCandidates == nil || values.RRFK == nil {
		return Settings{}, errors.New("retrieval.top_k, max_bytes, index_dir, max_chunk_bytes, window_overlap, hybrid_candidates and rrf_k are required")
	}
	result := Settings{
		TopK: *values.TopK, MaxBytes: *values.MaxBytes, IndexDir: values.IndexDir,
		MaxChunkBytes: *values.MaxChunkBytes, WindowOverlap: *values.WindowOverlap,
		HybridCandidates: *values.HybridCandidates, RRFK: *values.RRFK,
	}
	if result.TopK < 1 || result.MaxBytes < 1 || result.MaxChunkBytes < 1 || result.HybridCandidates < 1 || result.RRFK < 1 {
		return Settings{}, errors.New("retrieval.top_k, max_bytes, max_chunk_bytes, hybrid_candidates and rrf_k must be positive")
	}
	if result.WindowOverlap < 0 || result.WindowOverlap >= result.MaxChunkBytes {
		return Settings{}, errors.New("retrieval.window_overlap must be at least 0 and below max_chunk_bytes")
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

// OpenIndex rejects an index built from another corpus revision, embedding
// model or chunking than the current ones.
func OpenIndex(ctx context.Context, path string, embedding inference.Metadata, settings Settings) (*retrieval.Index, error) {
	index, err := retrieval.OpenIndex(ctx, path)
	if err != nil {
		return nil, err
	}
	built := index.Metadata()
	wantOverlap := settings.WindowOverlap
	if built.Chunking != retrieval.Windows {
		wantOverlap = 0
	}
	if built.MaxChunkBytes != settings.MaxChunkBytes || built.WindowOverlap != wantOverlap {
		err = fmt.Errorf("index %s was built with other chunking parameters", path)
	} else if pinned := os.Getenv("CORPUS_REVISION"); built.CorpusRevision != pinned {
		err = fmt.Errorf("index %s was built from corpus %s, want %s", path, built.CorpusRevision, pinned)
	} else if built.Embedding.Artifact != embedding.Artifact || built.Embedding.SHA256 != embedding.SHA256 {
		err = fmt.Errorf("index %s was embedded with %s, the embedding client uses %s", path, built.Embedding.Artifact, embedding.Artifact)
	}
	if err != nil {
		_ = index.Close()
		return nil, fmt.Errorf("%w; rebuild it with make retrieval-index", err)
	}
	index.Hybrid = retrieval.HybridParams{Candidates: settings.HybridCandidates, RRFK: settings.RRFK}
	return index, nil
}
