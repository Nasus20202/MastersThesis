// Package inference contains provider-neutral inference boundaries used by
// benchmark agents.
package inference

import (
	"context"
)

// Client is the chat-completion boundary consumed by agents. Provider
// adapters, such as llama.Adapter, implement this interface.
type Client interface {
	Chat(context.Context, []Message, []Tool, Options) (Result, error)
}

// Options contains provider-neutral generation controls.
type Options struct {
	Temperature *float64
	MaxTokens   *int
}
