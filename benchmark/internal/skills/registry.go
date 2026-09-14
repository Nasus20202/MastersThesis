// Package skills loads the repository-owned skills exposed to benchmark
// agents. A registry is deliberately allowlisted: model requests may select a
// known skill and one of its direct reference files, but never an arbitrary
// path.
package skills

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goccy/go-yaml"
)

const MainFilename = "SKILL.md"

// Header is the universal skill manifest stored in SKILL.md front matter.
// Metadata is intentionally open-ended so skill authors can add provenance or
// version information without changing the loader.
type Header struct {
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description" json:"description"`
	Metadata    map[string]any `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Descriptor is the model-visible summary of one registered skill.
type Descriptor struct {
	Header     Header
	References []string
}

// Loaded is the content returned after a skill or one of its references is
// selected.
type Loaded struct {
	Header    Header
	Reference string
	Content   string
}

type document struct {
	header         Header
	content        string
	references     map[string]string
	referenceNames []string
}

// Registry contains validated skills rooted at one repository directory.
type Registry struct {
	documents map[string]document
}

// Load discovers and validates all direct skill directories under root.
func Load(root string) (*Registry, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("skills root is required")
	}
	rootInfo, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("stat skills root: %w", err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("skills root %q is not a directory", root)
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("read skills root: %w", err)
	}
	registry := &Registry{documents: make(map[string]document)}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !entry.IsDir() {
			continue
		}
		directory := filepath.Join(root, entry.Name())
		mainPath := filepath.Join(directory, MainFilename)
		if _, err := os.Stat(mainPath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("skill %q is missing %s", entry.Name(), MainFilename)
			}
			return nil, fmt.Errorf("stat skill %q: %w", entry.Name(), err)
		}
		header, content, err := readSkillMarkdown(mainPath)
		if err != nil {
			return nil, fmt.Errorf("load skill %q: %w", entry.Name(), err)
		}
		if header.Name != entry.Name() {
			return nil, fmt.Errorf("skill directory %q does not match manifest name %q", entry.Name(), header.Name)
		}
		references, referenceNames, err := readReferences(directory)
		if err != nil {
			return nil, fmt.Errorf("load references for skill %q: %w", entry.Name(), err)
		}
		registry.documents[header.Name] = document{
			header:          header,
			content:         content,
			references:      references,
			referenceNames: referenceNames,
		}
	}
	if len(registry.documents) == 0 {
		return nil, errors.New("skills root contains no skills")
	}
	return registry, nil
}

// Descriptors returns deterministic, detached manifest summaries.
func (r *Registry) Descriptors() []Descriptor {
	if r == nil {
		return nil
	}
	descriptors := make([]Descriptor, 0, len(r.documents))
	for _, item := range r.documents {
		descriptors = append(descriptors, Descriptor{
			Header: Header{
				Name:        item.header.Name,
				Description: item.header.Description,
				Metadata:    maps.Clone(item.header.Metadata),
			},
			References: slices.Clone(item.referenceNames),
		})
	}
	slices.SortFunc(descriptors, func(left, right Descriptor) int {
		return strings.Compare(left.Header.Name, right.Header.Name)
	})
	return descriptors
}

// Catalog formats only manifest data and discovered reference names for use in
// the skill condition's short system message.
func (r *Registry) Catalog() string {
	var builder strings.Builder
	for _, descriptor := range r.Descriptors() {
		fmt.Fprintf(&builder, "- %s: %s\n", descriptor.Header.Name, descriptor.Header.Description)
		if len(descriptor.Header.Metadata) > 0 {
			metadata, err := json.Marshal(descriptor.Header.Metadata)
			if err == nil {
				fmt.Fprintf(&builder, "  metadata: %s\n", metadata)
			}
		}
		if len(descriptor.References) > 0 {
			fmt.Fprintf(&builder, "  references: %s\n", strings.Join(descriptor.References, ", "))
		}
	}
	return strings.TrimSpace(builder.String())
}

// Load returns a skill's main content or one exact direct reference file.
// An empty reference selects SKILL.md.
func (r *Registry) Load(name, reference string) (Loaded, error) {
	if r == nil {
		return Loaded{}, errors.New("skills registry is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Loaded{}, errors.New("skill name is required")
	}
	item, ok := r.documents[name]
	if !ok {
		return Loaded{}, fmt.Errorf("skill %q is not available", name)
	}
	reference = strings.TrimSpace(reference)
	if reference == "" {
		return Loaded{Header: cloneHeader(item.header), Content: item.content}, nil
	}
	if !safeReferenceName(reference) {
		return Loaded{}, fmt.Errorf("skill reference %q is invalid", reference)
	}
	content, ok := item.references[reference]
	if !ok {
		return Loaded{}, fmt.Errorf("skill %q has no reference %q", name, reference)
	}
	return Loaded{Header: cloneHeader(item.header), Reference: reference, Content: content}, nil
}

func readSkillMarkdown(path string) (Header, string, error) {
	data, err := readRegularFile(path)
	if err != nil {
		return Header{}, "", err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(text, "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return Header{}, "", errors.New("SKILL.md must start with YAML front matter")
	}
	closing := -1
	for index := 1; index < len(lines); index++ {
		if strings.TrimSpace(lines[index]) == "---" {
			closing = index
			break
		}
	}
	if closing < 0 {
		return Header{}, "", errors.New("SKILL.md front matter is not closed")
	}
	var header Header
	if err := yaml.Unmarshal([]byte(strings.Join(lines[1:closing], "\n")), &header); err != nil {
		return Header{}, "", fmt.Errorf("parse front matter: %w", err)
	}
	header.Name = strings.TrimSpace(header.Name)
	header.Description = strings.TrimSpace(header.Description)
	if !validSkillName(header.Name) {
		return Header{}, "", errors.New("manifest name must contain only lowercase letters, digits, and hyphens")
	}
	if header.Description == "" {
		return Header{}, "", errors.New("manifest description is required")
	}
	if strings.ContainsAny(header.Description, "\r\n") {
		return Header{}, "", errors.New("manifest description must be one line")
	}
	if len(header.Metadata) == 0 {
		return Header{}, "", errors.New("manifest metadata is required")
	}
	body := strings.TrimSpace(strings.Join(lines[closing+1:], "\n"))
	if body == "" {
		return Header{}, "", errors.New("skill body is required")
	}
	return header, body, nil
}

func readReferences(directory string) (map[string]string, []string, error) {
	referencesDirectory := filepath.Join(directory, "references")
	entryInfo, err := os.Lstat(referencesDirectory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if entryInfo.Mode()&os.ModeSymlink != 0 {
		return nil, nil, errors.New("references symbolic links are not allowed")
	}
	if !entryInfo.IsDir() {
		return nil, nil, errors.New("references is not a directory")
	}
	entries, err := os.ReadDir(referencesDirectory)
	if err != nil {
		return nil, nil, err
	}
	references := make(map[string]string)
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if entry.IsDir() {
			return nil, nil, fmt.Errorf("nested reference directory %q is not supported", entry.Name())
		}
		if strings.ToLower(filepath.Ext(entry.Name())) != ".md" {
			return nil, nil, fmt.Errorf("reference %q must be a Markdown file", entry.Name())
		}
		if !safeReferenceName(entry.Name()) {
			return nil, nil, fmt.Errorf("reference %q is not a direct filename", entry.Name())
		}
		content, err := readRegularFile(filepath.Join(referencesDirectory, entry.Name()))
		if err != nil {
			return nil, nil, fmt.Errorf("read %q: %w", entry.Name(), err)
		}
		if strings.TrimSpace(string(content)) == "" {
			return nil, nil, fmt.Errorf("reference %q is empty", entry.Name())
		}
		references[entry.Name()] = strings.TrimSpace(string(content))
		names = append(names, entry.Name())
	}
	slices.Sort(names)
	return references, names, nil
}

func readRegularFile(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("symbolic links are not allowed")
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("skill file is not regular")
	}
	return os.ReadFile(path)
}

func cloneHeader(header Header) Header {
	header.Metadata = maps.Clone(header.Metadata)
	return header
}

func validSkillName(name string) bool {
	if name == "" || name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	for _, character := range name {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '-' {
			return false
		}
	}
	return true
}

func safeReferenceName(reference string) bool {
	return reference != "." && reference != ".." && filepath.Base(reference) == reference && !strings.ContainsAny(reference, `/\\`)
}
