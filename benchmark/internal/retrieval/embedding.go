package retrieval

import (
	"context"
	"fmt"
	"log/slog"

	vec "github.com/asg017/sqlite-vec-go-bindings/cgo"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const embeddingBatchSize = 32

// The prompts follow the retrieval task format of the EmbeddingGemma model card.
func documentPrompt(chunk Chunk) string {
	text := chunk.Body
	if chunk.Headings != "" {
		text = chunk.Headings + "\n" + text
	}
	return "title: " + chunk.Title + " | text: " + text
}

func queryPrompt(query string) string {
	return "task: search result | query: " + query
}

func embedChunks(ctx context.Context, embedder inference.Embedder, chunks []Chunk) ([][]float32, error) {
	embeddings := make([][]float32, 0, len(chunks))
	for start := 0; start < len(chunks); start += embeddingBatchSize {
		batch := chunks[start:min(start+embeddingBatchSize, len(chunks))]
		texts := make([]string, len(batch))
		for index, chunk := range batch {
			texts[index] = documentPrompt(chunk)
		}
		vectors, err := embedder.Embed(ctx, texts)
		if err != nil {
			return nil, fmt.Errorf("embed chunks %d-%d: %w", start+1, start+len(batch), err)
		}
		embeddings = append(embeddings, vectors...)
		if done := len(embeddings); done%(embeddingBatchSize*50) == 0 || done == len(chunks) {
			slog.InfoContext(ctx, "embedded chunks", "done", done, "total", len(chunks))
		}
	}
	return embeddings, nil
}

func embedQuery(ctx context.Context, embedder inference.Embedder, query string) ([]byte, error) {
	if embedder == nil {
		return nil, fmt.Errorf("semantic search requires an embedder")
	}
	embeddings, err := embedder.Embed(ctx, []string{queryPrompt(query)})
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}
	return vec.SerializeFloat32(embeddings[0])
}
