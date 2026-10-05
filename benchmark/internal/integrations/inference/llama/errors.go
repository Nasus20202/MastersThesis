package llama

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const maxErrorBodyBytes = 8 << 10

type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("llama server returned %s", e.Status)
	}
	return fmt.Sprintf("llama server returned %s: %s", e.Status, e.Body)
}

// Unwrap maps llama-server's context-size rejection to
// inference.ErrContextOverflow.
func (e *HTTPError) Unwrap() error {
	if strings.Contains(e.Body, "exceed_context_size_error") {
		return inference.ErrContextOverflow
	}
	return nil
}

func newHTTPError(response *http.Response) *HTTPError {
	body, err := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes))
	if err != nil {
		return &HTTPError{
			StatusCode: response.StatusCode,
			Status:     response.Status,
			Body:       fmt.Sprintf("read error response: %v", err),
		}
	}
	return &HTTPError{
		StatusCode: response.StatusCode,
		Status:     response.Status,
		Body:       strings.TrimSpace(string(body)),
	}
}
