package agent

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/config"
	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/baseline"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	promptagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/prompt"
	skillagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/skill"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
)

type Name string

const (
	All      Name = "all"
	Baseline Name = "baseline"
	Prompt   Name = "prompt"
	Skill    Name = "skill"
)

// ConfiguredLlamaParallelism returns the effective llama.cpp slot count.
func ConfiguredLlamaParallelism() (int, error) {
	value := strings.TrimSpace(os.Getenv(envLlamaParallel))
	if value == "" {
		value = "1"
	}
	parallelism, err := strconv.Atoi(value)
	if err != nil || parallelism < 1 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", envLlamaParallel, value)
	}
	return parallelism, nil
}

func Select(values ...string) ([]Name, error) {
	if len(values) == 0 {
		return []Name{Baseline, Prompt, Skill}, nil
	}

	selected := make([]Name, 0, len(values))
	seen := make(map[Name]struct{}, len(values))
	allSelected := false
	for _, value := range values {
		for _, item := range strings.Split(value, ",") {
			normalized := Name(strings.ToLower(strings.TrimSpace(item)))
			if normalized == "" {
				return nil, errors.New("agent must not be blank")
			}
			if normalized == All {
				if len(values) != 1 || len(strings.Split(value, ",")) != 1 || len(selected) > 0 {
					return nil, errors.New("agent all cannot be combined with other agents")
				}
				allSelected = true
				continue
			}
			if normalized != Baseline && normalized != Prompt && normalized != Skill {
				return nil, fmt.Errorf("unsupported agent %q; expected all, baseline, prompt, or skill", item)
			}
			if _, exists := seen[normalized]; exists {
				return nil, fmt.Errorf("agent %q was selected more than once", normalized)
			}
			seen[normalized] = struct{}{}
			selected = append(selected, normalized)
		}
	}
	if allSelected {
		return []Name{Baseline, Prompt, Skill}, nil
	}
	return selected, nil
}

func NewFactory(name Name, benchmarkConfig benchmarkconfig.Config) (rootagent.Factory, error) {
	loopConfig, err := configuredLoopConfig(benchmarkConfig.Agents.Loop)
	if err != nil {
		return nil, err
	}
	var customPrompt string
	var skillConfig benchmarkconfig.SkillAgentConfig
	if name == Prompt {
		customPrompt, err = loadPromptSystemPrompt(benchmarkConfig.Agents.Prompt.SystemPromptFile)
		if err != nil {
			return nil, err
		}
	} else if name == Skill {
		skillConfig = benchmarkConfig.Agents.Skill
	}
	inferenceClient, err := newInferenceClient()
	if err != nil {
		return nil, err
	}
	switch name {
	case Baseline:
		return func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
			return baseline.New(inferenceClient, shell, loopConfig)
		}, nil
	case Prompt:
		return func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
			if customPrompt != "" {
				return promptagent.NewWithSystemPrompt(inferenceClient, shell, loopConfig, customPrompt)
			}
			return promptagent.New(inferenceClient, shell, loopConfig)
		}, nil
	case Skill:
		return func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
			return skillagent.NewWithConfig(inferenceClient, shell, loopConfig, skillConfig.SystemPromptFile, skillConfig.SkillsDir)
		}, nil
	default:
		return nil, fmt.Errorf("unsupported agent %q", name)
	}
}

func configuredLoopConfig(values benchmarkconfig.LoopConfig) (common.Config, error) {
	result := common.DefaultConfig()
	if values.MaxTurns != nil {
		if *values.MaxTurns < 1 {
			return common.Config{}, errors.New("agents.loop.max_turns must be at least 1")
		}
		result.MaxTurns = *values.MaxTurns
	}
	if values.MaxToolCalls != nil {
		if *values.MaxToolCalls < 1 {
			return common.Config{}, errors.New("agents.loop.max_tool_calls must be at least 1")
		}
		result.MaxToolCalls = *values.MaxToolCalls
	}
	if values.ToolTimeoutSeconds != nil {
		if *values.ToolTimeoutSeconds <= 0 {
			return common.Config{}, errors.New("agents.loop.tool_timeout_seconds must be greater than 0")
		}
		result.ToolTimeoutSeconds = *values.ToolTimeoutSeconds
	}
	if values.TimeoutSeconds != nil {
		if *values.TimeoutSeconds <= 0 {
			return common.Config{}, errors.New("agents.loop.timeout_seconds must be greater than 0")
		}
		result.TimeoutSeconds = *values.TimeoutSeconds
	}
	return result, nil
}

func loadPromptSystemPrompt(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read prompt system prompt %q: %w", path, err)
	}
	prompt := string(data)
	if strings.TrimSpace(prompt) == "" {
		return "", fmt.Errorf("prompt system prompt file must not be blank")
	}
	return prompt, nil
}

func NewBaselineFactory(benchmarkConfig benchmarkconfig.Config) (rootagent.Factory, error) {
	return NewFactory(Baseline, benchmarkConfig)
}

func newInferenceClient() (inference.Client, error) {
	model := strings.TrimSpace(os.Getenv(envLlamaModelName))
	if model == "" {
		return nil, fmt.Errorf("%s is required", envLlamaModelName)
	}

	host := strings.TrimSpace(os.Getenv(envLlamaClientHost))
	if host == "" {
		host = "127.0.0.1"
	}
	port := strings.TrimSpace(os.Getenv(envLlamaPort))
	if port == "" {
		port = "8080"
	}
	client, err := llama.NewClient(llama.Config{
		BaseURL: fmt.Sprintf("http://%s:%s", host, port),
		Model:   model,
		Metadata: inference.Metadata{
			Model:           model,
			Artifact:        modelArtifact(),
			Quantization:    strings.TrimSpace(os.Getenv(envLlamaModelQuant)),
			SHA256:          strings.TrimSpace(os.Getenv(envLlamaModelSHA256)),
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

func modelArtifact() string {
	repository := strings.TrimSpace(os.Getenv(envLlamaModelRepository))
	revision := strings.TrimSpace(os.Getenv(envLlamaModelRevision))
	file := strings.TrimSpace(os.Getenv(envLlamaModelFile))
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
	settings := make(map[string]string)
	for _, name := range llamaRuntimeEnvNames {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			settings[name] = value
		}
	}
	if len(settings) == 0 {
		return nil
	}
	return settings
}
