package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/baseline"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/prompt"
	skillagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/skill"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
	projectskills "github.com/Nasus20202/MastersThesis/benchmark/internal/skills"
)

const (
	Baseline = "baseline"
	Prompt   = "prompt"
	Skill    = "skill"
)

var allAgents = []string{Baseline, Prompt, Skill}

// AllNames returns the stable default condition order.
func AllNames() []string { return slices.Clone(allAgents) }

// ResolveNames validates the repeatable --agent values and expands an omitted
// selection to every supported condition.
func ResolveNames(requested []string) ([]string, error) {
	if len(requested) == 0 {
		return AllNames(), nil
	}
	resolved := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, value := range requested {
		name := strings.TrimSpace(value)
		if name == "" {
			return nil, fmt.Errorf("agent name must not be blank")
		}
		switch name {
		case Baseline, Prompt, Skill:
		default:
			return nil, fmt.Errorf("unsupported agent %q (choose %s, %s, or %s)", name, Baseline, Prompt, Skill)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("agent %q was selected more than once", name)
		}
		seen[name] = struct{}{}
		resolved = append(resolved, name)
	}
	return resolved, nil
}

// NewFactories constructs factories for the selected conditions while sharing
// the same inference client and loop configuration across them.
func NewFactories(selected []string) (map[string]rootagent.Factory, error) {
	names, err := ResolveNames(selected)
	if err != nil {
		return nil, err
	}
	inferenceClient, err := newInferenceClient()
	if err != nil {
		return nil, err
	}
	loopConfig := common.DefaultConfig()
	var registry *projectskills.Registry
	for _, name := range names {
		if name != Skill {
			continue
		}
		root, err := skillsDirectory()
		if err != nil {
			return nil, err
		}
		registry, err = projectskills.Load(root)
		if err != nil {
			return nil, fmt.Errorf("load benchmark skills: %w", err)
		}
		break
	}

	factories := make(map[string]rootagent.Factory, len(names))
	for _, name := range names {
		switch name {
		case Baseline:
			factories[name] = func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
				return baseline.New(inferenceClient, shell, loopConfig)
			}
		case Prompt:
			factories[name] = func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
				return prompt.New(inferenceClient, shell, loopConfig)
			}
		case Skill:
			factories[name] = func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
				return skillagent.New(inferenceClient, shell, loopConfig, registry)
			}
		}
	}
	return factories, nil
}

func NewBaselineFactory() (rootagent.Factory, error) {
	factories, err := NewFactories([]string{Baseline})
	if err != nil {
		return nil, err
	}
	return factories[Baseline], nil
}

func newInferenceClient() (inference.Client, error) {
	model := strings.TrimSpace(os.Getenv("LLAMA_MODEL_NAME"))
	if model == "" {
		return nil, fmt.Errorf("LLAMA_MODEL_NAME is required")
	}
	host := strings.TrimSpace(os.Getenv("LLAMA_CLIENT_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}
	port := strings.TrimSpace(os.Getenv("LLAMA_PORT"))
	if port == "" {
		port = "8080"
	}
	client, err := llama.NewClient(llama.Config{
		BaseURL: fmt.Sprintf("http://%s:%s", host, port),
		Model:   model,
		Metadata: inference.Metadata{
			Model:           model,
			Artifact:        modelArtifact(),
			Quantization:    strings.TrimSpace(os.Getenv("LLAMA_MODEL_QUANTIZATION")),
			SHA256:          strings.TrimSpace(os.Getenv("LLAMA_MODEL_SHA256")),
			RuntimeSettings: runtimeSettings(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create llama client: %w", err)
	}
	inferenceClient, err := llama.NewAdapter(client)
	if err != nil {
		return nil, fmt.Errorf("create inference adapter: %w", err)
	}
	return inferenceClient, nil
}

func skillsDirectory() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("BENCHMARK_SKILLS_DIR")); configured != "" {
		return configured, nil
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("find benchmark skills: get working directory: %w", err)
	}
	for current := filepath.Clean(workingDirectory); ; current = filepath.Dir(current) {
		candidate := filepath.Join(current, "skills")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
	}
	return "", fmt.Errorf("find benchmark skills: skills directory not found from %q; set BENCHMARK_SKILLS_DIR", workingDirectory)
}

func modelArtifact() string {
	repository := strings.TrimSpace(os.Getenv("LLAMA_MODEL_REPOSITORY"))
	revision := strings.TrimSpace(os.Getenv("LLAMA_MODEL_REVISION"))
	file := strings.TrimSpace(os.Getenv("LLAMA_MODEL_FILE"))
	artifact := repository
	if revision != "" {
		artifact += "@" + revision
	}
	if file != "" {
		artifact += "/" + file
	}
	return artifact
}

func runtimeSettings() map[string]string {
	const names = "LLAMA_CONTEXT_SIZE LLAMA_GPU_LAYERS LLAMA_VULKAN_DEVICE LLAMA_PARALLEL LLAMA_FLASH_ATTN LLAMA_CACHE_TYPE_K LLAMA_CACHE_TYPE_V LLAMA_MODELS_MAX LLAMA_REASONING LLAMA_REASONING_BUDGET LLAMA_HOST LLAMA_PORT LLAMA_CLIENT_HOST"
	settings := make(map[string]string)
	for _, name := range strings.Fields(names) {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			settings[name] = value
		}
	}
	if len(settings) == 0 {
		return nil
	}
	return settings
}
