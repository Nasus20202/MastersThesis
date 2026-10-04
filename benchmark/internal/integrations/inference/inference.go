// Package inference defines the chat-completion and embedding interfaces used
// by agents and retrieval.
package inference

import (
	"context"
	"errors"
)

// ErrContextOverflow reports a request that does not fit in the model context.
var ErrContextOverflow = errors.New("request exceeds the model context")

// Client sends chat-completion requests. llama.Adapter implements it.
type Client interface {
	Chat(context.Context, []Message, []Tool, Options) (Result, error)
}

// MetadataProvider is implemented by clients that can report provider and
// model metadata for the attempt record.
type MetadataProvider interface {
	Metadata() Metadata
}

// Embedder turns texts into embedding vectors, one per input in order.
type Embedder interface {
	Embed(context.Context, []string) ([][]float32, error)
}

type Options struct {
	Temperature *float64
	MaxTokens   *int
	Seed        *int
}
