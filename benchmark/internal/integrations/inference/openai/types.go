package openai

import (
	"encoding/json"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// message is sent without reasoning_content, which strict APIs reject; it is
// only read from responses.
type message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
}

type responseMessage struct {
	message
	ReasoningContent string `json:"reasoning_content,omitempty"`
}

type toolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function toolFunction `json:"function"`
}

type toolFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type tool struct {
	Type     string             `json:"type"`
	Function functionDefinition `json:"function"`
}

type functionDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

type chatRequest struct {
	Model               string    `json:"model"`
	Messages            []message `json:"messages"`
	Tools               []tool    `json:"tools,omitempty"`
	Temperature         *float64  `json:"temperature,omitempty"`
	MaxCompletionTokens *int      `json:"max_completion_tokens,omitempty"`
	Seed                *int      `json:"seed,omitempty"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message      responseMessage `json:"message"`
		FinishReason string          `json:"finish_reason"`
	} `json:"choices"`
	Usage *usage `json:"usage"`
}

type usage struct {
	PromptTokens        int `json:"prompt_tokens"`
	CompletionTokens    int `json:"completion_tokens"`
	TotalTokens         int `json:"total_tokens"`
	PromptTokensDetails *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
}

func toMessages(messages []inference.Message) []message {
	converted := make([]message, len(messages))
	for i, m := range messages {
		converted[i] = message{Role: m.Role, Content: m.Content, Name: m.Name, ToolCallID: m.ToolCallID}
		for _, call := range m.ToolCalls {
			converted[i].ToolCalls = append(converted[i].ToolCalls, toolCall{
				ID:       call.ID,
				Type:     call.Type,
				Function: toolFunction{Name: call.Name, Arguments: call.Arguments},
			})
		}
	}
	return converted
}

func toTools(tools []inference.Tool) []tool {
	converted := make([]tool, len(tools))
	for i, t := range tools {
		converted[i] = tool{Type: "function", Function: functionDefinition{Name: t.Name, Description: t.Description, Parameters: t.Parameters}}
	}
	return converted
}

func fromMessage(m responseMessage) inference.Message {
	converted := inference.Message{
		Role:             m.Role,
		Content:          m.Content,
		ReasoningContent: m.ReasoningContent,
		Name:             m.Name,
		ToolCallID:       m.ToolCallID,
	}
	for _, call := range m.ToolCalls {
		converted.ToolCalls = append(converted.ToolCalls, inference.ToolCall{
			ID:        call.ID,
			Type:      call.Type,
			Name:      call.Function.Name,
			Arguments: call.Function.Arguments,
		})
	}
	return converted
}

func fromUsage(u *usage) *inference.Usage {
	if u == nil {
		return nil
	}
	converted := &inference.Usage{
		PromptTokens:     u.PromptTokens,
		CompletionTokens: u.CompletionTokens,
		TotalTokens:      u.TotalTokens,
	}
	if u.PromptTokensDetails != nil {
		converted.CachedTokens = u.PromptTokensDetails.CachedTokens
	}
	return converted
}
