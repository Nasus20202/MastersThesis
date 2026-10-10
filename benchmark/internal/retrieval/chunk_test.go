package retrieval

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunkSectionsTracksHeadingsOutsideCodeFences(t *testing.T) {
	document := Document{Path: "docs/pod.md", Title: "Pods", Text: "Intro.\n\n# Pods\n\nTop.\n\n## Lifecycle\n\n```bash\n# not a heading\n```\n\n### Phases\n\nPhase text.\n\n## Empty\n\n## Probes\n\nProbe text."}

	chunks := ChunkDocument(document, Sections, testChunkParams)
	var headings, bodies []string
	for _, chunk := range chunks {
		headings = append(headings, chunk.Headings)
		bodies = append(bodies, chunk.Body)
	}
	assert.Equal(t, []string{"", "", "Lifecycle", "Lifecycle > Phases", "Probes"}, headings)
	assert.Equal(t, []string{"Intro.", "Top.", "```bash\n# not a heading\n```", "Phase text.", "Probe text."}, bodies)
	assert.Equal(t, "docs/pod.md\nPods > Lifecycle > Phases\nPhase text.", chunks[3].Rendered())
}

var testChunkParams = ChunkParams{MaxBytes: 1536, WindowOverlap: 256}

func TestChunkSectionsSplitsOversizedSectionsAtParagraphs(t *testing.T) {
	paragraph := strings.Repeat("word ", 150)
	document := Document{Path: "docs/long.md", Title: "Long", Text: "## Section\n\n" + strings.Repeat(paragraph+"\n\n", 6)}

	chunks := ChunkDocument(document, Sections, testChunkParams)
	require.Greater(t, len(chunks), 1)
	for _, chunk := range chunks {
		assert.LessOrEqual(t, len(chunk.Rendered()), testChunkParams.MaxBytes)
		assert.Equal(t, "Section", chunk.Headings)
		for _, word := range strings.Fields(chunk.Body) {
			assert.Equal(t, "word", word)
		}
	}
}

func TestChunkWindowsOverlapAndCoverDocument(t *testing.T) {
	var text strings.Builder
	text.WriteString("## First\n\n")
	for index := range 400 {
		text.WriteString("token")
		text.WriteString(strings.Repeat("x", index%7))
		text.WriteString(" ")
	}
	text.WriteString("\n\n## Second\n\nfinal-word")
	document := Document{Path: "docs/w.md", Title: "W", Text: text.String()}

	chunks := ChunkDocument(document, Windows, testChunkParams)
	require.Greater(t, len(chunks), 2)
	for index, chunk := range chunks {
		assert.LessOrEqual(t, len(chunk.Rendered()), testChunkParams.MaxBytes)
		if index > 0 {
			previous := chunks[index-1].Body
			assert.Contains(t, previous, strings.Fields(chunk.Body)[0], "window %d does not overlap the previous one", index)
		}
	}
	assert.Equal(t, "First", chunks[0].Headings)
	assert.True(t, strings.HasSuffix(chunks[len(chunks)-1].Body, "final-word"))
}

func TestSplitTextFallsBackToWordsAndRunes(t *testing.T) {
	pieces := splitText("alpha beta gamma delta ééééé", 11)
	assert.Equal(t, []string{"alpha beta", "gamma delta", "ééééé"}, pieces)
	assert.Equal(t, []string{"éé", "éé", "é"}, splitText("ééééé", 5))
}

func TestChunkWindowsKeepLongTokens(t *testing.T) {
	token := strings.Repeat("x", 3*testChunkParams.MaxBytes)
	document := Document{Path: "docs/t.md", Title: "T", Text: "intro " + token + " tail"}

	covered := 0
	for _, chunk := range ChunkDocument(document, Windows, testChunkParams) {
		covered += strings.Count(chunk.Body, "x")
	}
	assert.GreaterOrEqual(t, covered, len(token))
}
