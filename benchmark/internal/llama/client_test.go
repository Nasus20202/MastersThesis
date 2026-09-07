package llama

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestChatSendsRequestAndDecodesResponse(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{
		BaseURL: "http://llama.test",
		Model:   "gemma-test",
		APIKey:  "test-key",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", request.Method)
			}
			if request.URL.Path != "/v1/chat/completions" {
				t.Errorf("path = %s, want /v1/chat/completions", request.URL.Path)
			}
			if got := request.Header.Get("Content-Type"); got != "application/json" {
				t.Errorf("content type = %s, want application/json", got)
			}
			if got := request.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Errorf("authorization = %s, want Bearer test-key", got)
			}

			var payload struct {
				Model       string    `json:"model"`
				Messages    []Message `json:"messages"`
				MaxTokens   int       `json:"max_tokens"`
				Temperature float64   `json:"temperature"`
			}
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if payload.Model != "gemma-test" {
				t.Errorf("model = %s, want gemma-test", payload.Model)
			}
			if len(payload.Messages) != 1 || payload.Messages[0].Content != "Inspect the service" {
				t.Errorf("messages = %#v, want one inspection message", payload.Messages)
			}
			if payload.MaxTokens != 128 {
				t.Errorf("max tokens = %d, want 128", payload.MaxTokens)
			}
			if payload.Temperature != 0.2 {
				t.Errorf("temperature = %v, want 0.2", payload.Temperature)
			}

			return testResponse(http.StatusOK, `{
            "id": "chatcmpl-test",
            "choices": [{
                "index": 0,
                "message": {"role": "assistant", "content": "I would inspect the service."},
                "finish_reason": "stop"
            }],
            "usage": {"prompt_tokens": 10, "completion_tokens": 7, "total_tokens": 17},
            "timings": {"prompt_n": 10, "prompt_ms": 2.5, "predicted_n": 7, "predicted_ms": 3.5}
        }`)
		})},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	maxTokens := 128
	temperature := 0.2
	response, err := client.Chat(context.Background(), ChatRequest{
		Messages:    []Message{{Role: "user", Content: "Inspect the service"}},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	})
	if err != nil {
		t.Fatalf("chat: %v", err)
	}
	if response.ID != "chatcmpl-test" {
		t.Errorf("response ID = %s, want chatcmpl-test", response.ID)
	}
	if len(response.Choices) != 1 || response.Choices[0].Message.Content != "I would inspect the service." {
		t.Fatalf("choices = %#v, want one assistant response", response.Choices)
	}
	if response.Usage == nil || response.Usage.TotalTokens != 17 {
		t.Errorf("usage = %#v, want total tokens 17", response.Usage)
	}
	if response.Timings == nil || response.Timings.PredictedN != 7 {
		t.Errorf("timings = %#v, want predicted tokens 7", response.Timings)
	}
}

func TestNewClientAndChatValidateConfiguration(t *testing.T) {
	t.Parallel()

	for name, config := range map[string]Config{
		"missing URL":   {Model: "gemma-test"},
		"missing model": {BaseURL: "http://localhost:8080"},
		"bad scheme":    {BaseURL: "ftp://localhost:8080", Model: "gemma-test"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewClient(config); err == nil {
				t.Fatal("new client succeeded, want validation error")
			}
		})
	}

	client, err := NewClient(Config{BaseURL: "http://localhost:8080", Model: "gemma-test"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	_, err = client.Chat(context.Background(), ChatRequest{})
	if err == nil {
		t.Fatal("chat with no messages succeeded, want validation error")
	}
}
