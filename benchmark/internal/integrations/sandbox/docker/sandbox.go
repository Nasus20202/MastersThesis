package docker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
)

const (
	dockerProgram = "docker"
	tmpfsMode     = ":rw,mode=1777"
)

type ImageLayout struct {
	User                string
	Workdir             string
	KubeconfigMountPath string
}

func DefaultImageLayout() ImageLayout {
	return ImageLayout{
		User:                "benchmark",
		Workdir:             "/workspace",
		KubeconfigMountPath: "/home/benchmark/.kube/config",
	}
}

type Config struct {
	Name           string
	Image          string
	DockerfilePath string
	BuildContext   string
	KubeconfigPath string
	Network        string
	NetworkTarget  string
	Layout         ImageLayout
}

type Sandbox struct {
	executor command.Executor
	config   Config
}

func New(executor command.Executor, config Config) (*Sandbox, error) {
	if executor == nil {
		return nil, errors.New("sandbox executor is required")
	}
	if strings.TrimSpace(config.Name) == "" {
		return nil, errors.New("sandbox name is required")
	}
	if strings.TrimSpace(config.Image) == "" {
		return nil, errors.New("sandbox image is required")
	}
	if strings.TrimSpace(config.DockerfilePath) == "" {
		return nil, errors.New("sandbox Dockerfile path is required")
	}
	if strings.TrimSpace(config.BuildContext) == "" {
		return nil, errors.New("sandbox build context is required")
	}
	if strings.TrimSpace(config.KubeconfigPath) == "" {
		return nil, errors.New("sandbox kubeconfig path is required")
	}
	if !filepath.IsAbs(config.KubeconfigPath) {
		return nil, errors.New("sandbox kubeconfig path must be absolute")
	}
	if strings.TrimSpace(config.Network) == "" {
		return nil, errors.New("sandbox network is required")
	}
	if strings.TrimSpace(config.NetworkTarget) == "" {
		return nil, errors.New("sandbox network target is required")
	}
	if strings.TrimSpace(config.Layout.User) == "" {
		return nil, errors.New("sandbox image user is required")
	}
	if strings.TrimSpace(config.Layout.Workdir) == "" {
		return nil, errors.New("sandbox image workdir is required")
	}
	if !filepath.IsAbs(config.Layout.Workdir) {
		return nil, errors.New("sandbox image workdir must be absolute")
	}
	if strings.TrimSpace(config.Layout.KubeconfigMountPath) == "" {
		return nil, errors.New("sandbox kubeconfig mount path is required")
	}
	if !filepath.IsAbs(config.Layout.KubeconfigMountPath) {
		return nil, errors.New("sandbox kubeconfig mount path must be absolute")
	}
	return &Sandbox{executor: executor, config: config}, nil
}

func (s *Sandbox) Build(ctx context.Context) error {
	logger := slog.With("sandbox_image", s.config.Image)
	logger.InfoContext(ctx, "building sandbox image",
		"dockerfile", s.config.DockerfilePath,
		"context", s.config.BuildContext,
	)
	_, err := s.executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args: []string{
			"build",
			"--file", s.config.DockerfilePath,
			"--tag", s.config.Image,
			s.config.BuildContext,
		},
	})
	if err != nil {
		logger.ErrorContext(ctx, "sandbox image build failed", "error", err)
		return fmt.Errorf("build sandbox image %q: %w", s.config.Image, err)
	}
	logger.InfoContext(ctx, "sandbox image built")
	return nil
}

