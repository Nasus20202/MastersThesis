package evaluation

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

// topicEmbedder embeds text by the topic words it contains.
type topicEmbedder struct {
	calls int
}

func (e *topicEmbedder) Embed(_ context.Context, texts []string) ([][]float32, error) {
	e.calls++
	embeddings := make([][]float32, len(texts))
	for index, text := range texts {
		vector := []float32{0.01, 0.01, 0.01}
		for topic, word := range []string{"volume", "network", "schedul"} {
			if strings.Contains(strings.ToLower(text), word) {
				vector[topic] = 1
			}
		}
		embeddings[index] = vector
	}
	return embeddings, nil
}

func TestEvaluateScoresEveryConfiguration(t *testing.T) {
	embedder := &topicEmbedder{}
	path := retrieval.IndexPath(t.TempDir(), retrieval.Sections)
	_, err := retrieval.BuildIndex(context.Background(), path, []retrieval.Document{
		{Path: "docs/storage.md", Title: "Storage", Text: "A PersistentVolumeClaim binds a volume."},
		{Path: "docs/network.md", Title: "Networking", Text: "NetworkPolicy isolates network traffic between Pods."},
		{Path: "docs/scheduler.md", Title: "Scheduling", Text: "The scheduler avoids nodes with taints."},
	}, retrieval.Sections, embedder, retrieval.IndexMetadata{})
	require.NoError(t, err)
	index, err := retrieval.OpenIndex(context.Background(), path)
	require.NoError(t, err)
	defer index.Close()

	probes := []ProbeQueries{{
		Probe: Probe{ID: "K01", Prompt: "Pods avoid tainted nodes", Primary: []string{"docs/scheduler.md"}, Relevant: []string{"docs/scheduler.md", "docs/network.md"}},
		// Only the first rewrite matches a relevant file.
		Queries: []string{"taints", "PersistentVolumeClaim"},
	}}
	indexes := map[retrieval.Chunking]*retrieval.Index{retrieval.Sections: index, retrieval.Windows: index}

	report, err := Evaluate(context.Background(), indexes, embedder, probes, 8192)
	require.NoError(t, err)
	require.Len(t, report.Configs, len(retrieval.Modes)*len(retrieval.Chunkings)*len(Ks))
	assert.Len(t, report.Results, len(retrieval.Modes)*len(retrieval.Chunkings)*3)

	lexical := report.Configs[0]
	assert.Equal(t, "lexical/sections/k=3", lexical.Name())
	assert.Equal(t, 1.0, lexical.AsIs.Hit)
	assert.Equal(t, 0.5, lexical.Rewrite.Hit)
	assert.Equal(t, 0.5, lexical.Rewrite.PrimaryHit)
	assert.Equal(t, 0, lexical.Truncated)
	// One index build call plus one call per distinct query.
	assert.Equal(t, 1+3, embedder.calls)
}

func TestSelectAppliesTieRule(t *testing.T) {
	configs := []ConfigScore{
		{Mode: retrieval.Hybrid, Chunking: retrieval.Sections, K: 5, Rewrite: Scores{Hit: 0.90, RR: 0.70}},
		{Mode: retrieval.Semantic, Chunking: retrieval.Windows, K: 3, Rewrite: Scores{Hit: 0.87, RR: 0.70}},
		{Mode: retrieval.Lexical, Chunking: retrieval.Sections, K: 3, Rewrite: Scores{Hit: 0.80, RR: 0.95}},
		{Mode: retrieval.Lexical, Chunking: retrieval.Windows, K: 3, Rewrite: Scores{Hit: 0.86, RR: 0.70}},
	}
	// With 24 probes 0.87 and 0.86 tie with 0.90; equal MRR then prefers k=3
	// and lexical. 0.80 is more than one probe behind.
	assert.Equal(t, configs[3], Select(configs, 24))
	assert.Equal(t, configs[0], Select(configs, 100))
}
