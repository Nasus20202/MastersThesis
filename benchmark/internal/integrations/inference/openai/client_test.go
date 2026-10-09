package openai

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func testResponse(statusCode int, body string) (*http.Response, error) {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestChatSendsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{
		BaseURL: "https://api.test/v1/",
		Model:   "gpt-test",
		APIKey:  "test-key",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			assert.Equal(t, "https://api.test/v1/chat/completions", request.URL.String())
			assert.Equal(t, "Bearer test-key", request.Header.Get("Authorization"))

			var payload map[string]any
			require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
			assert.Equal(t, "gpt-test", payload["model"])
			assert.EqualValues(t, 128, payload["max_completion_tokens"])
			assert.NotContains(t, payload, "max_tokens")
			messages := payload["messages"].([]any)
			assistant := messages[1].(map[string]any)
			assert.NotContains(t, assistant, "reasoning_content")
			assert.Equal(t, "call-1", assistant["tool_calls"].([]any)[0].(map[string]any)["id"])
			tool := payload["tools"].([]any)[0].(map[string]any)
			assert.Equal(t, "function", tool["type"])

			return testResponse(http.StatusOK, `{
				"id": "chatcmpl-test",
				"choices": [{"message": {"role": "assistant", "content": "", "reasoning_content": "thinking",
					"tool_calls": [{"id": "call-2", "type": "function", "function": {"name": "shell", "arguments": "{}"}}]},
					"finish_reason": "tool_calls"}],
				"usage": {"prompt_tokens": 10, "completion_tokens": 7, "total_tokens": 17,
					"prompt_tokens_details": {"cached_tokens": 4}}
			}`)
		})},
	})
	require.NoError(t, err)

	maxTokens := 128
	result, err := client.Chat(context.Background(),
		[]inference.Message{
			{Role: "user", Content: "Inspect"},
			{Role: "assistant", ReasoningContent: "secret", ToolCalls: []inference.ToolCall{{ID: "call-1", Type: "function", Name: "shell", Arguments: "{}"}}},
		},
		[]inference.Tool{{Name: "shell", Parameters: json.RawMessage(`{"type":"object"}`)}},
		inference.Options{MaxTokens: &maxTokens},
	)
	require.NoError(t, err)
	assert.Equal(t, "chatcmpl-test", result.ID)
	assert.Equal(t, "tool_calls", result.FinishReason)
	assert.Equal(t, "thinking", result.Message.ReasoningContent)
	assert.Equal(t, []inference.ToolCall{{ID: "call-2", Type: "function", Name: "shell", Arguments: "{}"}}, result.Message.ToolCalls)
	assert.Equal(t, &inference.Usage{PromptTokens: 10, CompletionTokens: 7, TotalTokens: 17, CachedTokens: 4}, result.Usage)
	assert.Nil(t, result.Timings)
}

func TestNewClientValidatesConfiguration(t *testing.T) {
	t.Parallel()

	for name, config := range map[string]Config{
		"missing URL":   {Model: "gpt-test"},
		"missing model": {BaseURL: "https://api.test/v1"},
		"bad scheme":    {BaseURL: "ftp://api.test", Model: "gpt-test"},
		"query":         {BaseURL: "https://api.test/v1?x=1", Model: "gpt-test"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewClient(config)
			assert.Error(t, err)
		})
	}

	client, err := NewClient(Config{BaseURL: "https://api.test/v1", Model: "gpt-test"})
	require.NoError(t, err)
	_, err = client.Chat(context.Background(), nil, nil, inference.Options{})
	assert.Error(t, err)
}

func TestMetadataDefaults(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{BaseURL: "https://api.test/v1", Model: "gpt-test"})
	require.NoError(t, err)
	assert.Equal(t, inference.Metadata{Provider: "openai", Model: "gpt-test"}, client.Metadata())
}

func TestChatMapsErrors(t *testing.T) {
	t.Parallel()

	chat := func(status int, body string) error {
		client, err := NewClient(Config{
			BaseURL: "https://api.test/v1",
			Model:   "gpt-test",
			HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				return testResponse(status, body)
			})},
		})
		require.NoError(t, err)
		_, err = client.Chat(context.Background(), []inference.Message{{Role: "user", Content: "hi"}}, nil, inference.Options{})
		return err
	}

	err := chat(http.StatusTooManyRequests, "rate limited\n")
	var httpErr *HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, http.StatusTooManyRequests, httpErr.StatusCode)
	assert.Equal(t, "rate limited", httpErr.Body)
	assert.False(t, errors.Is(err, inference.ErrContextOverflow))

	err = chat(http.StatusBadRequest, `{"error":{"code":"context_length_exceeded"}}`)
	assert.ErrorIs(t, err, inference.ErrContextOverflow)

	assert.ErrorContains(t, chat(http.StatusOK, `{"choices":[]}`), "no choices")
}
