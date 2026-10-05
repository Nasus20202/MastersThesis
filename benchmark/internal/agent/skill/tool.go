package skill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/goccy/go-yaml"
)

const (
	loadSkillToolName     = "load_skill"
	loadReferenceToolName = "load_reference"
)

type skillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

type skillEvidence struct {
	Skill     string `json:"skill"`
	Reference string `json:"reference,omitempty"`
}

// routingSystemPrompt appends the name and description of every skill to the
// base prompt.
func routingSystemPrompt(basePrompt string, files fs.FS) (string, error) {
	if strings.TrimSpace(basePrompt) == "" {
		return "", errors.New("skill system prompt is required")
	}
	paths, err := discoverSkillPaths(files)
	if err != nil {
		return "", err
	}
	entries := make([]string, 0, len(paths))
	for _, name := range slices.Sorted(maps.Keys(paths)) {
		metadata, _, err := readSkillDocument(files, name)
		if err != nil {
			return "", err
		}
		entries = append(entries, fmt.Sprintf("- %s: %s", metadata.Name, metadata.Description))
	}
	return fmt.Sprintf("%s\n\nAvailable skills:\n%s", basePrompt, strings.Join(entries, "\n")), nil
}

type skillTool struct{ files fs.FS }

func (skillTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        loadSkillToolName,
		Description: "Load the body of one skill from the Available skills list.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Skill name to load"}},"required":["name"],"additionalProperties":false}`),
	}
}

func (t skillTool) Execute(_ context.Context, call inference.ToolCall) common.ToolResult {
	var arguments struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		return common.Failed(fmt.Errorf("malformed load_skill arguments: %w", err), nil)
	}
	name := strings.ToLower(strings.TrimSpace(arguments.Name))
	if name == "" {
		return common.Failed(errors.New("skill name is required"), nil)
	}
	_, body, err := readSkillDocument(t.files, name)
	if err != nil {
		if owner, reference, ok := findReferenceOwner(t.files, name); ok {
			err = fmt.Errorf("skill %q is not available; %q is a reference of skill %q: call load_reference with {\"skill\":%q,\"reference\":%q} instead", name, reference, owner, owner, reference)
		}
		return common.Failed(err, nil)
	}
	return common.ToolResult{Content: body, Details: skillEvidence{Skill: name}}
}

type referenceTool struct{ files fs.FS }

func (referenceTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        loadReferenceToolName,
		Description: "Load one exact reference file listed by a skill.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"skill":{"type":"string","description":"Skill that owns the reference"},"reference":{"type":"string","description":"Exact reference filename from the skill's reference list"}},"required":["skill","reference"],"additionalProperties":false}`),
	}
}

func (t referenceTool) Execute(_ context.Context, call inference.ToolCall) common.ToolResult {
	var arguments struct {
		Skill     string `json:"skill"`
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		return common.Failed(fmt.Errorf("malformed load_reference arguments: %w", err), nil)
	}
	name := strings.ToLower(strings.TrimSpace(arguments.Skill))
	if name == "" {
		return common.Failed(errors.New("skill name is required"), nil)
	}
	reference := strings.ToLower(strings.TrimSpace(arguments.Reference))
	if reference == "" {
		return common.Failed(errors.New("reference filename is required"), nil)
	}
	content, err := readReference(t.files, name, reference)
	if err != nil {
		return common.Failed(err, nil)
	}
	return common.ToolResult{Content: content, Details: skillEvidence{Skill: name, Reference: reference}}
}

func readReference(files fs.FS, name, reference string) (string, error) {
	if _, _, err := readSkillDocument(files, name); err != nil {
		return "", err
	}
	references, err := discoverReferences(files, name)
	if err != nil {
		return "", err
	}
	if len(references) == 0 {
		return "", fmt.Errorf("skill %q has no reference files", name)
	}
	path, ok := references[reference]
	if !ok {
		return "", fmt.Errorf("reference %q is not available for skill %q", reference, name)
	}
	data, err := fs.ReadFile(files, path)
	if err != nil {
		return "", fmt.Errorf("read reference %q for skill %q: %w", reference, name, err)
	}
	return string(data), nil
}

