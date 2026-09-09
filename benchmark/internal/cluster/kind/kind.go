package kind

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/lifecycle"
)

const program = "kind"

type Config struct {
	Name       string
	ConfigPath string
}

type Cluster struct {
	executor   command.Executor
	name       string
	configPath string
}

// Ensure Cluster implements lifecycle.Cluster at compile time.
var _ lifecycle.Cluster = (*Cluster)(nil)

func New(executor command.Executor, config Config) (*Cluster, error) {
	if executor == nil {
		return nil, errors.New("kind cluster executor is required")
	}
	if strings.TrimSpace(config.Name) == "" {
		return nil, errors.New("kind cluster name is required")
	}
	if config.ConfigPath != "" && strings.TrimSpace(config.ConfigPath) == "" {
		return nil, errors.New("kind cluster config path is invalid")
	}
	return &Cluster{executor: executor, name: config.Name, configPath: config.ConfigPath}, nil
}

func (c *Cluster) Create(ctx context.Context) error {
	logger := slog.With("cluster", c.name)
	logger.InfoContext(ctx, "creating kind cluster")
	if err := c.run(ctx, "create", "cluster", "--name", c.name); err != nil {
		logger.ErrorContext(ctx, "kind cluster creation failed", "error", err)
		return err
	}
	logger.InfoContext(ctx, "kind cluster created", "context", c.Context())
	return nil
}

func (c *Cluster) Delete(ctx context.Context) error {
	logger := slog.With("cluster", c.name)
	logger.InfoContext(ctx, "deleting kind cluster")
	if err := c.run(ctx, "delete", "cluster", "--name", c.name); err != nil {
		logger.ErrorContext(ctx, "kind cluster deletion failed", "error", err)
		return err
	}
	logger.InfoContext(ctx, "kind cluster deleted")
	return nil
}

func (c *Cluster) Context() string {
	return "kind-" + c.name
}

func (c *Cluster) run(ctx context.Context, args ...string) error {
	spec := command.Spec{Program: program, Args: args}
	if c.configPath != "" && args[0] == "create" {
		args = append(args, "--config", c.configPath)
		spec.Args = args
		spec.Dir = filepath.Dir(c.configPath)
	}
	_, err := c.executor.Run(ctx, spec)
	if err != nil {
		return fmt.Errorf("kind cluster %q %s: %w", c.name, args[0], err)
	}
	return nil
}
