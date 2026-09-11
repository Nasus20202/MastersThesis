package inference

import "encoding/json"

// Message is the provider-neutral message representation used by agents.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall describes a callable tool requested by the model.
type ToolCall struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Tool is the provider-neutral definition of a callable model tool.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

// Result is one provider-neutral model response.
type Result struct {
	ID           string   `json:"id"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
	Usage        *Usage   `json:"usage,omitempty"`
	Timings      *Timings `json:"timings,omitempty"`
}

// Usage contains token counts returned by an inference provider.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Timings contains server-side generation measurements when provided.
type Timings struct {
	PromptN            int     `json:"prompt_n"`
	PromptMS           float64 `json:"prompt_ms"`
	PromptPerSecond    float64 `json:"prompt_per_second"`
	PredictedN         int     `json:"predicted_n"`
	PredictedMS        float64 `json:"predicted_ms"`
	PredictedPerSecond float64 `json:"predicted_per_second"`
}
