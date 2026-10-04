package common

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// dirShell runs commands locally in dir, standing in for the sandbox working
// directory.
type dirShell struct{ dir string }

func (s dirShell) Exec(ctx context.Context, spec command.Spec) (command.Result, error) {
	spec.Dir = s.dir
	return command.LocalExecutor{Environment: os.Environ()}.Run(ctx, spec)
}

func fileTool(t *testing.T, dir, name string) Tool {
	t.Helper()
	tools, err := NewSandboxTools(dirShell{dir: dir}, Config{FileTools: true})
	require.NoError(t, err)
	for _, tool := range tools {
		if tool.Definition().Name == name {
			return tool
		}
	}
	t.Fatalf("tool %s not found", name)
	return nil
}

func callFile(t *testing.T, tool Tool, arguments string) ToolResult {
	t.Helper()
	return tool.Execute(context.Background(), inference.ToolCall{Type: "function", Name: tool.Definition().Name, Arguments: arguments})
}

func TestNewSandboxToolsAddsFileToolsOnlyWhenEnabled(t *testing.T) {
	names := func(config Config) []string {
		tools, err := NewSandboxTools(dirShell{}, config)
		require.NoError(t, err)
		var result []string
		for _, tool := range tools {
			result = append(result, tool.Definition().Name)
		}
		return result
	}
	assert.Equal(t, []string{"bash"}, names(Config{}))
	assert.Equal(t, []string{"bash", "read_file", "write_file", "edit_file"}, names(Config{FileTools: true}))
	_, err := NewSandboxTools(nil, Config{})
	assert.Error(t, err)
}

func TestWriteReadAndEditFile(t *testing.T) {
	dir := t.TempDir()
	manifest := "spec:\n  containers:\n  - name: app\n    image: nginx:1.0\n    args: [\"$HOME\", 'quoted']\n"

	written := callFile(t, fileTool(t, dir, "write_file"), `{"path":"manifests/app.yaml","content":`+jsonString(manifest)+`}`)
	require.NoError(t, written.Error)
	assert.Equal(t, "wrote 85 bytes to manifests/app.yaml", written.Content)
	data, err := os.ReadFile(filepath.Join(dir, "manifests", "app.yaml"))
	require.NoError(t, err)
	assert.Equal(t, manifest, string(data))

	read := callFile(t, fileTool(t, dir, "read_file"), `{"path":"manifests/app.yaml"}`)
	require.NoError(t, read.Error)
	assert.Equal(t, manifest, read.Content)
	assert.Equal(t, FileEvidence{Path: "manifests/app.yaml", Bytes: len(manifest)}, read.Details)

	edited := callFile(t, fileTool(t, dir, "edit_file"), `{"path":"manifests/app.yaml","old_string":"image: nginx:1.0","new_string":"image: nginx:1.1"}`)
	require.NoError(t, edited.Error)
	assert.Equal(t, "replaced 1 occurrence in manifests/app.yaml", edited.Content)
	data, err = os.ReadFile(filepath.Join(dir, "manifests", "app.yaml"))
	require.NoError(t, err)
	assert.Equal(t, strings.Replace(manifest, "nginx:1.0", "nginx:1.1", 1), string(data))
}

func TestEditFileRequiresExactlyOneMatch(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "app.yaml"), []byte("a: 1\nb: 1\n"), 0o600))
	edit := fileTool(t, dir, "edit_file")

	missing := callFile(t, edit, `{"path":"app.yaml","old_string":"c: 1","new_string":"c: 2"}`)
	require.Error(t, missing.Error)
	assert.Equal(t, "old_string was not found in app.yaml; read the file and copy the text exactly", missing.Content)

	repeated := callFile(t, edit, `{"path":"app.yaml","old_string":": 1","new_string":": 2"}`)
	require.Error(t, repeated.Error)
	assert.Equal(t, "old_string occurs 2 times in app.yaml; include more surrounding text so it matches once", repeated.Content)

	data, err := os.ReadFile(filepath.Join(dir, "app.yaml"))
	require.NoError(t, err)
	assert.Equal(t, "a: 1\nb: 1\n", string(data))
}

func TestFileToolsReportFailures(t *testing.T) {
	dir := t.TempDir()

	missing := callFile(t, fileTool(t, dir, "read_file"), `{"path":"absent.yaml"}`)
	require.Error(t, missing.Error)
	assert.Contains(t, missing.Content, "could not read absent.yaml: ")
	assert.Contains(t, missing.Content, "No such file or directory")
	assert.Equal(t, 1, missing.Details.(FileEvidence).ExitCode)

	for name, arguments := range map[string]string{
		"read_file":  `{"path":" "}`,
		"write_file": `{"path":"app.yaml"}`,
		"edit_file":  `{"path":"app.yaml","old_string":"","new_string":"x"}`,
	} {
		result := callFile(t, fileTool(t, dir, name), arguments)
		assert.Error(t, result.Error, name)
	}
	malformed := callFile(t, fileTool(t, dir, "write_file"), `{`)
	assert.Contains(t, malformed.Content, "malformed write_file arguments")
}

func TestReadFileLimitsModelVisibleContent(t *testing.T) {
	dir := t.TempDir()
	content := strings.Repeat("x", maxModelVisibleCommandOutputBytes) + "end"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "big.txt"), []byte(content), 0o600))

	result := callFile(t, fileTool(t, dir, "read_file"), `{"path":"big.txt"}`)
	require.NoError(t, result.Error)
	assert.Len(t, result.Content, maxModelVisibleCommandOutputBytes)
	assert.Contains(t, result.Content, truncatedFileMarker)
	assert.True(t, strings.HasSuffix(result.Content, "end"))
	assert.Equal(t, len(content), result.Details.(FileEvidence).Bytes)
}

func jsonString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
