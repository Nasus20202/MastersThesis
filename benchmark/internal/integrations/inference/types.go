package inference

import "encoding/json"

// Metadata identifies the inference implementation and the model execution
// configuration used for an attempt.
type Metadata struct {
	Provider        string            `json:"provider,omitempty"`
	Model           string            `json:"model,omitempty"`
	Artifact        string            `json:"artifact,omitempty"`
	RuntimeSettings map[string]string `json:"runtime_settings,omitempty"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type Result struct {
	ID           string   `json:"id"`
	Message      Message  `json:"message"`
	FinishReason string   `json:"finish_reason"`
	Usage        *Usage   `json:"usage,omitempty"`
	Timings      *Timings `json:"timings,omitempty"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Timings struct {
	PromptN            int     `json:"prompt_n"`
	PromptMS           float64 `json:"prompt_ms"`
	PromptPerSecond    float64 `json:"prompt_per_second"`
	PredictedN         int     `json:"predicted_n"`
	PredictedMS        float64 `json:"predicted_ms"`
	PredictedPerSecond float64 `json:"predicted_per_second"`
}
