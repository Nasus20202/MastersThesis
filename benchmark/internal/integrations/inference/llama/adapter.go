package llama

import (
	"context"
	"errors"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// Adapter exposes a llama.Client through the provider-neutral inference
// client interface consumed by agents.
type Adapter struct {
	client *Client
}

// NewAdapter wraps a llama client for use by an agent.
func NewAdapter(client *Client) (inference.Client, error) {
	if client == nil {
		return nil, errors.New("llama client is required")
	}
	return Adapter{client: client}, nil
}

func (a Adapter) Chat(ctx context.Context, messages []inference.Message, tools []inference.Tool, options inference.Options) (inference.Result, error) {
	response, err := a.client.Chat(ctx, ChatRequest{
		Messages:    toLlamaMessages(messages),
		Tools:       toLlamaTools(tools),
		Temperature: options.Temperature,
		MaxTokens:   options.MaxTokens,
	})
	if err != nil {
		return inference.Result{}, err
	}
	if len(response.Choices) == 0 {
		return inference.Result{}, errors.New("llama response contained no choices")
	}
	return inference.Result{
		ID:           response.ID,
		Message:      fromLlamaMessage(response.Choices[0].Message),
		FinishReason: response.Choices[0].FinishReason,
		Usage:        fromLlamaUsage(response.Usage),
		Timings:      fromLlamaTimings(response.Timings),
	}, nil
}

func toLlamaMessages(messages []inference.Message) []Message {
	converted := make([]Message, len(messages))
	for index, message := range messages {
		converted[index] = Message{
			Role:       message.Role,
			Content:    message.Content,
			Name:       message.Name,
			ToolCallID: message.ToolCallID,
			ToolCalls:  toLlamaToolCalls(message.ToolCalls),
		}
	}
	return converted
}

func toLlamaToolCalls(calls []inference.ToolCall) []ToolCall {
	converted := make([]ToolCall, len(calls))
	for index, call := range calls {
		converted[index] = ToolCall{
			ID:       call.ID,
			Type:     call.Type,
			Function: ToolCallFunction{Name: call.Name, Arguments: call.Arguments},
		}
	}
	return converted
}

func toLlamaTools(tools []inference.Tool) []Tool {
	converted := make([]Tool, len(tools))
	for index, tool := range tools {
		converted[index] = Tool{
			Type:     "function",
			Function: FunctionDefinition{Name: tool.Name, Description: tool.Description, Parameters: tool.Parameters},
		}
	}
	return converted
}

func fromLlamaMessage(message Message) inference.Message {
	return inference.Message{
		Role:       message.Role,
		Content:    message.Content,
		Name:       message.Name,
		ToolCallID: message.ToolCallID,
		ToolCalls:  fromLlamaToolCalls(message.ToolCalls),
	}
}

func fromLlamaToolCalls(calls []ToolCall) []inference.ToolCall {
	converted := make([]inference.ToolCall, len(calls))
	for index, call := range calls {
		converted[index] = inference.ToolCall{
			ID:        call.ID,
			Type:      call.Type,
			Name:      call.Function.Name,
			Arguments: call.Function.Arguments,
		}
	}
	return converted
}

func fromLlamaUsage(usage *Usage) *inference.Usage {
	if usage == nil {
		return nil
	}
	return &inference.Usage{
		PromptTokens:     usage.PromptTokens,
		CompletionTokens: usage.CompletionTokens,
		TotalTokens:      usage.TotalTokens,
	}
}

func fromLlamaTimings(timings *Timings) *inference.Timings {
	if timings == nil {
		return nil
	}
	return &inference.Timings{
		PromptN:            timings.PromptN,
		PromptMS:           timings.PromptMS,
		PromptPerSecond:    timings.PromptPerSecond,
		PredictedN:         timings.PredictedN,
		PredictedMS:        timings.PredictedMS,
		PredictedPerSecond: timings.PredictedPerSecond,
	}
}

var _ inference.Client = Adapter{}
