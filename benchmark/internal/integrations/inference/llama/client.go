// Package llama provides the small HTTP boundary used to communicate with
// llama-server.
package llama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	chatCompletionsPath = "/v1/chat/completions"
	healthPath          = "/health"
)

// Config contains the connection details for a llama-server instance.
type Config struct {
	BaseURL    string
	Model      string
	APIKey     string
	HTTPClient *http.Client
}

// Client is an HTTP client for the llama-server API.
type Client struct {
	baseURL    *url.URL
	model      string
	apiKey     string
	httpClient *http.Client
}

// NewClient validates cfg and returns a client for the configured server.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("llama server base URL is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, errors.New("llama model is required")
	}

	baseURL, err := url.Parse(strings.TrimRight(cfg.BaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse llama server base URL: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("llama server base URL must use http or https, got %q", baseURL.Scheme)
	}
	if baseURL.Host == "" {
		return nil, errors.New("llama server base URL must include a host")
	}
	if baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, errors.New("llama server base URL must not include a query or fragment")
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	return &Client{
		baseURL:    baseURL,
		model:      cfg.Model,
		apiKey:     cfg.APIKey,
		httpClient: httpClient,
	}, nil
}

// Chat sends a non-streaming chat completion request to llama-server.
func (c *Client) Chat(ctx context.Context, request ChatRequest) (ChatResponse, error) {
	if len(request.Messages) == 0 {
		return ChatResponse{}, errors.New("at least one chat message is required")
	}

	payload := struct {
		Model string `json:"model"`
		ChatRequest
	}{
		Model:       c.model,
		ChatRequest: request,
	}

	var response ChatResponse
	if err := c.doJSON(ctx, http.MethodPost, chatCompletionsPath, payload, &response); err != nil {
		return ChatResponse{}, err
	}
	return response, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode llama request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create llama request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	started := time.Now()
	logger := slog.With("method", method, "path", path)
	logger.DebugContext(ctx, "llama request started")
	response, err := c.httpClient.Do(request)
	if err != nil {
		requestErr := fmt.Errorf("call llama endpoint: %w", err)
		logger.ErrorContext(ctx, "llama request failed",
			"duration", time.Since(started),
			"error", requestErr,
		)
		return requestErr
	}
	defer response.Body.Close()
	logger.DebugContext(ctx, "llama response received",
		"status", response.StatusCode,
		"duration", time.Since(started),
	)

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		httpErr := newHTTPError(response)
		logger.ErrorContext(ctx, "llama request returned an error",
			"status", response.StatusCode,
			"error", httpErr,
		)
		return httpErr
	}
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		decodeErr := fmt.Errorf("decode llama response: %w", err)
		logger.ErrorContext(ctx, "llama response decode failed",
			"error", decodeErr,
		)
		return decodeErr
	}
	return nil
}

func (c *Client) endpoint(path string) string {
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path
	endpoint.RawPath = ""
	return endpoint.String()
}
