package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

type settings struct {
	mode     retrieval.Mode
	chunking retrieval.Chunking
	topK     int
	maxBytes int
	indexDir string
}

func loadSettings(paths []string) (settings, error) {
	config, err := benchmarkconfig.Load(paths...)
	if err != nil {
		return settings{}, err
	}
	values := config.Retrieval
	if values.TopK == nil || values.MaxBytes == nil || values.IndexDir == "" {
		return settings{}, errors.New("retrieval.top_k, retrieval.max_bytes and retrieval.index_dir are required")
	}
	result := settings{topK: *values.TopK, maxBytes: *values.MaxBytes, indexDir: values.IndexDir}
	if result.topK < 1 || result.maxBytes < 1 {
		return settings{}, errors.New("retrieval.top_k and retrieval.max_bytes must be positive")
	}
	if result.mode, err = retrieval.ParseMode(values.Mode); err != nil {
		return settings{}, err
	}
	if result.chunking, err = retrieval.ParseChunking(values.Chunking); err != nil {
		return settings{}, err
	}
	return result, nil
}

// openIndex rejects an index built from another corpus revision or embedding
// model than the current ones.
func openIndex(ctx context.Context, path string, embedding inference.Metadata) (*retrieval.Index, error) {
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

// corpusCheckout verifies the checkout is unmodified at the pinned revision and
// returns its files, the corpus subtree and the revision.
func corpusCheckout(ctx context.Context) (fs.FS, string, string, error) {
	dir, subtree, pinned := os.Getenv("CORPUS_DIR"), os.Getenv("CORPUS_SUBTREE"), os.Getenv("CORPUS_REVISION")
	if dir == "" || subtree == "" || pinned == "" {
		return nil, "", "", errors.New("CORPUS_DIR, CORPUS_SUBTREE and CORPUS_REVISION are required")
	}
	checkout := filepath.Join(dir, "website")
	if revision, dirty := results.RepositoryProvenance(ctx, checkout); revision != pinned || dirty {
		return nil, "", "", fmt.Errorf("corpus checkout %s is at %q (modified: %t), want %s unmodified; run make corpus-download", checkout, revision, dirty, pinned)
	}
	return os.DirFS(checkout), subtree, pinned, nil
}
