package llama

import (
	"context"
	"fmt"
	"net/http"
)

// Ready checks whether llama-server has finished loading its model. A loading
// server returns (false, nil); transport failures and unexpected statuses are
// returned as errors.
func (c *Client) Ready(ctx context.Context) (bool, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint(healthPath), nil)
	if err != nil {
		return false, fmt.Errorf("create llama health request: %w", err)
	}
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return false, fmt.Errorf("call llama health endpoint: %w", err)
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusServiceUnavailable:
		return false, nil
	default:
		return false, newHTTPError(response)
	}
}
