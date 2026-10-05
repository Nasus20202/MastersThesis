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

func TestClientExposesInferenceMetadata(t *testing.T) {
	client, err := NewClient(Config{
		BaseURL: "http://localhost:8080",
		Model:   "gemma-test",
		Metadata: inference.Metadata{
			Artifact: "repo@revision/model.gguf",
			RuntimeSettings: map[string]string{
				"LLAMA_CONTEXT_SIZE": "32768",
			},
		},
	})
	require.NoError(t, err)

	metadata := client.Metadata()
	assert.Equal(t, "llama.cpp", metadata.Provider)
	assert.Equal(t, "gemma-test", metadata.Model)
	assert.Equal(t, "repo@revision/model.gguf", metadata.Artifact)
	assert.Equal(t, map[string]string{"LLAMA_CONTEXT_SIZE": "32768"}, metadata.RuntimeSettings)

	metadata.RuntimeSettings["LLAMA_CONTEXT_SIZE"] = "1"
	assert.Equal(t, "32768", client.Metadata().RuntimeSettings["LLAMA_CONTEXT_SIZE"])
}

func TestEmbedReturnsVectorsInInputOrder(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{
		BaseURL: "http://llama.test",
		Model:   "embedding-test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			assert.Equal(t, "/v1/embeddings", request.URL.Path)
			var payload embeddingRequest
			require.NoError(t, json.NewDecoder(request.Body).Decode(&payload))
			assert.Equal(t, embeddingRequest{Model: "embedding-test", Input: []string{"first", "second"}}, payload)
			return testResponse(http.StatusOK, `{"data": [
                {"index": 1, "embedding": [0, 1]},
                {"index": 0, "embedding": [1, 0]}
            ]}`)
		})},
	})
	require.NoError(t, err)

	embeddings, err := client.Embed(context.Background(), []string{"first", "second"})
	require.NoError(t, err)
	assert.Equal(t, [][]float32{{1, 0}, {0, 1}}, embeddings)
}

func TestEmbedRejectsIncompleteResponse(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{
		BaseURL: "http://llama.test",
		Model:   "embedding-test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return testResponse(http.StatusOK, `{"data": [{"index": 1, "embedding": [0, 1]}, {"index": 1, "embedding": [0, 1]}]}`)
		})},
	})
	require.NoError(t, err)

	_, err = client.Embed(context.Background(), []string{"first", "second"})
	assert.ErrorContains(t, err, "invalid embedding index 1")
}

func TestSamplingReadsModelProps(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{
		BaseURL: "http://llama.test",
		Model:   "Qwen3.5-9B-Q4_K_M",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			assert.Equal(t, http.MethodGet, request.Method)
			assert.Equal(t, "/props", request.URL.Path)
			assert.Equal(t, "Qwen3.5-9B-Q4_K_M", request.URL.Query().Get("model"))
			return testResponse(http.StatusOK, `{"default_generation_settings": {"params": {
                "temperature": 1.0, "top_k": 20, "top_p": 0.95, "min_p": 0.0,
                "presence_penalty": 1.5, "frequency_penalty": 0.0, "repeat_penalty": 1.0, "seed": 4294967295
            }}}`)
		})},
	})
	require.NoError(t, err)

	sampling, err := client.Sampling(context.Background())
	require.NoError(t, err)
	assert.Equal(t, inference.Sampling{Temperature: 1, TopK: 20, TopP: 0.95, PresencePenalty: 1.5, RepeatPenalty: 1}, sampling)
}
