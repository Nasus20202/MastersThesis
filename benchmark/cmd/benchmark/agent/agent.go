// Package agent resolves the CLI's agent-name selection into the inference
// client and agent factories used to run a benchmark.
package agent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

	benchmarkconfig "github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/config"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/llamaenv"
	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/baseline"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	promptagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/prompt"
	ragagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/rag"
	skillagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent/skill"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
)

type Name string

const (
	All      Name = "all"
	Baseline Name = "baseline"
	Prompt   Name = "prompt"
	Skill    Name = "skill"
	RAG      Name = "rag"
)

// allAgents is the default selection and the expansion of "all".
var allAgents = []Name{Baseline, Prompt, Skill, RAG}

// Select parses repeated or comma-separated agent names; none selects all.
func Select(values ...string) ([]Name, error) {
	if len(values) == 0 {
		return slices.Clone(allAgents), nil
	}
	items := strings.Split(strings.Join(values, ","), ",")
	selected := make([]Name, 0, len(items))
	for _, item := range items {
		name := Name(strings.ToLower(strings.TrimSpace(item)))
		switch {
		case name == "":
			return nil, errors.New("agent must not be blank")
		case name == All:
			if len(items) != 1 {
				return nil, errors.New("agent all cannot be combined with other agents")
			}
			return slices.Clone(allAgents), nil
		case !slices.Contains(allAgents, name):
			return nil, fmt.Errorf("unsupported agent %q; expected all, baseline, prompt, skill, or rag", item)
		case slices.Contains(selected, name):
			return nil, fmt.Errorf("agent %q was selected more than once", name)
		}
		selected = append(selected, name)
	}
	return selected, nil
}

// NewFactory builds the factory for one agent. search is required only for
// the RAG agent; see OpenSearch.
func NewFactory(name Name, benchmarkConfig benchmarkconfig.Config, search *ragagent.Search) (rootagent.Factory, error) {
	loopConfig, err := configuredLoopConfig(benchmarkConfig.Agents.Loop)
	if err != nil {
		return nil, err
	}
	var newAgent func(inference.Client, sandboxintegration.Executor) (rootagent.Agent, error)
	switch name {
	case Baseline:
		newAgent = func(client inference.Client, shell sandboxintegration.Executor) (rootagent.Agent, error) {
			return baseline.New(client, shell, loopConfig)
		}
	case Prompt:
		systemPrompt, err := readSystemPrompt(benchmarkConfig.Agents.Prompt.SystemPromptFile)
		if err != nil {
			return nil, err
		}
		newAgent = func(client inference.Client, shell sandboxintegration.Executor) (rootagent.Agent, error) {
			return promptagent.New(client, shell, loopConfig, systemPrompt)
		}
	case Skill:
		systemPrompt, err := readSystemPrompt(benchmarkConfig.Agents.Skill.SystemPromptFile)
		if err != nil {
			return nil, err
		}
		var files fs.FS
		if dir := strings.TrimSpace(benchmarkConfig.Agents.Skill.SkillsDir); dir != "" {
			files = os.DirFS(dir)
		}
		newAgent = func(client inference.Client, shell sandboxintegration.Executor) (rootagent.Agent, error) {
			return skillagent.New(client, shell, loopConfig, systemPrompt, files)
		}
	case RAG:
		if search == nil {
			return nil, errors.New("rag agent requires an opened retrieval index")
		}
		systemPrompt, err := readSystemPrompt(benchmarkConfig.Agents.RAG.SystemPromptFile)
		if err != nil {
			return nil, err
		}
		newAgent = func(client inference.Client, shell sandboxintegration.Executor) (rootagent.Agent, error) {
			return ragagent.New(client, shell, loopConfig, systemPrompt, search)
		}
	default:
		return nil, fmt.Errorf("unsupported agent %q", name)
	}
	inferenceClient, err := llamaenv.NewChatClient()
	if err != nil {
		return nil, err
	}
	return func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
		return newAgent(inferenceClient, shell)
	}, nil
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
	if values.FileTools != nil {
		result.FileTools = *values.FileTools
	}
	return result, nil
}

// readSystemPrompt reads a configured system prompt file. An empty path returns
// an empty prompt, which selects the condition's embedded default.
func readSystemPrompt(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read system prompt %q: %w", path, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return "", fmt.Errorf("system prompt file %q must not be blank", path)
	}
	return string(data), nil
}