// discoverSkillPaths maps each skill directory name to its SKILL.md.
func discoverSkillPaths(files fs.FS) (map[string]string, error) {
	root := skillRoot(files)
	entries, err := fs.ReadDir(files, root)
	if err != nil {
		return nil, fmt.Errorf("read skills directory: %w", err)
	}
	paths := make(map[string]string, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := path.Join(root, entry.Name(), "SKILL.md")
		info, err := fs.Stat(files, path)
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("skill %q is missing SKILL.md", entry.Name())
		}
		if err != nil {
			return nil, fmt.Errorf("stat skill %q: %w", entry.Name(), err)
		}
		if info.IsDir() {
			return nil, fmt.Errorf("skill %q SKILL.md is a directory", entry.Name())
		}
		paths[entry.Name()] = path
	}
	return paths, nil
}

// discoverReferences maps each reference file name of a skill to its path.
func discoverReferences(files fs.FS, name string) (map[string]string, error) {
	directory := path.Join(skillRoot(files), name, "references")
	entries, err := fs.ReadDir(files, directory)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read references for skill %q: %w", name, err)
	}
	references := make(map[string]string, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			references[entry.Name()] = directory + "/" + entry.Name()
		}
	}
	return references, nil
}

// skillRoot accepts both the embedded tree, which keeps the skills directory,
// and a configured skills directory.
func skillRoot(files fs.FS) string {
	if info, err := fs.Stat(files, "skills"); err == nil && info.IsDir() {
		return "skills"
	}
	return "."
}

// findReferenceOwner finds the skill whose reference the model asked for as a
// skill name, with or without the .md extension.
func findReferenceOwner(files fs.FS, query string) (skill, reference string, ok bool) {
	paths, err := discoverSkillPaths(files)
	if err != nil {
		return "", "", false
	}
	for _, name := range slices.Sorted(maps.Keys(paths)) {
		references, err := discoverReferences(files, name)
		if err != nil {
			continue
		}
		for _, reference := range slices.Sorted(maps.Keys(references)) {
			if lowered := strings.ToLower(reference); lowered == query || lowered == query+".md" {
				return name, reference, true
			}
		}
	}
	return "", "", false
}

func readSkillDocument(files fs.FS, name string) (skillMetadata, string, error) {
	paths, err := discoverSkillPaths(files)
	if err != nil {
		return skillMetadata{}, "", err
	}
	path, ok := paths[name]
	if !ok {
		return skillMetadata{}, "", fmt.Errorf("skill %q is not available; choose one of %s", name, strings.Join(slices.Sorted(maps.Keys(paths)), ", "))
	}
	data, err := fs.ReadFile(files, path)
	if err != nil {
		return skillMetadata{}, "", fmt.Errorf("read skill %q: %w", name, err)
	}
	metadata, body, err := parseSkillDocument(string(data), name)
	if err != nil {
		return skillMetadata{}, "", fmt.Errorf("invalid skill %q: %w", name, err)
	}
	return metadata, body, nil
}

// parseSkillDocument splits SKILL.md into its YAML front matter and body.
func parseSkillDocument(content, name string) (skillMetadata, string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return skillMetadata{}, "", errors.New("SKILL.md must start with YAML front matter")
	}
	closing := slices.IndexFunc(lines[1:], func(line string) bool { return strings.TrimSpace(line) == "---" }) + 1
	if closing == 0 {
		return skillMetadata{}, "", errors.New("SKILL.md front matter is not closed")
	}

	var metadata skillMetadata
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:closing], "\n")), &metadata); err != nil {
		return skillMetadata{}, "", fmt.Errorf("parse front matter: %w", err)
	}
	if strings.TrimSpace(metadata.Name) == "" {
		return skillMetadata{}, "", errors.New("front matter name is required")
	}
	if metadata.Name != name {
		return skillMetadata{}, "", fmt.Errorf("front matter name %q does not match directory %q", metadata.Name, name)
	}
	if strings.TrimSpace(metadata.Description) == "" {
		return skillMetadata{}, "", errors.New("front matter description is required")
	}
	return metadata, strings.TrimSpace(strings.Join(lines[closing+1:], "\n")), nil
}
