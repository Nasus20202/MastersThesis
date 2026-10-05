package common

import (
	"context"
	"errors"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const (
	TerminationCompleted    = "completed"
	TerminationTurnLimit    = "turn_limit"
	TerminationToolLimit    = "tool_limit"
	TerminationTokenLimit   = "token_limit"
	TerminationFinishReason = "finish_reason"
	TerminationInference    = "inference_error"
	TerminationCancellation = "cancellation"
	TerminationTimeout      = "timeout"
)

// ResponseEvidence preserves the raw response and the client round-trip
// duration for later result persistence.
type ResponseEvidence struct {
	Response        inference.Result `json:"response"`
	DurationSeconds float64          `json:"duration_seconds"`
}

// TokenUsage contains the aggregate token usage reported by all model
// responses in one loop run. PeakContextTokens is the largest prompt plus
// completion of a single response: the most context the attempt occupied.
type TokenUsage struct {
	PromptTokens      int `json:"prompt_tokens"`
	CompletionTokens  int `json:"completion_tokens"`
	TotalTokens       int `json:"total_tokens"`
	CachedTokens      int `json:"cached_tokens,omitempty"`
	PeakContextTokens int `json:"peak_context_tokens"`
}

// Result is the complete model-loop evidence produced by Run. ContextOverflow
// marks an attempt that ran out of model context: the server rejected a
// request as too long, or generation stopped at the length limit without a
// configured max_tokens being reached. A rejected request reports no usage, so
// TokenUsage.PeakContextTokens then underestimates the context.
type Result struct {
	Condition       string              `json:"condition,omitempty"`
	Task            string              `json:"task"`
	Inference       inference.Metadata  `json:"inference"`
	LoopConfig      Config              `json:"loop_config"`
	Tools           []inference.Tool    `json:"tools"`
	Messages        []inference.Message `json:"messages"`
	Responses       []ResponseEvidence  `json:"responses"`
	TokenUsage      TokenUsage          `json:"token_usage"`
	ToolCalls       []ToolCallEvidence  `json:"tool_calls"`
	Turns           int                 `json:"turns"`
	ToolCallCount   int                 `json:"tool_call_count"`
	Termination     string              `json:"termination"`
	Error           string              `json:"error,omitempty"`
	DurationSeconds float64             `json:"duration_seconds"`
	ContextOverflow bool                `json:"context_overflow,omitempty"`
}

func (u *TokenUsage) add(usage *inference.Usage) {
	if usage == nil {
		return
	}
	u.PromptTokens += usage.PromptTokens
	u.CompletionTokens += usage.CompletionTokens
	u.TotalTokens += usage.TotalTokens
	u.CachedTokens += usage.CachedTokens
	u.PeakContextTokens = max(u.PeakContextTokens, usage.PromptTokens+usage.CompletionTokens)
}

func (r *Result) finish(termination string, loopErr error) (Result, error) {
	r.Termination = termination
	if loopErr != nil {
		r.Error = loopErr.Error()
	}
	return *r, loopErr
}

func (r *Result) finishContext(ctxErr error) (Result, error) {
	termination := TerminationCancellation
	if errors.Is(ctxErr, context.DeadlineExceeded) {
		termination = TerminationTimeout
	}
	return r.finish(termination, ctxErr)
}
