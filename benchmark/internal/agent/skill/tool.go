package skill

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/goccy/go-yaml"
)

const (
	loadSkillToolName     = "load_skill"
	loadReferenceToolName = "load_reference"
)

const (
	routingPromptFormat     = "%s\n\nAvailable skills:\n%s\n\n%s"
	routingSkillEntryFormat = "- %s: %s"

	routingGuidance = "Use load_skill to load a skill body when its knowledge is useful. Skills may list references; use load_reference for an exact listed reference when more detail is useful."
)

type skillTool struct {
	files fs.FS
}

type skillMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func newSkillTool() common.Tool {
	return skillTool{files: skillFiles}
}

func newReferenceTool() common.Tool {
	return referenceTool{files: skillFiles}
}

func routingSystemPrompt() (string, error) {
	basePrompt := systemPrompt
	if strings.TrimSpace(basePrompt) == "" {
		return "", errors.New("skill system prompt is required")
	}
	paths, err := discoverSkillPaths(skillFiles)
	if err != nil {
		return "", err
	}
	entries := make([]string, 0, len(paths))
	for _, name := range skillNames(paths) {
		metadata, _, err := readSkillDocument(skillFiles, name)
		if err != nil {
			return "", err
		}
		entry := fmt.Sprintf(routingSkillEntryFormat, metadata.Name, metadata.Description)
		entries = append(entries, entry)
	}
	return fmt.Sprintf(routingPromptFormat, basePrompt, strings.Join(entries, "\n"), routingGuidance), nil
}

func (skillTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        loadSkillToolName,
		Description: "Load the body of one skill from the Available skills list.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Skill name to load"}},"required":["name"],"additionalProperties":false}`),
	}
}

type skillToolArguments struct {
	Name string `json:"name"`
}

type referenceTool struct {
	files fs.FS
}

func (referenceTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        loadReferenceToolName,
		Description: "Load one exact reference file listed by a skill.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"skill":{"type":"string","description":"Skill that owns the reference"},"reference":{"type":"string","description":"Exact reference filename from the skill's reference list"}},"required":["skill","reference"],"additionalProperties":false}`),
	}
}

type skillEvidence struct {
	Skill     string `json:"skill"`
	Reference string `json:"reference,omitempty"`
}

func (t skillTool) Execute(_ context.Context, call inference.ToolCall) common.ToolResult {
	var arguments skillToolArguments
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		err = fmt.Errorf("malformed load_skill arguments: %w", err)
		return common.ToolResult{Content: err.Error(), Error: err}
	}

	name := strings.ToLower(strings.TrimSpace(arguments.Name))
	if name == "" {
		err := errors.New("skill name is required")
		return common.ToolResult{Content: err.Error(), Error: err}
	}
	_, body, err := readSkillDocument(t.files, name)
	if err != nil {
		if owner, reference, ok := findReferenceOwner(t.files, name); ok {
			err = fmt.Errorf("skill %q is not available; %q is a reference of skill %q: call load_reference with {\"skill\":%q,\"reference\":%q} instead", name, reference, owner, owner, reference)
		}
		return common.ToolResult{Content: err.Error(), Error: err}
	}
	return common.ToolResult{
		Content: body,
		Details: skillEvidence{Skill: name},
	}
}

type referenceToolArguments struct {
	Skill     string `json:"skill"`
	Reference string `json:"reference"`
}

func (t referenceTool) Execute(_ context.Context, call inference.ToolCall) common.ToolResult {
	var arguments referenceToolArguments
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil {
		err = fmt.Errorf("malformed load_reference arguments: %w", err)
		return common.ToolResult{Content: err.Error(), Error: err}
	}

	name := strings.ToLower(strings.TrimSpace(arguments.Skill))
	if name == "" {
		err := errors.New("skill name is required")
		return common.ToolResult{Content: err.Error(), Error: err}
	}
	reference := strings.ToLower(strings.TrimSpace(arguments.Reference))
	if reference == "" {
		err := errors.New("reference filename is required")
		return common.ToolResult{Content: err.Error(), Error: err}
	}

	content, err := readReference(t.files, name, reference)
	if err != nil {
		return common.ToolResult{Content: err.Error(), Error: err}
	}
	return common.ToolResult{
		Content: content,
		Details: skillEvidence{Skill: name, Reference: reference},
	}
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

func discoverSkillPaths(files fs.FS) (map[string]string, error) {
	entries, err := fs.ReadDir(files, "skills")
	if err != nil {
		return nil, fmt.Errorf("read skills directory: %w", err)
	}
	paths := make(map[string]string, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		path := "skills/" + entry.Name() + "/SKILL.md"
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

func skillNames(paths map[string]string) []string {
	result := make([]string, 0, len(paths))
	for name := range paths {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

func discoverReferences(files fs.FS, name string) (map[string]string, error) {
	directory := "skills/" + name + "/references"
	entries, err := fs.ReadDir(files, directory)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read references for skill %q: %w", name, err)
	}
	references := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		references[entry.Name()] = directory + "/" + entry.Name()
	}
	return references, nil
}

func referenceNames(references map[string]string) []string {
	result := make([]string, 0, len(references))
	for reference := range references {
		result = append(result, reference)
	}
	sort.Strings(result)
	return result
}

func findReferenceOwner(files fs.FS, query string) (skill, reference string, ok bool) {
	paths, err := discoverSkillPaths(files)
	if err != nil {
		return "", "", false
	}
	for _, name := range skillNames(paths) {
		references, err := discoverReferences(files, name)
		if err != nil {
			continue
		}
		for _, reference := range referenceNames(references) {
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
		err := fmt.Errorf("skill %q is not available; choose one of %s", name, strings.Join(skillNames(paths), ", "))
		return skillMetadata{}, "", err
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

func parseSkillDocument(content, name string) (skillMetadata, string, error) {
	lines := strings.Split(content, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return skillMetadata{}, "", errors.New("SKILL.md must start with YAML front matter")
	}

	closing := -1
	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) == "---" {
			closing = index
			break
		}
	}
	if closing < 0 {
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
