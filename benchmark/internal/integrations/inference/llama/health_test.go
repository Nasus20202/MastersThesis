package llama

import (
	"context"
	"net/http"
	"testing"
)

func TestReadyDistinguishesLoadingAndReady(t *testing.T) {
	t.Parallel()

	status := http.StatusServiceUnavailable
	client, err := NewClient(Config{
		BaseURL: "http://llama.test",
		Model:   "gemma-test",
		HTTPClient: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodGet || request.URL.Path != "/health" {
				t.Errorf("health request = %s %s, want GET /health", request.Method, request.URL.Path)
			}
			return testResponse(status, "")
		})},
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	ready, err := client.Ready(context.Background())
	if err != nil {
		t.Fatalf("ready while loading: %v", err)
	}
	if ready {
		t.Fatal("ready while server reports loading = true, want false")
	}

	status = http.StatusOK
	ready, err = client.Ready(context.Background())
	if err != nil {
		t.Fatalf("ready after load: %v", err)
	}
	if !ready {
		t.Fatal("ready after server reports OK = false, want true")
	}
}
