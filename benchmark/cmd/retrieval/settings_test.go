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
)

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
