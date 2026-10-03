// Package rag implements the RAG condition: the prompt condition's Bash tool
// loop with a search_docs tool over the frozen documentation corpus, and a
// search the harness runs when a kubectl change fails.
package rag

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval"
)

// prompt.md is the prompt condition's prompt followed by the search
// instructions, so the two conditions differ only in retrieval.
//
//go:embed prompt.md
var defaultSystemPrompt string

const searchToolName = "search_docs"

// Error searches: when a kubectl change fails, the harness searches with the
// command shape and the error line and appends the top excerpts to the Bash
// result. Bounded per attempt so repeated failures do not fill the context.
const (
	errorSearchTopK     = 3
	errorSearchMaxBytes = 4 * 1024
	maxErrorSearches    = 3
	errorTrigger        = "error"
)

// Search is the configured retrieval over one opened index. It is read-only
// and shared by concurrent attempts.
type Search struct {
	Index    *retrieval.Index
	Embedder inference.Embedder
	Mode     retrieval.Mode
	TopK     int
	MaxBytes int
}

// New builds the RAG condition. A blank systemPrompt uses the embedded
// prompt.md.
func New(client inference.Client, shell common.Shell, config common.Config, systemPrompt string, search *Search) (*common.Condition, error) {
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = defaultSystemPrompt
	}
	if search == nil || search.Index == nil {
		return nil, errors.New("rag search index is required")
	}
	failures := &errorSearch{search: search}
	bash, err := common.NewHintedBashTool(shell, failures.hint)
	if err != nil {
		return nil, err
	}
	return common.NewCondition("rag", client, []common.Tool{bash, searchTool{search: search}}, config, systemPrompt, nil)
}

// errorSearch holds one attempt's error searches; New builds one per attempt.
type errorSearch struct {
	search  *Search
	queries []string
}

func (e *errorSearch) hint(ctx context.Context, change common.FailedChange) (string, any) {
	query := change.Query()
	if len(e.queries) >= maxErrorSearches || slices.Contains(e.queries, query) {
		return "", nil
	}
	e.queries = append(e.queries, query)
	evidence := SearchEvidence{Query: query, Trigger: errorTrigger}
	hits, err := e.search.Index.Search(ctx, e.search.Embedder, e.search.Mode, query, errorSearchTopK)
	if err != nil {
		evidence.Error = err.Error()
		return "", evidence
	}
	if len(hits) == 0 {
		return "", evidence
	}
	content, truncated := retrieval.Render(hits, errorSearchMaxBytes)
	evidence.Truncated, evidence.Hits = truncated, hits
	return fmt.Sprintf("Documentation for this error (search: %q):\n%s", query, content), evidence
}

type searchTool struct{ search *Search }

func (searchTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        searchToolName,
		Description: "Search the official Kubernetes documentation and return the most relevant excerpts.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string","description":"Short search query"}},"required":["query"],"additionalProperties":false}`),
	}
}

// SearchEvidence records one search for later analysis; the model-visible
// excerpts are in the tool call content. Trigger is empty for the agent's own
// search_docs calls and "error" for a search after a failed change, which is
// recorded as the hint of that Bash call.
type SearchEvidence struct {
	Query     string          `json:"query"`
	Trigger   string          `json:"trigger,omitempty"`
	Truncated bool            `json:"truncated"`
	Hits      []retrieval.Hit `json:"hits"`
	Error     string          `json:"error,omitempty"`
}

func (t searchTool) Execute(ctx context.Context, call inference.ToolCall) common.ToolResult {
	var arguments struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		return common.ToolResult{Content: fmt.Sprintf("malformed %s arguments: %v", searchToolName, err), Error: fmt.Errorf("malformed %s arguments: %w", searchToolName, err)}
	}
	query := strings.TrimSpace(arguments.Query)
	if query == "" {
		err := errors.New("search query is required")
		return common.ToolResult{Content: err.Error(), Error: err}
	}
	hits, err := t.search.Index.Search(ctx, t.search.Embedder, t.search.Mode, query, t.search.TopK)
	if err != nil {
		err = fmt.Errorf("search failed: %w", err)
		return common.ToolResult{Content: err.Error(), Details: SearchEvidence{Query: query}, Error: err}
	}
	content, truncated := retrieval.Render(hits, t.search.MaxBytes)
	return common.ToolResult{Content: content, Details: SearchEvidence{Query: query, Truncated: truncated, Hits: hits}}
}