func (s *Sandbox) Start(ctx context.Context) error {
	logger := slog.With("sandbox", s.config.Name)
	logger.InfoContext(ctx, "starting sandbox")
	if _, err := s.executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args:    []string{"network", "create", "--internal", s.config.Network},
	}); err != nil {
		logger.ErrorContext(ctx, "sandbox network creation failed", "error", err)
		return fmt.Errorf("create sandbox network %q: %w", s.config.Network, err)
	}
	if _, err := s.executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args:    []string{"network", "connect", s.config.Network, s.config.NetworkTarget},
	}); err != nil {
		_ = s.removeNetwork(ctx)
		logger.ErrorContext(ctx, "sandbox network connection failed", "error", err)
		return fmt.Errorf("connect sandbox network %q to %q: %w", s.config.Network, s.config.NetworkTarget, err)
	}
	if _, err := s.executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args: []string{
			"run",
			"--detach",
			"--rm",
			"--name", s.config.Name,
			"--hostname", s.config.Name,
			"--network", s.config.Network,
			"--user", s.config.Layout.User,
			"--workdir", s.config.Layout.Workdir,
			"--cap-drop", "ALL",
			"--security-opt", "no-new-privileges",
			"--read-only",
			"--tmpfs", "/tmp" + tmpfsMode,
			"--tmpfs", s.config.Layout.Workdir + tmpfsMode,
			"--mount", fmt.Sprintf("type=bind,src=%s,dst=%s,readonly", filepath.Clean(s.config.KubeconfigPath), s.config.Layout.KubeconfigMountPath),
			s.config.Image,
			"sleep", "infinity",
		},
	}); err != nil {
		_ = s.disconnectNetwork(ctx)
		_ = s.removeNetwork(ctx)
		logger.ErrorContext(ctx, "sandbox start failed", "error", err)
		return fmt.Errorf("start sandbox %q: %w", s.config.Name, err)
	}
	if _, err := s.Exec(ctx, command.Spec{Program: "kubectl", Args: []string{"get", "nodes"}}); err != nil {
		cleanupErr := s.Stop(context.Background())
		if cleanupErr != nil {
			return errors.Join(fmt.Errorf("validate sandbox Kubernetes access: %w", err), cleanupErr)
		}
		return fmt.Errorf("validate sandbox Kubernetes access: %w", err)
	}
	logger.InfoContext(ctx, "sandbox started")
	return nil
}

func (s *Sandbox) Exec(ctx context.Context, spec command.Spec) (command.Result, error) {
	if err := spec.Validate(); err != nil {
		return command.Result{}, err
	}

	args := []string{
		"exec",
		"--user", s.config.Layout.User,
		"--workdir", workdir(spec.Dir, s.config.Layout.Workdir),
	}
	for _, key := range slices.Sorted(maps.Keys(spec.Env)) {
		args = append(args, "--env", key+"="+spec.Env[key])
	}
	args = append(args, s.config.Name, spec.Program)
	args = append(args, spec.Args...)

	result, err := s.executor.Run(ctx, command.Spec{Program: dockerProgram, Args: args})
	if err != nil {
		return result, fmt.Errorf("execute sandbox command: %w", err)
	}
	return result, nil
}

func (s *Sandbox) Stop(ctx context.Context) error {
	logger := slog.With("sandbox", s.config.Name)
	logger.InfoContext(ctx, "stopping sandbox")
	var stopErr error
	_, err := s.executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args:    []string{"stop", s.config.Name},
	})
	if err != nil {
		logger.ErrorContext(ctx, "sandbox stop failed", "error", err)
		stopErr = fmt.Errorf("stop sandbox %q: %w", s.config.Name, err)
	}
	var cleanupErrs []error
	if stopErr != nil {
		cleanupErrs = append(cleanupErrs, stopErr)
	}
	if err := s.disconnectNetwork(ctx); err != nil {
		logger.ErrorContext(ctx, "sandbox network disconnection failed", "error", err)
		cleanupErrs = append(cleanupErrs, err)
	}
	if err := s.removeNetwork(ctx); err != nil {
		logger.ErrorContext(ctx, "sandbox network removal failed", "error", err)
		cleanupErrs = append(cleanupErrs, err)
	}
	if err := errors.Join(cleanupErrs...); err != nil {
		return err
	}
	logger.InfoContext(ctx, "sandbox stopped")
	return nil
}

func (s *Sandbox) disconnectNetwork(ctx context.Context) error {
	_, err := s.executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args:    []string{"network", "disconnect", "--force", s.config.Network, s.config.NetworkTarget},
	})
	if err != nil {
		return fmt.Errorf("disconnect sandbox network %q from %q: %w", s.config.Network, s.config.NetworkTarget, err)
	}
	return nil
}

func (s *Sandbox) removeNetwork(ctx context.Context) error {
	_, err := s.executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args:    []string{"network", "rm", s.config.Network},
	})
	if err != nil {
		return fmt.Errorf("remove sandbox network %q: %w", s.config.Network, err)
	}
	return nil
}

func workdir(dir, defaultDir string) string {
	if dir == "" {
		return defaultDir
	}
	return dir
}
