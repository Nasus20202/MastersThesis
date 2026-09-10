package llama

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

// Ready checks whether the llama-server router accepts requests. A router that
// is still starting returns (false, nil); transport failures and unexpected
// statuses are returned as errors.
func (c *Client) Ready(ctx context.Context) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(healthPath), nil)
	if err != nil {
		return false, fmt.Errorf("create llama health request: %w", err)
	}
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	started := time.Now()
	logger := slog.With("method", http.MethodGet, "path", healthPath)
	logger.DebugContext(ctx, "llama health check started")
	response, err := c.httpClient.Do(request)
	if err != nil {
		requestErr := fmt.Errorf("call llama health endpoint: %w", err)
		logger.ErrorContext(ctx, "llama health check failed",
			"duration", time.Since(started),
			"error", requestErr,
		)
		return false, requestErr
	}
	defer response.Body.Close()
	logger.DebugContext(ctx, "llama health response received",
		"status", response.StatusCode,
		"duration", time.Since(started),
	)

	switch response.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusServiceUnavailable:
		return false, nil
	default:
		httpErr := newHTTPError(response)
		logger.ErrorContext(ctx, "llama health check returned an error",
			"status", response.StatusCode,
			"error", httpErr,
		)
		return false, httpErr
	}
}
