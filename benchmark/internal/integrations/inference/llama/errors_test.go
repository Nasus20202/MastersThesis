package llama

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
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
