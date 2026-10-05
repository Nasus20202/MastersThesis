package jsonfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteIndentsAndWriteLinesCompacts(t *testing.T) {
	dir := t.TempDir()
	values := []map[string]int{{"a": 1}, {"b": 2}}

	require.NoError(t, Write(filepath.Join(dir, "values.json"), values))
	require.NoError(t, WriteLines(filepath.Join(dir, "values.jsonl"), values))

	indented, err := os.ReadFile(filepath.Join(dir, "values.json"))
	require.NoError(t, err)
	assert.Equal(t, "[\n  {\n    \"a\": 1\n  },\n  {\n    \"b\": 2\n  }\n]\n", string(indented))
	lines, err := os.ReadFile(filepath.Join(dir, "values.jsonl"))
	require.NoError(t, err)
	assert.Equal(t, "{\"a\":1}\n{\"b\":2}\n", string(lines))
}
