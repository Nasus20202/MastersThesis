package llama

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdapterTranslatesGenericChatToLlama(t *testing.T) {
	client, err := NewClient(Config{
		BaseURL: "http://llama.test",
		Model:   "gemma-test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			var payload struct {
				Model string `json:"model"`
				ChatRequest
			}
			require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
			assert.Equal(t, "gemma-test", payload.Model)
			assert.Equal(t, "Inspect", payload.Messages[0].Content)
			assert.Equal(t, "bash", payload.Tools[0].Function.Name)
			assert.Equal(t, 64, *payload.MaxTokens)
			return testResponse(http.StatusOK, `{
                "id": "chatcmpl-adapter",
                "choices": [{
                    "index": 0,
                    "message": {"role": "assistant", "content": "done", "tool_calls": [{"id": "call-1", "type": "function", "function": {"name": "bash", "arguments": "{\"command\":\"true\"}"}}]},
                    "finish_reason": "tool_calls"
                }],
                "usage": {"prompt_tokens": 11, "completion_tokens": 5, "total_tokens": 16},
                "timings": {"prompt_n": 11, "prompt_ms": 1.5, "predicted_n": 5, "predicted_ms": 2.5}
            }`)
		})},
	})
	require.NoError(t, err)
	adapter, err := NewAdapter(client)
	require.NoError(t, err)

	maxTokens := 64
	response, err := adapter.Chat(context.Background(), []inference.Message{{Role: "user", Content: "Inspect"}}, []inference.Tool{{
		Name:        "bash",
		Description: "Run shell",
	}}, inference.Options{MaxTokens: &maxTokens})
	require.NoError(t, err)
	assert.Equal(t, "chatcmpl-adapter", response.ID)
	assert.Equal(t, "done", response.Message.Content)
	assert.Equal(t, "bash", response.Message.ToolCalls[0].Name)
	assert.Equal(t, 16, response.Usage.TotalTokens)
	assert.Equal(t, 5, response.Timings.PredictedN)
}

func TestNewAdapterRejectsNilClient(t *testing.T) {
	adapter, err := NewAdapter(nil)
	assert.Error(t, err)
	assert.Nil(t, adapter)
}
