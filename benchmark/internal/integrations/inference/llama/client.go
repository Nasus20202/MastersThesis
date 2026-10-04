// Package llama is an HTTP client for llama-server.
package llama

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const (
	chatCompletionsPath = "/v1/chat/completions"
	embeddingsPath      = "/v1/embeddings"
	propsPath           = "/props"
	healthPath          = "/health"
)

type Config struct {
	BaseURL    string
	Model      string
	APIKey     string
	HTTPClient *http.Client
	Metadata   inference.Metadata
}

type Client struct {
	baseURL    *url.URL
	model      string
	apiKey     string
	httpClient *http.Client
	metadata   inference.Metadata
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
		metadata:   clientMetadata(cfg),
	}, nil
}

func clientMetadata(cfg Config) inference.Metadata {
	metadata := cfg.Metadata
	if metadata.Provider == "" {
		metadata.Provider = "llama.cpp"
	}
	if metadata.Model == "" {
		metadata.Model = cfg.Model
	}
	metadata.RuntimeSettings = maps.Clone(metadata.RuntimeSettings)
	return metadata
}

func (c *Client) Metadata() inference.Metadata {
	metadata := c.metadata
	metadata.RuntimeSettings = maps.Clone(metadata.RuntimeSettings)
	return metadata
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

// Embed requests one embedding per text from a llama-server started with
// --embeddings, returned in input order.
func (c *Client) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("at least one text to embed is required")
	}
	var response embeddingResponse
	if err := c.doJSON(ctx, http.MethodPost, embeddingsPath, embeddingRequest{Model: c.model, Input: texts}, &response); err != nil {
		return nil, err
	}
	if len(response.Data) != len(texts) {
		return nil, fmt.Errorf("llama returned %d embeddings for %d texts", len(response.Data), len(texts))
	}
	embeddings := make([][]float32, len(texts))
	for _, item := range response.Data {
		if item.Index < 0 || item.Index >= len(texts) || embeddings[item.Index] != nil {
			return nil, fmt.Errorf("llama returned an invalid embedding index %d", item.Index)
		}
		embeddings[item.Index] = item.Embedding
	}
	return embeddings, nil
}

// Sampling returns the default sampler settings llama-server applies to the
// client's model. The router loads the model if it is not loaded yet.
func (c *Client) Sampling(ctx context.Context) (inference.Sampling, error) {
	var response struct {
		DefaultGenerationSettings struct {
			Params inference.Sampling `json:"params"`
		} `json:"default_generation_settings"`
	}
	path := propsPath + "?" + url.Values{"model": {c.model}}.Encode()
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &response); err != nil {
		return inference.Sampling{}, err
	}
	return response.DefaultGenerationSettings.Params, nil
}

func (c *Client) doJSON(ctx context.Context, method, path string, payload, result any) error {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode llama request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequestWithContext(ctx, method, c.endpoint(path), body)
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
	path, endpoint.RawQuery, _ = strings.Cut(path, "?")
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + path
	endpoint.RawPath = ""
	return endpoint.String()
}
