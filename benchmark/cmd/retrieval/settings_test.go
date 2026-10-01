package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func TestOpenIndexRejectsOtherCorpusOrEmbedding(t *testing.T) {
	ctx := context.Background()
	path := retrieval.IndexPath(t.TempDir(), retrieval.Sections)
	embedding := inference.Metadata{Artifact: "repo@rev/model.gguf", SHA256: "hash"}
	_, err := retrieval.BuildIndex(ctx, path, []retrieval.Document{{Path: "docs/a.md", Title: "A", Text: "Body"}},
		retrieval.Sections, unitEmbedder{}, retrieval.IndexMetadata{CorpusRevision: "pinned", Embedding: embedding})
	require.NoError(t, err)

	t.Setenv("CORPUS_REVISION", "pinned")
	index, err := openIndex(ctx, path, embedding)
	require.NoError(t, err)
	require.NoError(t, index.Close())
	_, err = openIndex(ctx, path, inference.Metadata{Artifact: "other@rev/model.gguf", SHA256: "hash"})
	assert.ErrorContains(t, err, "was embedded with")

	t.Setenv("CORPUS_REVISION", "newer")
	_, err = openIndex(ctx, path, embedding)
	assert.ErrorContains(t, err, "was built from corpus")
}

func TestCorpusCheckoutRequiresPinnedUnmodifiedRevision(t *testing.T) {
	dir := t.TempDir()
	checkout := filepath.Join(dir, "website")
	file := filepath.Join(checkout, "docs", "a.md")
	git := func(args ...string) string {
		output, err := exec.CommandContext(t.Context(), "git", append([]string{"-C", checkout, "-c", "user.name=test", "-c", "user.email=test@example.com", "-c", "commit.gpgsign=false"}, args...)...).Output()
		require.NoError(t, err)
		return strings.TrimSpace(string(output))
	}
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o750))
	require.NoError(t, os.WriteFile(file, []byte("a\n"), 0o600))
	git("init", "--quiet")
	git("add", ".")
	git("commit", "--quiet", "--message", "corpus")
	t.Setenv("CORPUS_DIR", dir)
	t.Setenv("CORPUS_SUBTREE", "docs")
	t.Setenv("CORPUS_REVISION", git("rev-parse", "HEAD"))

	_, _, _, err := corpusCheckout(context.Background())
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(file, []byte("changed\n"), 0o600))
	_, _, _, err = corpusCheckout(context.Background())
	assert.ErrorContains(t, err, "modified: true")
}
