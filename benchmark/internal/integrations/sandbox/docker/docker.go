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
	"sync"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
)

const (
	dockerProgram = "docker"
	tmpfsMode     = ":rw,mode=1777"
)

type Mount struct {
	Source   string
	Target   string
	ReadOnly bool
}

// SecurityConfig controls the container hardening flags. The evaluator setup
// container leaves all of them disabled.
type SecurityConfig struct {
	ReadOnlyRoot     bool
	DropCapabilities bool
	NoNewPrivileges  bool
}

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
	Env            map[string]string
	Mounts         []Mount
	Security       SecurityConfig
	Command        []string
}

type ImageConfig struct {
	Image          string
	DockerfilePath string
	BuildContext   string
}

type ImageBuilder struct {
	executor command.Executor
	config   ImageConfig
	once     sync.Once
	err      error
}

type Sandbox struct {
	executor command.Executor
	config   Config
}

var _ sandboxintegration.Sandbox = (*Sandbox)(nil)
var _ sandboxintegration.Executor = (*Sandbox)(nil)

func NewImageBuilder(executor command.Executor, config ImageConfig) (*ImageBuilder, error) {
	if executor == nil {
		return nil, errors.New("sandbox image executor is required")
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
	return &ImageBuilder{executor: executor, config: config}, nil
}

func (b *ImageBuilder) Build(ctx context.Context) error {
	b.once.Do(func() {
		b.err = ensureImage(ctx, b.executor, b.config)
	})
	return b.err
}

func ensureImage(ctx context.Context, executor command.Executor, config ImageConfig) error {
	logger := slog.With("sandbox_image", config.Image)
	if _, err := executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args:    []string{"image", "inspect", config.Image},
	}); err == nil {
		logger.InfoContext(ctx, "sandbox image already available")
		return nil
	} else if ctx.Err() != nil {
		return fmt.Errorf("check sandbox image %q: %w", config.Image, ctx.Err())
	}
	return buildImage(ctx, executor, config)
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
	if strings.TrimSpace(config.KubeconfigPath) != "" && !filepath.IsAbs(config.KubeconfigPath) {
		return nil, errors.New("sandbox kubeconfig path must be absolute")
	}
	if strings.TrimSpace(config.Network) == "" {
		return nil, errors.New("sandbox network is required")
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
	if strings.TrimSpace(config.KubeconfigPath) != "" && strings.TrimSpace(config.Layout.KubeconfigMountPath) == "" {
		return nil, errors.New("sandbox kubeconfig mount path is required")
	}
	if strings.TrimSpace(config.KubeconfigPath) != "" && !filepath.IsAbs(config.Layout.KubeconfigMountPath) {
		return nil, errors.New("sandbox kubeconfig mount path must be absolute")
	}
	for _, mount := range config.Mounts {
		if strings.TrimSpace(mount.Source) == "" || strings.TrimSpace(mount.Target) == "" {
			return nil, errors.New("sandbox mount source and target are required")
		}
		if !filepath.IsAbs(mount.Source) || !filepath.IsAbs(mount.Target) {
			return nil, errors.New("sandbox mount source and target must be absolute")
		}
	}
	return &Sandbox{executor: executor, config: config}, nil
}

func (s *Sandbox) Build(ctx context.Context) error {
	return buildImage(ctx, s.executor, ImageConfig{
		Image:          s.config.Image,
		DockerfilePath: s.config.DockerfilePath,
		BuildContext:   s.config.BuildContext,
	})
}

func buildImage(ctx context.Context, executor command.Executor, config ImageConfig) error {
	logger := slog.With("sandbox_image", config.Image)
	logger.InfoContext(ctx, "building sandbox image",
		"dockerfile", config.DockerfilePath,
		"context", config.BuildContext,
	)
	_, err := executor.Run(ctx, command.Spec{
		Program: dockerProgram,
		Args: []string{
			"build",
			"--file", config.DockerfilePath,
			"--tag", config.Image,
			config.BuildContext,
		},
	})
	if err != nil {
		logger.ErrorContext(ctx, "sandbox image build failed", "error", err)
		return fmt.Errorf("build sandbox image %q: %w", config.Image, err)
	}
	logger.InfoContext(ctx, "sandbox image built")
	return nil
}

func (s *Sandbox) Start(ctx context.Context) error {
	logger := slog.With("sandbox", s.config.Name)
	logger.InfoContext(ctx, "starting sandbox")
	if s.config.NetworkTarget != "" {
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
	}
	if _, err := s.executor.Run(ctx, command.Spec{Program: dockerProgram, Args: s.runArgs()}); err != nil {
		if s.config.NetworkTarget != "" {
			_ = s.disconnectNetwork(ctx)
			_ = s.removeNetwork(ctx)
		}
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

func (s *Sandbox) runArgs() []string {
	args := []string{
		"run",
		"--detach",
		"--rm",
		"--name", s.config.Name,
		"--hostname", s.config.Name,
		"--network", s.config.Network,
		"--user", s.config.Layout.User,
		"--workdir", s.config.Layout.Workdir,
	}
	if s.config.Security.DropCapabilities {
		args = append(args, "--cap-drop", "ALL")
	}
	if s.config.Security.NoNewPrivileges {
		args = append(args, "--security-opt", "no-new-privileges")
	}
	if s.config.Security.ReadOnlyRoot {
		args = append(args,
			"--read-only",
			"--tmpfs", "/tmp"+tmpfsMode,
			"--tmpfs", s.config.Layout.Workdir+tmpfsMode,
		)
	}
	if strings.TrimSpace(s.config.KubeconfigPath) != "" {
		args = append(args, "--mount", fmt.Sprintf("type=bind,src=%s,dst=%s,readonly",
			filepath.Clean(s.config.KubeconfigPath), s.config.Layout.KubeconfigMountPath))
	}
	for _, key := range slices.Sorted(maps.Keys(s.config.Env)) {
		args = append(args, "--env", key+"="+s.config.Env[key])
	}
	for _, mount := range s.config.Mounts {
		mode := "rw"
		if mount.ReadOnly {
			mode = "ro"
		}
		args = append(args, "--volume", mount.Source+":"+mount.Target+":"+mode)
	}
	args = append(args, s.config.Image)
	if len(s.config.Command) == 0 {
		args = append(args, "sleep", "infinity")
	} else {
		args = append(args, s.config.Command...)
	}
	return args
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
	if s.config.NetworkTarget != "" {
		if err := s.disconnectNetwork(ctx); err != nil {
			logger.ErrorContext(ctx, "sandbox network disconnection failed", "error", err)
			cleanupErrs = append(cleanupErrs, err)
		}
		if err := s.removeNetwork(ctx); err != nil {
			logger.ErrorContext(ctx, "sandbox network removal failed", "error", err)
			cleanupErrs = append(cleanupErrs, err)
		}
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
