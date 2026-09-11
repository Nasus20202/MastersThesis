package kind

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	clusterintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/cluster"
)

const (
	program           = "kind"
	kindContextPrefix = "kind-"
)

type Config struct {
	Name       string
	ConfigPath string
}

type Cluster struct {
	executor               command.Executor
	name                   string
	configPath             string
	kubeconfigPath         string
	internalKubeconfigPath string
}

// Ensure Cluster implements the provider-neutral cluster contract.
var _ clusterintegration.Cluster = (*Cluster)(nil)

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
	return &Cluster{
		executor:               executor,
		name:                   config.Name,
		configPath:             config.ConfigPath,
		kubeconfigPath:         filepath.Join(os.TempDir(), config.Name+".kubeconfig"),
		internalKubeconfigPath: filepath.Join(os.TempDir(), config.Name+".internal.kubeconfig"),
	}, nil
}

func (c *Cluster) Create(ctx context.Context) error {
	logger := slog.With("cluster", c.name)
	logger.InfoContext(ctx, "creating kind cluster")
	if _, err := c.run(ctx, "create", "cluster", "--name", c.name); err != nil {
		logger.ErrorContext(ctx, "kind cluster creation failed", "error", err)
		return err
	}
	if err := c.generateInternalKubeconfig(ctx); err != nil {
		logger.ErrorContext(ctx, "kind internal kubeconfig generation failed", "error", err)
		return err
	}
	logger.InfoContext(ctx, "kind cluster created")
	return nil
}

func (c *Cluster) Delete(ctx context.Context) error {
	logger := slog.With("cluster", c.name)
	defer func() {
		for _, path := range []string{c.kubeconfigPath, c.internalKubeconfigPath} {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				logger.WarnContext(ctx, "kind kubeconfig cleanup failed", "path", path, "error", err)
			}
		}
	}()
	logger.InfoContext(ctx, "deleting kind cluster")
	if _, err := c.run(ctx, "delete", "cluster", "--name", c.name); err != nil {
		logger.ErrorContext(ctx, "kind cluster deletion failed", "error", err)
		return err
	}
	logger.InfoContext(ctx, "kind cluster deleted")
	return nil
}

func (c *Cluster) KubeconfigPath() string {
	return c.kubeconfigPath
}

func (c *Cluster) InternalKubeconfigPath() string {
	return c.internalKubeconfigPath
}

func (c *Cluster) generateInternalKubeconfig(ctx context.Context) error {
	result, err := c.run(ctx, "get", "kubeconfig", "--name", c.name, "--internal")
	if err != nil {
		return fmt.Errorf("get internal kubeconfig for cluster %q: %w", c.name, err)
	}
	if strings.TrimSpace(result.Stdout) == "" {
		return fmt.Errorf("get internal kubeconfig for cluster %q: command returned empty output", c.name)
	}
	// The container's benchmark UID is intentionally independent of the host UID.
	if err := os.WriteFile(c.internalKubeconfigPath, []byte(result.Stdout), 0o644); err != nil {
		return fmt.Errorf("write internal kubeconfig for cluster %q: %w", c.name, err)
	}
	return nil
}

func (c *Cluster) KubeconfigContext() string {
	return kindContextPrefix + c.name
}

func (c *Cluster) run(ctx context.Context, args ...string) (command.Result, error) {
	spec := command.Spec{Program: program, Args: args}
	if c.configPath != "" && args[0] == "create" {
		args = append(args, "--kubeconfig", c.kubeconfigPath)
		args = append(args, "--config", c.configPath)
		spec.Args = args
		spec.Dir = filepath.Dir(c.configPath)
	} else if args[0] == "create" {
		args = append(args, "--kubeconfig", c.kubeconfigPath)
		spec.Args = args
	}
	result, err := c.executor.Run(ctx, spec)
	if err != nil {
		stderr := strings.TrimSpace(result.Stderr)
		if stderr != "" {
			return result, fmt.Errorf("kind cluster %q %s: %w: %s", c.name, args[0], err, stderr)
		}
		return result, fmt.Errorf("kind cluster %q %s: %w", c.name, args[0], err)
	}
	return result, nil
}
