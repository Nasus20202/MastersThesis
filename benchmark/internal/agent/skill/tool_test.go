package skill

import (
	"context"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSkillFilesAreEmbedded(t *testing.T) {
	skills, err := discoverSkillPaths(skillFiles)
	require.NoError(t, err)
	assert.NotEmpty(t, skills)
	for _, path := range skills {
		data, err := fs.ReadFile(skillFiles, path)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	}
	references, err := discoverReferences(skillFiles, "kubernetes")
	require.NoError(t, err)
	assert.Len(t, references, 18)
	for _, path := range references {
		data, err := fs.ReadFile(skillFiles, path)
		require.NoError(t, err)
		assert.NotEmpty(t, data)
	}
}

func TestRoutingPromptListsSkillManifestsWithoutReferences(t *testing.T) {
	prompt, err := routingSystemPrompt()
	require.NoError(t, err)
	for _, entry := range []string{
		"- authorization: Repairing RBAC and ServiceAccount faults",
		"- configuration: Repairing ConfigMap and Secret references",
		"- governance: Repairing admission rejections from LimitRange and ResourceQuota",
		"- kubectl: kubectl command syntax",
		"- kubernetes: Kubernetes concepts",
		"- networking: Repairing Service reachability, NetworkPolicy, DNS, Ingress and Gateway routing",
		"- resources: Repairing CPU and memory requests and limits",
		"- scheduling: Placing or repairing Pending and unschedulable Pods and DaemonSets",
		"- security: Making workloads compliant with Pod Security Admission",
		"- storage: Repairing PersistentVolumeClaims, volume mounts and shared volumes",
		"- troubleshooting: General diagnosis, repair, and verification workflow",
		"- workloads: Repairing rollouts, probes, crash loops, StatefulSets, headless Services and Jobs",
	} {
		assert.Contains(t, prompt, entry)
	}
	assert.Contains(t, prompt, "Before the first Bash call, load the `troubleshooting` skill")
	assert.Contains(t, prompt, "Match the observed symptom to the subsystem")
	assert.Contains(t, prompt, "load the reference that covers the affected subsystem with `load_reference`")
	assert.Contains(t, prompt, "using `skill: \"kubernetes\"` and the exact filename")
	assert.Contains(t, prompt, "Load `kubectl` as well when command or patch semantics are uncertain")
	assert.Contains(t, prompt, "Prove the requested outcome before reporting success")
	assert.Contains(t, prompt, "Avoid these common wrong fixes")
	assert.Contains(t, prompt, "Report `Outcome: complete` only when every required check passes")
	assert.NotContains(t, prompt, "Load other skills and listed references only when their guidance is useful")
	assert.NotContains(t, prompt, "Execute every fix")
}

func TestKubernetesReferencesDoNotContainTroubleshootingPlaybooks(t *testing.T) {
	references, err := discoverReferences(skillFiles, "kubernetes")
	require.NoError(t, err)
	for name, path := range references {
		data, err := fs.ReadFile(skillFiles, path)
		require.NoError(t, err)
		content := strings.ToLower(string(data))
		assert.NotContains(t, content, "## troubleshooting", name)
		assert.NotContains(t, content, "follow the chain", name)
	}
}

func TestSkillToolSuggestsReferenceForReferenceName(t *testing.T) {
	tool := newSkillTool()

	for _, name := range []string{"api", "api.md"} {
		result := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"name":"` + name + `"}`})
		require.Error(t, result.Error)
		assert.Contains(t, result.Content, `skill "api`)
		assert.Contains(t, result.Content, "is a reference of skill")
		assert.Contains(t, result.Content, `"kubernetes"`)
		assert.Contains(t, result.Content, "load_reference")
	}

	unknown := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"name":"unknown"}`})
	require.Error(t, unknown.Error)
	assert.Contains(t, unknown.Content, `skill "unknown" is not available`)
	assert.NotContains(t, unknown.Content, "load_reference")
}

func TestSkillToolDefinition(t *testing.T) {
	definition := newSkillTool().Definition()
	assert.Equal(t, loadSkillToolName, definition.Name)
	assert.JSONEq(t, `{"type":"object","properties":{"name":{"type":"string","description":"Skill name to load"}},"required":["name"],"additionalProperties":false}`, string(definition.Parameters))
}

func TestSkillToolLoadsSkill(t *testing.T) {
	tool := newSkillTool()

	skillResult := tool.Execute(context.Background(), inference.ToolCall{Arguments: `{"name":"workloads"}`})
	require.NoError(t, skillResult.Error)
	assert.Contains(t, skillResult.Content, "# Workloads")
	assert.NotContains(t, skillResult.Content, "name: workloads")
	assert.NotContains(t, skillResult.Content, "description:")
	assert.Equal(t, skillEvidence{Skill: "workloads"}, skillResult.Details)
}

func TestReferenceToolDefinition(t *testing.T) {
	definition := newReferenceTool().Definition()
	assert.Equal(t, loadReferenceToolName, definition.Name)
	assert.JSONEq(t, `{"type":"object","properties":{"skill":{"type":"string","description":"Skill that owns the reference"},"reference":{"type":"string","description":"Exact reference filename from the skill's reference list"}},"required":["skill","reference"],"additionalProperties":false}`, string(definition.Parameters))
}

func TestReferenceToolLoadsReference(t *testing.T) {
	result := newReferenceTool().Execute(context.Background(), inference.ToolCall{Arguments: `{"skill":"kubernetes","reference":"workloads.md"}`})
	require.NoError(t, result.Error)
	assert.Contains(t, result.Content, "# Kubernetes Workloads")
	assert.Equal(t, skillEvidence{Skill: "kubernetes", Reference: "workloads.md"}, result.Details)
}

func TestSkillToolRejectsInvalidRequests(t *testing.T) {
	for _, test := range []struct {
		name      string
		arguments string
		want      string
	}{
		{name: "malformed", arguments: `{`, want: "malformed load_skill arguments"},
		{name: "missing name", arguments: `{}`, want: "skill name is required"},
		{name: "unknown skill", arguments: `{"name":"unknown"}`, want: `skill "unknown" is not available`},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := newSkillTool().Execute(context.Background(), inference.ToolCall{Arguments: test.arguments})
			assert.ErrorContains(t, result.Error, test.want)
			assert.Contains(t, result.Content, test.want)
		})
	}
}

func TestReferenceToolRejectsInvalidRequests(t *testing.T) {
	for _, test := range []struct {
		name      string
		arguments string
		want      string
	}{
		{name: "malformed", arguments: `{`, want: "malformed load_reference arguments"},
		{name: "missing skill", arguments: `{}`, want: "skill name is required"},
		{name: "missing reference", arguments: `{"skill":"kubernetes"}`, want: "reference filename is required"},
		{name: "unknown skill", arguments: `{"skill":"unknown","reference":"workloads.md"}`, want: `skill "unknown" is not available`},
		{name: "unsupported reference", arguments: `{"skill":"kubectl","reference":"workloads.md"}`, want: `skill "kubectl" has no reference files`},
		{name: "unknown reference", arguments: `{"skill":"kubernetes","reference":"missing.md"}`, want: `reference "missing.md" is not available`},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := newReferenceTool().Execute(context.Background(), inference.ToolCall{Arguments: test.arguments})
			assert.ErrorContains(t, result.Error, test.want)
			assert.Contains(t, result.Content, test.want)
		})
	}
}

func TestSkillToolUsesOnlyAllowlistedEmbeddedPaths(t *testing.T) {
	result := newReferenceTool().Execute(context.Background(), inference.ToolCall{Arguments: `{"skill":"kubernetes","reference":"../bash.md"}`})
	assert.Error(t, result.Error)
	assert.Contains(t, result.Content, "reference")

	assert.Equal(t, json.RawMessage(`{"type":"object","properties":{"name":{"type":"string","description":"Skill name to load"}},"required":["name"],"additionalProperties":false}`), newSkillTool().Definition().Parameters)
}

func TestSkillToolDiscoversReferencesFromTheEmbeddedFilesystem(t *testing.T) {
	files := fstest.MapFS{
		"skills/bash/SKILL.md":             &fstest.MapFile{Data: []byte("---\nname: bash\ndescription: Load for shell work.\n---\n# Bash\n")},
		"skills/bash/references/custom.md": &fstest.MapFile{Data: []byte("# Custom reference\n")},
	}

	result := referenceTool{files: files}.Execute(context.Background(), inference.ToolCall{
		Arguments: `{"skill":"bash","reference":"custom.md"}`,
	})
	require.NoError(t, result.Error)
	assert.Equal(t, "# Custom reference\n", result.Content)
}

func TestSkillDiscoveryRequiresManifest(t *testing.T) {
	files := fstest.MapFS{
		"skills/incomplete/references/custom.md": &fstest.MapFile{Data: []byte("# Custom reference\n")},
	}

	_, err := discoverSkillPaths(files)
	assert.ErrorContains(t, err, `skill "incomplete" is missing SKILL.md`)
}

func TestSkillMetadataIsValidatedAndBodyIsSeparated(t *testing.T) {
	metadata, body, err := parseSkillDocument("---\nname: bash\ndescription: Load for shell work.\n---\n\n# Bash body\n", "bash")
	require.NoError(t, err)
	assert.Equal(t, skillMetadata{Name: "bash", Description: "Load for shell work."}, metadata)
	assert.Equal(t, "# Bash body", body)

	_, _, err = parseSkillDocument("---\nname: kubectl\ndescription: Load for shell work.\n---\nBody", "bash")
	assert.ErrorContains(t, err, "does not match directory")
}
