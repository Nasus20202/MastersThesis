package llama

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

func TestChatReturnsHTTPError(t *testing.T) {
	t.Parallel()

	client, err := NewClient(Config{
		BaseURL: "http://llama.test",
		Model:   "gemma-test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return testResponse(http.StatusServiceUnavailable, "model is not loaded\n")
		})},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	_, err = client.Chat(context.Background(), ChatRequest{
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error = %v, want HTTPError", err)
	}
	if httpErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status code = %d, want 503", httpErr.StatusCode)
	}
	if !strings.Contains(httpErr.Body, "model is not loaded") {
		t.Errorf("error body = %q, want model-not-loaded message", httpErr.Body)
	}
}

func TestHTTPErrorMapsContextSizeRejection(t *testing.T) {
	t.Parallel()

	overflow := &HTTPError{StatusCode: http.StatusBadRequest, Body: `{"error":{"code":400,"message":"request (40012 tokens) exceeds the available context size (32768 tokens), try increasing it","type":"exceed_context_size_error","n_prompt_tokens":40012,"n_ctx":32768}}`}
	if !errors.Is(overflow, inference.ErrContextOverflow) {
		t.Errorf("errors.Is(%v, ErrContextOverflow) = false, want true", overflow)
	}
	other := &HTTPError{StatusCode: http.StatusServiceUnavailable, Body: "model is not loaded"}
	if errors.Is(other, inference.ErrContextOverflow) {
		t.Errorf("errors.Is(%v, ErrContextOverflow) = true, want false", other)
	}
}
