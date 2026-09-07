package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/llama"
)

type readinessClient struct {
	results []readinessResult
	calls   int
}

type readinessResult struct {
	ready bool
	err   error
}

func (c *readinessClient) Ready(context.Context) (bool, error) {
	result := c.results[c.calls]
	c.calls++
	return result.ready, result.err
}

func (*readinessClient) Chat(context.Context, llama.ChatRequest) (llama.ChatResponse, error) {
	return llama.ChatResponse{}, nil
}

func TestWaitUntilReadyRetriesTransientFailures(t *testing.T) {
	t.Parallel()

	client := &readinessClient{results: []readinessResult{
		{err: errors.New("connection refused")},
		{ready: false},
		{ready: true},
	}}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := waitUntilReady(ctx, client, time.Millisecond); err != nil {
		t.Fatalf("wait until ready: %v", err)
	}
	if client.calls != 3 {
		t.Fatalf("ready calls = %d, want 3", client.calls)
	}
}

func TestWaitUntilReadyHonorsContextDeadline(t *testing.T) {
	t.Parallel()

	client := &readinessClient{results: []readinessResult{{ready: false}}}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := waitUntilReady(ctx, client, time.Hour); !errors.Is(err, context.Canceled) {
		t.Fatalf("wait error = %v, want context canceled", err)
	}
}

func TestEnvIntOrDefault(t *testing.T) {
	t.Setenv("TEST_MAX_TOKENS", "64")
	value, err := envIntOrDefault("TEST_MAX_TOKENS", 32)
	if err != nil {
		t.Fatalf("parse environment integer: %v", err)
	}
	if value != 64 {
		t.Fatalf("value = %d, want 64", value)
	}

	t.Setenv("TEST_MAX_TOKENS", "invalid")
	if _, err := envIntOrDefault("TEST_MAX_TOKENS", 32); err == nil {
		t.Fatal("invalid environment integer succeeded")
	}
}

func TestRequiredEnv(t *testing.T) {
	t.Setenv("TEST_REQUIRED", " configured ")
	value, err := requiredEnv("TEST_REQUIRED")
	if err != nil {
		t.Fatalf("read required environment variable: %v", err)
	}
	if value != "configured" {
		t.Fatalf("value = %q, want configured", value)
	}

	t.Setenv("TEST_REQUIRED", "")
	if _, err := requiredEnv("TEST_REQUIRED"); err == nil {
		t.Fatal("missing required environment variable succeeded")
	}
}
