package llama

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxErrorBodyBytes = 8 << 10

// HTTPError reports a non-successful response from llama-server.
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
