package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Logging LoggingConfig `yaml:"logging,omitempty"`
	Agents  AgentsConfig  `yaml:"agents,omitempty"`
}

type LoggingConfig struct {
	Level  string `yaml:"level,omitempty"`
	Format string `yaml:"format,omitempty"`
	Color  string `yaml:"color,omitempty"`
}

type AgentsConfig struct {
	Loop   LoopConfig        `yaml:"loop,omitempty"`
	Prompt PromptAgentConfig `yaml:"prompt,omitempty"`
}

type LoopConfig struct {
	MaxTurns           *int     `yaml:"max_turns,omitempty"`
	MaxToolCalls       *int     `yaml:"max_tool_calls,omitempty"`
	ToolTimeoutSeconds *float64 `yaml:"tool_timeout_seconds,omitempty"`
	TimeoutSeconds     *float64 `yaml:"timeout_seconds,omitempty"`
}

type PromptAgentConfig struct {
	SystemPromptFile string `yaml:"system_prompt_file,omitempty"`
}

// Load reads benchmark-specific YAML configuration files in order. Later
// files override earlier files. Prompt file paths in a config file are
// resolved relative to that file.
func Load(paths ...string) (Config, error) {
	values := make(map[string]interface{})
	for _, path := range paths {
		fileValues, err := read(path)
		if err != nil {
			return Config{}, err
		}
		merge(values, fileValues)
	}

	data, err := yaml.Marshal(values)
	if err != nil {
		return Config{}, fmt.Errorf("marshal merged config: %w", err)
	}
	decoder := yaml.NewDecoder(bytes.NewReader(data), yaml.DisallowUnknownField())
	var result Config
	if err := decoder.Decode(&result); err != nil {
		return Config{}, fmt.Errorf("invalid benchmark config: %s", yaml.FormatError(err, false, true))
	}
	return result, nil
}

func read(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read benchmark config %q: %w", path, err)
	}

	decoder := yaml.NewDecoder(bytes.NewReader(data))
	values := make(map[string]interface{})
	if err := decoder.Decode(&values); err != nil {
		return nil, fmt.Errorf("invalid benchmark config %q: %s", path, yaml.FormatError(err, false, true))
	}
	var extraDocument interface{}
	if err := decoder.Decode(&extraDocument); err == nil {
		return nil, fmt.Errorf("benchmark config %q must contain exactly one document", path)
	} else if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("invalid benchmark config %q: %s", path, yaml.FormatError(err, false, true))
	}
	resolvePromptPath(values, path)
	return values, nil
}

func merge(destination, source map[string]interface{}) {
	for key, value := range source {
		sourceMap, sourceIsMap := value.(map[string]interface{})
		destinationMap, destinationIsMap := destination[key].(map[string]interface{})
		if sourceIsMap && destinationIsMap {
			merge(destinationMap, sourceMap)
			continue
		}
		destination[key] = value
	}
}

func resolvePromptPath(values map[string]interface{}, configPath string) {
	agents, ok := values["agents"].(map[string]interface{})
	if !ok {
		return
	}
	prompt, ok := agents["prompt"].(map[string]interface{})
	if !ok {
		return
	}
	path, ok := prompt["system_prompt_file"].(string)
	if !ok || path == "" || filepath.IsAbs(path) {
		return
	}
	prompt["system_prompt_file"] = filepath.Join(filepath.Dir(configPath), path)
}
