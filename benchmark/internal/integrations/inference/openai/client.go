// Package openai is a chat client for OpenAI-compatible APIs.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const (
	chatCompletionsPath = "/chat/completions"
	maxErrorBodyBytes   = 8 << 10
)

type Config struct {
	// BaseURL includes the API version prefix, e.g. https://api.openai.com/v1.
	BaseURL    string
	Model      string
	APIKey     string
	HTTPClient *http.Client
}

// Client implements inference.Client and inference.MetadataProvider.
type Client struct {
	baseURL    *url.URL
	model      string
	apiKey     string
	httpClient *http.Client
}

// NewClient validates cfg and returns a client for the configured API.
func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return nil, errors.New("openai base URL is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, errors.New("openai model is required")
	}
	baseURL, err := url.Parse(strings.TrimRight(cfg.BaseURL, "/"))
	if err != nil {
		return nil, fmt.Errorf("parse openai base URL: %w", err)
	}
	if baseURL.Scheme != "http" && baseURL.Scheme != "https" {
		return nil, fmt.Errorf("openai base URL must use http or https, got %q", baseURL.Scheme)
	}
	if baseURL.Host == "" {
		return nil, errors.New("openai base URL must include a host")
	}
	if baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return nil, errors.New("openai base URL must not include a query or fragment")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Client{baseURL: baseURL, model: cfg.Model, apiKey: cfg.APIKey, httpClient: httpClient}, nil
}

func (c *Client) Metadata() inference.Metadata {
	return inference.Metadata{Provider: "openai", Model: c.model}
}

// Chat sends a non-streaming chat completion request.
func (c *Client) Chat(ctx context.Context, messages []inference.Message, tools []inference.Tool, options inference.Options) (inference.Result, error) {
	if len(messages) == 0 {
		return inference.Result{}, errors.New("at least one chat message is required")
	}
	var response chatResponse
	err := c.post(ctx, chatCompletionsPath, chatRequest{
		Model:               c.model,
		Messages:            toMessages(messages),
		Tools:               toTools(tools),
		Temperature:         options.Temperature,
		MaxCompletionTokens: options.MaxTokens,
		Seed:                options.Seed,
	}, &response)
	if err != nil {
		return inference.Result{}, err
	}
	if len(response.Choices) == 0 {
		return inference.Result{}, errors.New("openai response contained no choices")
	}
	choice := response.Choices[0]
	return inference.Result{
		ID:           response.ID,
		Message:      fromMessage(choice.Message),
		FinishReason: choice.FinishReason,
		Usage:        fromUsage(response.Usage),
	}, nil
}

func (c *Client) post(ctx context.Context, path string, payload, result any) error {
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode openai request: %w", err)
	}
	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path
	endpoint.RawPath = ""
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("create openai request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	started := time.Now()
	logger := slog.With("path", path)
	logger.DebugContext(ctx, "openai request started")
	response, err := c.httpClient.Do(request)
	if err != nil {
		requestErr := fmt.Errorf("call openai endpoint: %w", err)
		logger.ErrorContext(ctx, "openai request failed", "duration", time.Since(started), "error", requestErr)
		return requestErr
	}
	defer response.Body.Close()
	logger.DebugContext(ctx, "openai response received", "status", response.StatusCode, "duration", time.Since(started))

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		httpErr := newHTTPError(response)
		logger.ErrorContext(ctx, "openai request returned an error", "status", response.StatusCode, "error", httpErr)
		return httpErr
	}
	if err := json.NewDecoder(response.Body).Decode(result); err != nil {
		return fmt.Errorf("decode openai response: %w", err)
	}
	return nil
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("openai API returned %s", e.Status)
	}
	return fmt.Sprintf("openai API returned %s: %s", e.Status, e.Body)
}

// Unwrap maps the context-size rejection to inference.ErrContextOverflow.
func (e *HTTPError) Unwrap() error {
	if strings.Contains(e.Body, "context_length_exceeded") {
		return inference.ErrContextOverflow
	}
	return nil
}

func newHTTPError(response *http.Response) *HTTPError {
	body, err := io.ReadAll(io.LimitReader(response.Body, maxErrorBodyBytes))
	text := strings.TrimSpace(string(body))
	if err != nil {
		text = fmt.Sprintf("read error response: %v", err)
	}
	return &HTTPError{StatusCode: response.StatusCode, Status: response.Status, Body: text}
}

var _ inference.Client = (*Client)(nil)
