package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
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

// corpusCheckout verifies the checkout is at the pinned revision and returns
// its files, the corpus subtree and the revision.
func corpusCheckout(ctx context.Context) (fs.FS, string, string, error) {
	dir, subtree, pinned := os.Getenv("CORPUS_DIR"), os.Getenv("CORPUS_SUBTREE"), os.Getenv("CORPUS_REVISION")
	if dir == "" || subtree == "" || pinned == "" {
		return nil, "", "", errors.New("CORPUS_DIR, CORPUS_SUBTREE and CORPUS_REVISION are required")
	}
	checkout := filepath.Join(dir, "website")
	revision, err := retrieval.CheckoutRevision(ctx, checkout)
	if err != nil {
		return nil, "", "", fmt.Errorf("%w (run make corpus-download)", err)
	}
	if revision != pinned {
		return nil, "", "", fmt.Errorf("corpus checkout is at %s, want %s", revision, pinned)
	}
	return os.DirFS(checkout), subtree, pinned, nil
}
