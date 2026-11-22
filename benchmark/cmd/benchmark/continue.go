package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// loadPrefixes reads every recorded attempt of a continuation run and keys
// its conversation by the scenario the attempt recorded. A scenario may
// appear once, so each task of a run continues its own attempt.
func loadPrefixes(paths []string) (map[string]*common.Prefix, error) {
	prefixes := make(map[string]*common.Prefix, len(paths))
	root, _ := repositoryRoot()
	for _, path := range paths {
		prefix, scenario, err := loadPrefix(root, path)
		if err != nil {
			return nil, err
		}
		if _, ok := prefixes[scenario]; ok {
			return nil, fmt.Errorf("two recorded attempts for scenario %q", scenario)
		}
		prefixes[scenario] = &prefix
	}
	return prefixes, nil
}

// repositoryRoot is the checked-out repository the run belongs to.
func repositoryRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not inside the repository")
		}
		dir = parent
	}
}

// scenariosRoot is the corpus a continuation run uses when it names its
// scenarios only in the records.
func scenariosRoot() (string, error) {
	root, err := repositoryRoot()
	if err != nil {
		return "", errors.New("continue-from without --scenario needs the repository")
	}
	dir := filepath.Join(root, "benchmark", "scenarios")
	if _, err := os.Stat(dir); err != nil {
		return "", fmt.Errorf("no scenario corpus at %s", dir)
	}
	return dir, nil
}

// recordedSource is the path an attempt stores: relative to the repository when
// the record lies inside it, so the evidence does not depend on the machine
// that produced it.
func recordedSource(root, path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path
	}
	return filepath.ToSlash(rel)
}

// loadPrefix reads a recorded attempt (with root as the repository it is
// stored in) and returns the conversation up to the
// point where its model stopped: every message before the final assistant
// message that made no tool call. It also returns the attempt's scenario id.
func loadPrefix(root, path string) (common.Prefix, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return common.Prefix{}, "", fmt.Errorf("read continue-from record: %w", err)
	}
	var record struct {
		ScenarioID string `json:"scenario_id"`
		Agent      struct {
			Messages []inference.Message `json:"messages"`
		} `json:"agent"`
	}
	if err := json.Unmarshal(data, &record); err != nil {
		return common.Prefix{}, "", fmt.Errorf("parse continue-from record: %w", err)
	}
	messages := record.Agent.Messages
	if len(messages) < 3 {
		return common.Prefix{}, "", errors.New("continue-from record has no conversation to continue")
	}
	last := messages[len(messages)-1]
	if last.Role != "assistant" || len(last.ToolCalls) > 0 {
		return common.Prefix{}, "", errors.New("continue-from record did not end with a final assistant message")
	}
	return common.Prefix{Source: recordedSource(root, path), Messages: messages[:len(messages)-1]}, record.ScenarioID, nil
}
