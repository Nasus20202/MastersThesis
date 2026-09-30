package retrieval

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadCorpusReadsTitlesAndBlobSHAs(t *testing.T) {
	files := fstest.MapFS{
		"docs/a.md":      {Data: []byte("---\ntitle: Pod QoS\nweight: 10\n---\n\nBody\n")},
		"docs/sub/b.md":  {Data: []byte("hello\n")},
		"docs/notes.txt": {Data: []byte("ignored")},
		"other/c.md":     {Data: []byte("outside the subtree")},
	}

	documents, err := LoadCorpus(files, "docs")
	require.NoError(t, err)
	require.Len(t, documents, 2)
	assert.Equal(t, Document{Path: "docs/a.md", BlobSHA: gitBlobSHA(files["docs/a.md"].Data), Title: "Pod QoS", Text: "Body"}, documents[0])
	assert.Equal(t, "b", documents[1].Title)
	// git hash-object of "hello\n"
	assert.Equal(t, "ce013625030ba8dba906f756967f9e9ca394464a", documents[1].BlobSHA)
}

func TestCleanMarkdownKeepsVisibleText(t *testing.T) {
	text := `<!-- overview -->
## {{% heading "prerequisites" %}}

Pods use {{< glossary_tooltip text="volumes"
term_id="volume" >}} and a {{< glossary_tooltip term_id="node" >}}.

{{< note >}}
Keep this note.
{{< /note >}}

{{% code_sample file="pods/pod.yaml" %}}
See [Pod]({{< ref "../pod-v1#Pod" >}}) for v{{< skew currentVersion >}}.`

	assert.Equal(t, `## Before you begin

Pods use volumes and a node.

Keep this note.

(code sample: pods/pod.yaml)
See [Pod](../pod-v1#Pod) for v.`, cleanMarkdown(text))
}
