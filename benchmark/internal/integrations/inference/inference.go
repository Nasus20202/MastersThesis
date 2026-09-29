// Package inference defines the chat-completion interface used by agents.
package inference

import (
	"context"
)

// Client sends chat-completion requests. llama.Adapter implements it.
type Client interface {
	Chat(context.Context, []Message, []Tool, Options) (Result, error)
}

// MetadataProvider is implemented by clients that can report provider and
// model metadata for the attempt record.
type MetadataProvider interface {
	Metadata() Metadata
}

type Options struct {
	Temperature *float64
	MaxTokens   *int
}
