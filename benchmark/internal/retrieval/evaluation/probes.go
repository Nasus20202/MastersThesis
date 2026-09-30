// Package evaluation compares retrieval configurations on the knowledge-check
// probes and selects the default one.
package evaluation

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"
)

// Primary holds the probe's own source files; Relevant adds the files cited by
// its criteria and provided context.
type Probe struct {
	ID       string   `json:"id"`
	Prompt   string   `json:"prompt"`
	Primary  []string `json:"primary"`
	Relevant []string `json:"relevant"`
}

var corpusPath = regexp.MustCompile(`content/en/docs/[^\s"#@\]]+\.md`)

// LoadProbes also returns the SHA-256 of the task file.
func LoadProbes(path string) ([]Probe, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	probes, err := ParseProbes(data)
	return probes, sha256Hex(data), err
}

// ParseProbes reads a knowledge-check task file such as context-tasks.json.
func ParseProbes(data []byte) ([]Probe, error) {
	var file struct {
		Tasks []struct {
			ID      string `json:"id"`
			Prompt  string `json:"prompt"`
			Source  string `json:"source"`
			Sources []struct {
				Path string `json:"path"`
			} `json:"sources"`
			Criteria        map[string]string `json:"criteria"`
			ProvidedContext []struct {
				Source struct {
					Path string `json:"path"`
				} `json:"source"`
			} `json:"provided_context"`
		} `json:"tasks"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("parse probes: %w", err)
	}
	probes := make([]Probe, 0, len(file.Tasks))
	for _, task := range file.Tasks {
		probe := Probe{ID: task.ID, Prompt: strings.TrimSpace(task.Prompt)}
		probe.Primary = corpusPath.FindAllString(task.Source, -1)
		for _, source := range task.Sources {
			probe.Primary = append(probe.Primary, source.Path)
		}
		probe.Relevant = slices.Clone(probe.Primary)
		for _, criterion := range task.Criteria {
			probe.Relevant = append(probe.Relevant, corpusPath.FindAllString(criterion, -1)...)
		}
		for _, context := range task.ProvidedContext {
			probe.Relevant = append(probe.Relevant, context.Source.Path)
		}
		probe.Primary = sortedUnique(probe.Primary)
		probe.Relevant = sortedUnique(probe.Relevant)
		if probe.ID == "" || probe.Prompt == "" || len(probe.Primary) == 0 {
			return nil, fmt.Errorf("probe %q needs an id, a prompt and a source", task.ID)
		}
		probes = append(probes, probe)
	}
	if len(probes) == 0 {
		return nil, errors.New("parse probes: no tasks")
	}
	return probes, nil
}

func sortedUnique(values []string) []string {
	var result []string
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	slices.Sort(result)
	return slices.Compact(result)
}
