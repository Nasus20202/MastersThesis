package skills

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDiscoversManifestMetadataAndReferences(t *testing.T) {
	root := t.TempDir()
	writeRegistryFile(t, filepath.Join(root, "zeta", "SKILL.md"), `---
name: zeta
description: Zeta guidance.
metadata:
  version: "2"
  scope: testing
---

# Zeta
`)
	writeRegistryFile(t, filepath.Join(root, "zeta", "references", "second.md"), "# Second\n")
	writeRegistryFile(t, filepath.Join(root, "alpha", "SKILL.md"), `---
name: alpha
description: Alpha guidance.
metadata:
  version: "1"
---

# Alpha
`)

	registry, err := Load(root)
	require.NoError(t, err)
	descriptors := registry.Descriptors()
	require.Len(t, descriptors, 2)
	assert.Equal(t, "alpha", descriptors[0].Header.Name)
	assert.Equal(t, "zeta", descriptors[1].Header.Name)
	assert.Equal(t, "2", descriptors[1].Header.Metadata["version"])
	assert.Equal(t, []string{"second.md"}, descriptors[1].References)
	assert.Contains(t, registry.Catalog(), "- alpha: Alpha guidance.")
	assert.Contains(t, registry.Catalog(), "metadata: {\"version\":\"1\"}")
	assert.Contains(t, registry.Catalog(), "references: second.md")

	main, err := registry.Load("zeta", "")
	require.NoError(t, err)
	assert.Equal(t, "# Zeta", main.Content)
	assert.Empty(t, main.Reference)
	assert.Equal(t, "testing", main.Header.Metadata["scope"])

	reference, err := registry.Load("zeta", "second.md")
	require.NoError(t, err)
	assert.Equal(t, "second.md", reference.Reference)
	assert.Equal(t, "# Second", reference.Content)
}

func TestLoadRejectsMalformedSkillDirectories(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "missing front matter", content: "# Skill\n", want: "must start with YAML front matter"},
		{name: "missing description", content: "---\nname: test\n---\n# Skill\n", want: "manifest description is required"},
		{name: "missing metadata", content: "---\nname: test\ndescription: missing metadata\n---\n# Skill\n", want: "manifest metadata is required"},
		{name: "invalid name", content: "---\nname: Test\ndescription: invalid\n---\n# Skill\n", want: "manifest name"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeRegistryFile(t, filepath.Join(root, "test", "SKILL.md"), test.content)
			_, err := Load(root)
			assert.ErrorContains(t, err, test.want)
		})
	}
}

func TestLoadRejectsUnknownAndUnsafeReferences(t *testing.T) {
	root := t.TempDir()
	writeRegistryFile(t, filepath.Join(root, "test", "SKILL.md"), "---\nname: test\ndescription: Test guidance.\nmetadata:\n  version: \"1\"\n---\n# Test\n")
	writeRegistryFile(t, filepath.Join(root, "test", "references", "known.md"), "# Known\n")
	registry, err := Load(root)
	require.NoError(t, err)

	_, err = registry.Load("unknown", "")
	assert.ErrorContains(t, err, "is not available")
	_, err = registry.Load("test", "missing.md")
	assert.ErrorContains(t, err, "has no reference")
	_, err = registry.Load("test", "../SKILL.md")
	assert.ErrorContains(t, err, "is invalid")
	_, err = registry.Load("test", "references/known.md")
	assert.ErrorContains(t, err, "is invalid")
}

func TestLoadRejectsNonMarkdownReferences(t *testing.T) {
	root := t.TempDir()
	writeRegistryFile(t, filepath.Join(root, "test", "SKILL.md"), "---\nname: test\ndescription: Test guidance.\nmetadata:\n  version: \"1\"\n---\n# Test\n")
	writeRegistryFile(t, filepath.Join(root, "test", "references", "notes.txt"), "notes\n")
	_, err := Load(root)
	assert.ErrorContains(t, err, "must be a Markdown file")
}

func writeRegistryFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
