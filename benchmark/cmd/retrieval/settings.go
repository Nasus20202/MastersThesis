package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/config"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
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

// corpusCheckout returns the checkout directory and subtree after verifying
// the checkout is at the pinned revision.
func corpusCheckout(ctx context.Context) (string, string, string, error) {
	dir, subtree, pinned := os.Getenv("CORPUS_DIR"), os.Getenv("CORPUS_SUBTREE"), os.Getenv("CORPUS_REVISION")
	if dir == "" || subtree == "" || pinned == "" {
		return "", "", "", errors.New("CORPUS_DIR, CORPUS_SUBTREE and CORPUS_REVISION are required")
	}
	checkout := filepath.Join(dir, "website")
	revision, err := retrieval.CheckoutRevision(ctx, checkout)
	if err != nil {
		return "", "", "", fmt.Errorf("%w (run make corpus-download)", err)
	}
	if revision != pinned {
		return "", "", "", fmt.Errorf("corpus checkout is at %s, want %s", revision, pinned)
	}
	return checkout, subtree, pinned, nil
}

func newEmbedder() (llama.Adapter, error) {
	model, port := os.Getenv("EMBEDDING_MODEL_NAME"), os.Getenv("EMBEDDING_PORT")
	if model == "" || port == "" {
		return llama.Adapter{}, errors.New("EMBEDDING_MODEL_NAME and EMBEDDING_PORT are required")
	}
	host := os.Getenv("LLAMA_CLIENT_HOST")
	if host == "" {
		host = "127.0.0.1"
	}
	client, err := llama.NewClient(llama.Config{
		BaseURL: fmt.Sprintf("http://%s:%s", host, port),
		Model:   model,
		Metadata: inference.Metadata{
			Provider:     "llama.cpp",
			Model:        model,
			Artifact:     fmt.Sprintf("%s@%s/%s", os.Getenv("EMBEDDING_MODEL_REPOSITORY"), os.Getenv("EMBEDDING_MODEL_REVISION"), os.Getenv("EMBEDDING_MODEL_FILE")),
			Quantization: os.Getenv("EMBEDDING_MODEL_QUANTIZATION"),
			SHA256:       os.Getenv("EMBEDDING_MODEL_SHA256"),
		},
	})
	if err != nil {
		return llama.Adapter{}, fmt.Errorf("create embedding client: %w", err)
	}
	return llama.NewAdapter(client)
}
