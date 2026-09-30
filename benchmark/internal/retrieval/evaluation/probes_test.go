package evaluation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseProbesCollectsPrimaryAndCitedFiles(t *testing.T) {
	data := []byte(`{"tasks": [
		{"id": "K01", "prompt": " Why Pending? ", "source": "content/en/docs/a.md#sha",
		 "criteria": {"diagnosis": "x [source: content/en/docs/b.md@sha]", "repair": "content/en/docs/a.md#sha"},
		 "provided_context": [{"source": {"path": "content/en/docs/c.md"}}]},
		{"id": "K02", "prompt": "Why denied?", "sources": [{"path": "content/en/docs/d.md"}, {"path": "content/en/docs/e.md"}]}
	]}`)

	probes, err := ParseProbes(data)
	require.NoError(t, err)
	assert.Equal(t, []Probe{
		{ID: "K01", Prompt: "Why Pending?", Primary: []string{"content/en/docs/a.md"}, Relevant: []string{"content/en/docs/a.md", "content/en/docs/b.md", "content/en/docs/c.md"}},
		{ID: "K02", Prompt: "Why denied?", Primary: []string{"content/en/docs/d.md", "content/en/docs/e.md"}, Relevant: []string{"content/en/docs/d.md", "content/en/docs/e.md"}},
	}, probes)

	_, err = ParseProbes([]byte(`{"tasks": [{"id": "K03", "prompt": "No source"}]}`))
	assert.ErrorContains(t, err, `probe "K03"`)
}
