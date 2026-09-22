package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

func TestFenceYAMLWrapsRawYAML(t *testing.T) {
	input := "Restore the workload.\n\napiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: app\n"
	output := fenceYAML(input)
	assert.Contains(t, output, "```yaml")
	assert.Contains(t, output, "apiVersion: apps/v1")
}

func TestFenceYAMLLeavesProseAndExistingFences(t *testing.T) {
	prose := "The pod failed because the image could not be pulled.\nNothing else to add here.\n"
	assert.NotContains(t, fenceYAML(prose), "```yaml")

	fenced := "```yaml\napiVersion: v1\nkind: Pod\n```\n"
	assert.Equal(t, fenced, fenceYAML(fenced))
}

func TestToolCallLinesDecodeCommand(t *testing.T) {
	renderer := NewRenderer(80)
	call := inference.ToolCall{Name: "bash", Arguments: `{"command":"kubectl get pods -n legacy-app\nkubectl apply -f app.yaml"}`}
	lines := strings.Join(ToolCallLines(call, renderer), "\n")
	assert.Contains(t, lines, "kubectl get pods")
	assert.NotContains(t, lines, `\n`)
}

func TestRendererCachesByContent(t *testing.T) {
	renderer := NewRenderer(80)
	require.NotEmpty(t, renderer.Render("**bold** text"))
	require.Len(t, renderer.cache, 1)
}
