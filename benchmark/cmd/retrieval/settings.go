package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/retrievalenv"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func loadSettings(paths []string) (retrievalenv.Settings, error) {
	config, err := benchmarkconfig.Load(paths...)
	if err != nil {
		return retrievalenv.Settings{}, err
	}
	return retrievalenv.Load(config.Retrieval)
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
