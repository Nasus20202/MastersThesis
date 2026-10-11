package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/command"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

// File tools let the model change a manifest through "export, edit, apply"
// instead of hand-written patches. They run in the sandbox through the same
// shell as bash, so the sandbox decides which paths are readable and writable;
// relative paths resolve against the sandbox working directory. Messages name
// only the file and the tool, never the scenario.
const (
	readFileToolName  = "read_file"
	writeFileToolName = "write_file"
	editFileToolName  = "edit_file"

	// fileContentEnv carries write_file content into the sandbox, which has no
	// stdin channel.
	fileContentEnv = "BENCHMARK_FILE_CONTENT"
	writeScript    = `mkdir -p -- "$(dirname -- "$1")" && printf '%s' "$` + fileContentEnv + `" > "$1"`

	truncatedFileMarker = "\n[file truncated; read the omitted part with bash, e.g. sed -n]\n"
)

// NewSandboxTools returns bash and, when config.FileTools is set, the file
// tools.
func NewSandboxTools(shell Shell, config Config) ([]Tool, error) {
	maxOutputBytes := config.MaxOutputBytes
	if maxOutputBytes == 0 {
		maxOutputBytes = DefaultConfig().MaxOutputBytes
	}
	bash, err := NewBashTool(shell, maxOutputBytes)
	if err != nil {
		return nil, err
	}
	tools := []Tool{bash}
	if config.FileTools {
		files := fileTools{shell: shell, maxOutputBytes: maxOutputBytes}
		tools = append(tools, readFileTool{files}, writeFileTool{files}, editFileTool{files})
	}
	return tools, nil
}

// FileEvidence records one file tool call. Bytes is the size read or written.
type FileEvidence struct {
	Path     string `json:"path"`
	Bytes    int    `json:"bytes"`
	ExitCode int    `json:"exit_code"`
	Stderr   string `json:"stderr,omitempty"`
}

type fileTools struct {
	shell          Shell
	maxOutputBytes int
}

func (f fileTools) read(ctx context.Context, path string) (string, FileEvidence, error) {
	result, err := f.shell.Exec(ctx, command.Spec{Program: "cat", Args: []string{"--", path}})
	evidence := FileEvidence{Path: path, Bytes: len(result.Stdout), ExitCode: result.ExitCode, Stderr: result.Stderr}
	if err != nil {
		return "", evidence, fileError("read", path, result, err)
	}
	return result.Stdout, evidence, nil
}

func (f fileTools) write(ctx context.Context, path, content string) (FileEvidence, error) {
	result, err := f.shell.Exec(ctx, command.Spec{
		Program: "bash",
		Args:    []string{"-c", writeScript, writeFileToolName, path},
		Env:     map[string]string{fileContentEnv: content},
	})
	evidence := FileEvidence{Path: path, Bytes: len(content), ExitCode: result.ExitCode, Stderr: result.Stderr}
	if err != nil {
		return evidence, fileError("write", path, result, err)
	}
	return evidence, nil
}

// fileError prefers the command's own message, such as "No such file or
// directory", over the sandbox's wrapper error.
func fileError(action, path string, result command.Result, err error) error {
	if message := strings.TrimSpace(result.Stderr); message != "" {
		return fmt.Errorf("could not %s %s: %s", action, path, message)
	}
	return fmt.Errorf("could not %s %s: %w", action, path, err)
}

func decodeFileArguments(call inference.ToolCall, target any) (string, error) {
	if err := json.Unmarshal([]byte(call.Arguments), target); err != nil {
		return "", fmt.Errorf("malformed %s arguments: %w", call.Name, err)
	}
	var path struct {
		Path string `json:"path"`
	}
	_ = json.Unmarshal([]byte(call.Arguments), &path)
	if strings.TrimSpace(path.Path) == "" {
		return "", fmt.Errorf("%s path is required", call.Name)
	}
	return path.Path, nil
}

type readFileTool struct{ fileTools }

func (readFileTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        readFileToolName,
		Description: "Read a text file inside the sandbox.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"File path; relative paths start in the working directory"}},"required":["path"],"additionalProperties":false}`),
	}
}

func (t readFileTool) Execute(ctx context.Context, call inference.ToolCall) ToolResult {
	var arguments struct{}
	path, err := decodeFileArguments(call, &arguments)
	if err != nil {
		return Failed(err, nil)
	}
	content, evidence, err := t.read(ctx, path)
	if err != nil {
		return Failed(err, evidence)
	}
	if content == "" {
		content = "(empty file)"
	}
	return ToolResult{Content: limitOutput(content, truncatedFileMarker, t.maxOutputBytes), Details: evidence}
}

type writeFileTool struct{ fileTools }

func (writeFileTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        writeFileToolName,
		Description: "Create or overwrite a text file inside the sandbox with the given content.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"File path; relative paths start in the working directory"},"content":{"type":"string","description":"Complete new file content"}},"required":["path","content"],"additionalProperties":false}`),
	}
}

func (t writeFileTool) Execute(ctx context.Context, call inference.ToolCall) ToolResult {
	var arguments struct {
		Content *string `json:"content"`
	}
	path, err := decodeFileArguments(call, &arguments)
	if err != nil {
		return Failed(err, nil)
	}
	if arguments.Content == nil {
		return Failed(errors.New("write_file content is required"), nil)
	}
	evidence, err := t.write(ctx, path, *arguments.Content)
	if err != nil {
		return Failed(err, evidence)
	}
	return ToolResult{Content: fmt.Sprintf("wrote %d bytes to %s", evidence.Bytes, path), Details: evidence}
}

type editFileTool struct{ fileTools }

func (editFileTool) Definition() inference.Tool {
	return inference.Tool{
		Name:        editFileToolName,
		Description: "Replace one exact occurrence of old_string with new_string in a text file inside the sandbox. old_string must match the file exactly, including indentation, and occur exactly once.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"path":{"type":"string","description":"File path; relative paths start in the working directory"},"old_string":{"type":"string","description":"Exact text to replace"},"new_string":{"type":"string","description":"Replacement text"}},"required":["path","old_string","new_string"],"additionalProperties":false}`),
	}
}

func (t editFileTool) Execute(ctx context.Context, call inference.ToolCall) ToolResult {
	var arguments struct {
		OldString string  `json:"old_string"`
		NewString *string `json:"new_string"`
	}
	path, err := decodeFileArguments(call, &arguments)
	if err != nil {
		return Failed(err, nil)
	}
	if arguments.OldString == "" || arguments.NewString == nil {
		return Failed(errors.New("edit_file old_string and new_string are required"), nil)
	}
	content, evidence, err := t.read(ctx, path)
	if err != nil {
		return Failed(err, evidence)
	}
	switch count := strings.Count(content, arguments.OldString); count {
	case 0:
		return Failed(fmt.Errorf("old_string was not found in %s; read the file and copy the text exactly", path), evidence)
	case 1:
	default:
		return Failed(fmt.Errorf("old_string occurs %d times in %s; include more surrounding text so it matches once", count, path), evidence)
	}
	evidence, err = t.write(ctx, path, strings.Replace(content, arguments.OldString, *arguments.NewString, 1))
	if err != nil {
		return Failed(err, evidence)
	}
	return ToolResult{Content: "replaced 1 occurrence in " + path, Details: evidence}
}
