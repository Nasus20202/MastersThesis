package evaluation

import (
	"context"
	"slices"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// cachingEmbedder embeds each distinct query once, although it is searched in
// every mode and chunking.
type cachingEmbedder struct {
	inner inference.Embedder
	cache map[string][]float32
}

func newCachingEmbedder(inner inference.Embedder) inference.Embedder {
	if inner == nil {
		return nil
	}
	return &cachingEmbedder{inner: inner, cache: make(map[string][]float32)}
}

func (e *cachingEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	var missing []string
	for _, text := range texts {
		if _, ok := e.cache[text]; !ok && !slices.Contains(missing, text) {
			missing = append(missing, text)
		}
	}
	if len(missing) > 0 {
		embeddings, err := e.inner.Embed(ctx, missing)
		if err != nil {
			return nil, err
		}
		for index, text := range missing {
			e.cache[text] = embeddings[index]
		}
	}
	result := make([][]float32, len(texts))
	for index, text := range texts {
		result[index] = e.cache[text]
	}
	return result, nil
}
