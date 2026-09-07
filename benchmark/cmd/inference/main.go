package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/llama"
)

const (
	defaultHost      = "localhost"
	defaultPort      = "8080"
	defaultModel     = "gemma-4-e4b"
	defaultPrompt    = "Reply with exactly: inference successful"
	defaultMaxTokens = 32
	defaultTimeout   = 5 * time.Minute
	retryInterval    = time.Second
)

type llamaClient interface {
	Ready(context.Context) (bool, error)
	Chat(context.Context, llama.ChatRequest) (llama.ChatResponse, error)
}

func main() {
	if err := run(context.Background(), os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "inference failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, output io.Writer) error {
	requestCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	client, err := llama.NewClient(llama.Config{
		BaseURL: "http://" + envOrDefault("LLAMA_CLIENT_HOST", defaultHost) + ":" + envOrDefault("LLAMA_PORT", defaultPort),
		Model:   envOrDefault("LLAMA_MODEL_NAME", defaultModel),
	})
	if err != nil {
		return err
	}

	if err := waitUntilReady(requestCtx, client, retryInterval); err != nil {
		return err
	}

	maxTokens, err := envIntOrDefault("LLAMA_INFERENCE_MAX_TOKENS", defaultMaxTokens)
	if err != nil {
		return err
	}
	temperature := 0.0
	response, err := client.Chat(requestCtx, llama.ChatRequest{
		Messages:    []llama.Message{{Role: "user", Content: envOrDefault("LLAMA_INFERENCE_PROMPT", defaultPrompt)}},
		MaxTokens:   &maxTokens,
		Temperature: &temperature,
	})
	if err != nil {
		return err
	}
	if len(response.Choices) == 0 || strings.TrimSpace(response.Choices[0].Message.Content) == "" {
		return errors.New("llama server returned no generated text")
	}

	fmt.Fprintln(output, response.Choices[0].Message.Content)
	if response.Usage != nil {
		fmt.Fprintf(output, "tokens: prompt=%d completion=%d total=%d\n", response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)
	}
	return nil
}

func waitUntilReady(ctx context.Context, client llamaClient, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		ready, err := client.Ready(ctx)
		if err == nil && ready {
			return nil
		}

		select {
		case <-ctx.Done():
			if err != nil {
				return fmt.Errorf("wait for llama server: %w", err)
			}
			return fmt.Errorf("wait for llama server: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

func envOrDefault(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}

func envIntOrDefault(name string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}
