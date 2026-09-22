package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

func TestBuildFlattensMessagesIntoSelectableCards(t *testing.T) {
	messages := []inference.Message{
		{Role: "user", Content: "Restore the deployment."},
		{Role: "assistant", Content: "Checking the workload."},
	}
	layout := Build(messages, 1, nil, ui.NewRenderer(60), 60)
	require.NotEmpty(t, layout.Lines)
	require.Len(t, layout.MessageSpans, 2)
	require.NotEmpty(t, layout.CardSpans)
	assert.Contains(t, layout.Lines[0], "user")
}

func TestBuildCollapsesLongMessagesUntilExpanded(t *testing.T) {
	long := ""
	for range 40 {
		long += "detail line\n"
	}
	messages := []inference.Message{{Role: "assistant", Content: long}}

	collapsed := Build(messages, -1, nil, ui.NewRenderer(60), 60)
	expanded := Build(messages, -1, map[int]bool{0: true}, ui.NewRenderer(60), 60)
	assert.Less(t, len(collapsed.Lines), len(expanded.Lines))
	assert.Contains(t, strings.Join(collapsed.Lines, "\n"), "expand")
}

func TestRenderCardProducesBorderedBox(t *testing.T) {
	item := card{title: "bash", body: []string{"kubectl get pods"}, border: ui.Green, titleStyle: ui.Warning}
	lines := renderCard(item, 40, false)
	require.GreaterOrEqual(t, len(lines), 3)
	assert.Contains(t, lines[0], "bash")
	assert.Contains(t, lines[len(lines)-1], "╰")
}
