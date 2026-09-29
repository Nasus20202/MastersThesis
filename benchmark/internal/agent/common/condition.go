package common

import (
	"context"
	"errors"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// Condition is one benchmark condition: a tool loop started with an optional
// system prompt and a condition-specific rendering of the task.
type Condition struct {
	name         string
	loop         *Loop
	systemPrompt string
	formatTask   func(string) string
}

// NewCondition builds a condition. A blank systemPrompt sends only the user
// message; a nil formatTask sends the task unchanged.
func NewCondition(name string, client inference.Client, tools []Tool, config Config, systemPrompt string, formatTask func(string) string) (*Condition, error) {
	loop, err := NewLoop(client, tools, config)
	if err != nil {
		return nil, err
	}
	return &Condition{name: name, loop: loop, systemPrompt: systemPrompt, formatTask: formatTask}, nil
}

func (c *Condition) Run(ctx context.Context, task string) (Result, error) {
	if c == nil || c.loop == nil {
		return Result{}, errors.New("agent is not initialized")
	}
	if strings.TrimSpace(task) == "" {
		return Result{}, errors.New("agent task is required")
	}
	content := task
	if c.formatTask != nil {
		content = c.formatTask(task)
	}
	var messages []inference.Message
	if strings.TrimSpace(c.systemPrompt) != "" {
		messages = append(messages, inference.Message{Role: "system", Content: c.systemPrompt})
	}
	messages = append(messages, inference.Message{Role: "user", Content: content})
	result, err := c.loop.run(ctx, task, messages)
	result.Condition = c.name
	return result, err
}
