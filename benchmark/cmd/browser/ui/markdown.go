// Markdown renders agent transcripts to terminal lines, auto-fencing
// raw YAML so glamour can highlight it and caching by width and source
package ui

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"charm.land/glamour/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const margin = 2

var (
	yamlKeyPattern  = regexp.MustCompile(`^[A-Za-z_][\w.-]*(/[A-Za-z_][\w.-]*)?\s*:`)
	yamlListPattern = regexp.MustCompile(`^-\s+\S`)
)

// Renderer wraps glamour with a width-specific cache so transcript lines are
// only rendered once per width.
type Renderer struct {
	width    int
	renderer *glamour.TermRenderer
	cache    map[string][]string
}

// New returns a renderer for the given content width.
func NewRenderer(width int) *Renderer {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(max(20, width)),
		glamour.WithPreservedNewLines(),
	)
	if err != nil {
		renderer = nil
	}
	return &Renderer{width: width, renderer: renderer, cache: make(map[string][]string)}
}

// Width reports the content width the renderer was built for.
func (r *Renderer) Width() int { return r.width }

// Render converts markdown (with embedded YAML fenced automatically) to
// indented terminal lines. Plain text falls back to a simple split.
func (r *Renderer) Render(source string) []string {
	source = strings.TrimRight(source, "\n")
	if source == "" {
		return nil
	}
	if cached, ok := r.cache[source]; ok {
		return cached
	}
	lines := r.renderUncached(source)
	r.cache[source] = lines
	return lines
}

func (r *Renderer) renderUncached(source string) []string {
	if r.renderer == nil {
		return strings.Split(source, "\n")
	}
	out, err := r.renderer.Render(fenceYAML(source))
	if err != nil {
		return strings.Split(source, "\n")
	}
	return clean(out)
}

func clean(out string) []string {
	out = strings.Trim(out, "\n")
	raw := strings.Split(out, "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimRight(line, " ")
		if len(line) >= margin && line[:margin] == strings.Repeat(" ", margin) {
			line = line[margin:]
		}
		lines = append(lines, line)
	}
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// ToolCallLines renders a tool call's decoded arguments. String arguments are
// shown as fenced blocks so commands and embedded YAML stay readable instead of
// a single escaped JSON line. The tool name is carried by the card title.
func ToolCallLines(call inference.ToolCall, renderer *Renderer) []string {
	var lines []string
	var arguments map[string]any
	if err := json.Unmarshal([]byte(call.Arguments), &arguments); err != nil || len(arguments) == 0 {
		return renderer.Render("```\n" + call.Arguments + "\n```")
	}
	keys := make([]string, 0, len(arguments))
	for key := range arguments {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value, ok := arguments[key].(string)
		if !ok {
			encoded, _ := json.Marshal(arguments[key])
			lines = append(lines, MutedStyle.Render(key+": "+string(encoded)))
			continue
		}
		language := ""
		if key == "command" {
			language = "bash"
		}
		lines = append(lines, MutedStyle.Render(key+":"))
		lines = append(lines, renderer.Render("```"+language+"\n"+value+"\n```")...)
	}
	return lines
}

// fenceYAML wraps contiguous raw YAML runs in ```yaml fences so glamour can
// syntax-highlight them. Fenced code is left untouched.
func fenceYAML(source string) string {
	lines := strings.Split(source, "\n")
	var out []string
	inFence := false
	for index := 0; index < len(lines); {
		trimmed := strings.TrimSpace(lines[index])
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			out = append(out, lines[index])
			index++
			continue
		}
		if !inFence {
			if end := yamlRunEnd(lines, index); end > index {
				out = append(out, "```yaml")
				out = append(out, lines[index:end]...)
				out = append(out, "```")
				index = end
				continue
			}
		}
		out = append(out, lines[index])
		index++
	}
	return strings.Join(out, "\n")
}

// yamlRunEnd returns the end index of a YAML-looking run starting at start, or
// start when the run is not confident enough.
func yamlRunEnd(lines []string, start int) int {
	first := strings.TrimSpace(lines[start])
	if !yamlKeyPattern.MatchString(first) && !yamlListPattern.MatchString(first) {
		return start
	}
	end := start
	scored, total := 0, 0
	for end < len(lines) {
		trimmed := strings.TrimSpace(lines[end])
		if trimmed == "" {
			break
		}
		total++
		if yamlKeyPattern.MatchString(trimmed) || yamlListPattern.MatchString(trimmed) || strings.HasPrefix(trimmed, "#") {
			scored++
		}
		end++
	}
	if total < 2 || scored*2 < total {
		return start
	}
	return end
}
